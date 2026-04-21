package ssh

import (
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

func StartSSHServer(cfg *config.Config) {
	if !cfg.AdminSSHEnabled && !cfg.PublicSSHEnabled {
		return
	}

	sshConfig := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == cfg.SSHUser && string(pass) == cfg.SSHPassword {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateKey, err := getOrGenerateHostKey()
	if err != nil {
		logger.Error("[SSH] Failed to handle host key: %v", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		logger.Error("[SSH] Failed to parse private key: %v", err)
		return
	}

	sshConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", "0.0.0.0:"+cfg.SSHPort)
	if err != nil {
		logger.Error("[SSH] Failed to listen on port %s: %v", cfg.SSHPort, err)
		return
	}

	logger.Success("[SSH] Tunnel Gateway listening on 0.0.0.0:%s", cfg.SSHPort)

	for {
		nConn, err := listener.Accept()
		if err != nil {
			logger.Error("[SSH] Failed to accept incoming connection: %v", err)
			continue
		}

		go handleSSHConn(nConn, sshConfig, cfg)
	}
}

func getOrGenerateHostKey() ([]byte, error) {
	keyPath := "system/.ssh/host_key.pem"
	if err := os.MkdirAll("system/.ssh", 0700); err != nil {
		return nil, err
	}

	if _, err := os.Stat(keyPath); err == nil {
		return os.ReadFile(keyPath)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	keyContent := pem.EncodeToMemory(pemBlock)
	if err := os.WriteFile(keyPath, keyContent, 0600); err != nil {
		return nil, err
	}

	return keyContent, nil
}

func handleSSHConn(nConn net.Conn, sshConfig *ssh.ServerConfig, cfg *config.Config) {
	_, chans, reqs, err := ssh.NewServerConn(nConn, sshConfig)
	if err != nil {
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		switch newChannel.ChannelType() {
		case "session":
			go handleSessionChannel(newChannel)
		case "direct-tcpip":
			go handleDirectTCPIP(newChannel, cfg)
		default:
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		}
	}
}

func handleSessionChannel(newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		return
	}

	go func() {
		defer channel.Close()
		for req := range requests {
			switch req.Type {
			case "shell":
				req.Reply(true, nil)
				fmt.Fprintf(channel, "\r\n\x1b[34m--- Blinky Tunnel Gateway ---\x1b[0m\r\n")
				fmt.Fprintf(channel, "\x1b[33mTerminal access is disabled for security.\x1b[0m\r\n")
				fmt.Fprintf(channel, "The tunnel is now \x1b[32mACTIVE\x1b[0m. You can minimize this screen.\r\n\r\n")
			case "pty-req":
				req.Reply(true, nil)
			default:
				req.Reply(false, nil)
			}
		}
	}()
}

func handleDirectTCPIP(newChannel ssh.NewChannel, cfg *config.Config) {
	var payload struct {
		Addr       string
		Port       uint32
		OriginAddr string
		OriginPort uint32
	}
	if err := ssh.Unmarshal(newChannel.ExtraData(), &payload); err != nil {
		return
	}

	isAllowed := false
	if payload.Addr == "127.0.0.1" || payload.Addr == "localhost" {
		if fmt.Sprintf("%d", payload.Port) == cfg.AdminPanelPort || fmt.Sprintf("%d", payload.Port) == cfg.PublicAPIPort {
			isAllowed = true
		}
	}

	if !isAllowed {
		newChannel.Reject(ssh.Prohibited, "forwarding to this address is prohibited")
		return
	}

	destConn, err := net.Dial("tcp", net.JoinHostPort(payload.Addr, fmt.Sprintf("%d", payload.Port)))
	if err != nil {
		newChannel.Reject(ssh.ConnectionFailed, "could not dial destination")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		destConn.Close()
		return
	}

	go ssh.DiscardRequests(requests)

	go func() {
		defer channel.Close()
		defer destConn.Close()
		io.Copy(channel, destConn)
	}()

	go func() {
		defer channel.Close()
		defer destConn.Close()
		io.Copy(destConn, channel)
	}()
}
