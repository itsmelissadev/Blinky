package postgresql

import (
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"blinky/internal"
	"blinky/internal/pkg/pathutil"
)

type Manager struct {
	cfg *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) getBinPath(binName string) string {
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return filepath.Join(pathutil.GetPostgresPath(), "bin", binName)
}

func (m *Manager) Start() error {
	logger.Info("[POSTGRES/PROCESS] Initializing managed PostgreSQL %s...", internal.AppPostgresVersion)

	if err := m.EnsureBinaries(); err != nil {
		return err
	}

	logger.Info("[POSTGRES/PROCESS] Checking PostgreSQL state...")

	if err := m.ensureInitialized(); err != nil {
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}

	if m.IsRunning() {
		logger.Success("[POSTGRES/PROCESS] PostgreSQL is already running")
		return nil
	}

	logger.Info("[POSTGRES/PROCESS] Starting PostgreSQL server directly...")

	postgresBin := m.getBinPath("postgres")
	args := []string{"-D", pathutil.GetPostgresDataPath(), "-p", m.cfg.PostgresDBPort}

	cmd := exec.Command(postgresBin, args...)

	pipe, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start postgres binary: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			logger.Info("[POSTGRES/LOG] %s", scanner.Text())
		}
	}()

	for range 30 {
		if m.IsRunning() {
			logger.Success("[POSTGRES/PROCESS] PostgreSQL is UP and running on port %s", m.cfg.PostgresDBPort)
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("postgresql failed to start within timeout")
}

func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return nil
	}

	logger.Warn("[POSTGRES/PROCESS] Stopping PostgreSQL server...")

	pgCtl := m.getBinPath("pg_ctl")
	args := []string{"stop", "-D", pathutil.GetPostgresDataPath(), "-m", "fast"}

	cmd := exec.Command(pgCtl, args...)
	return cmd.Run()
}

func (m *Manager) IsRunning() bool {
	pgCtl := m.getBinPath("pg_ctl")
	args := []string{"status", "-D", pathutil.GetPostgresDataPath()}

	cmd := exec.Command(pgCtl, args...)
	err := cmd.Run()
	return err == nil
}

func (m *Manager) ensureInitialized() error {
	controlFile := filepath.Join(pathutil.GetPostgresDataPath(), "global", "pg_control")
	if _, err := os.Stat(controlFile); err == nil {
		return nil
	}

	logger.Warn("[POSTGRES/PROCESS] Data directory is empty. Initializing...")

	if err := os.MkdirAll(pathutil.GetPostgresDataPath(), 0700); err != nil {
		return err
	}

	initDb := m.getBinPath("initdb")
	args := []string{"-D", pathutil.GetPostgresDataPath(), "-U", m.cfg.PostgresDBUser, "--encoding=UTF8", "--locale=C"}

	if m.cfg.PostgresDBPassword != "" {
		pwFile := filepath.Join(os.TempDir(), "pg_pw.txt")
		if err := os.WriteFile(pwFile, []byte(m.cfg.PostgresDBPassword), 0600); err == nil {
			args = append(args, "--pwfile", pwFile)
			defer os.Remove(pwFile)
		}
	}

	cmd := exec.Command(initDb, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("initdb failed: %v\nOutput: %s", err, string(output))
	}

	if err := m.hardenSecurity(); err != nil {
		logger.Warn("[POSTGRES/SECURITY] Failed to harden pg_hba.conf: %v", err)
	}

	return nil
}

func (m *Manager) hardenSecurity() error {
	hbaFile := filepath.Join(pathutil.GetPostgresDataPath(), "pg_hba.conf")
	content, err := os.ReadFile(hbaFile)
	if err != nil {
		return err
	}

	newContent := strings.ReplaceAll(string(content), "trust", "scram-sha-256")

	return os.WriteFile(hbaFile, []byte(newContent), 0600)
}
