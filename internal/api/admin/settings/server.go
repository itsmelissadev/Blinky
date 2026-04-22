package settings

import (
	"blinky/internal/api"
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
)

func (h *backupHandler) getServerConfig(c *fiber.Ctx) error {
	logger.Info("[ADMIN/SETTINGS] Fetching server configuration...")

	res := fiber.Map{
		"publicApiHost":    h.cfg.PublicAPIHost,
		"publicApiPort":    h.cfg.PublicAPIPort,
		"adminPanelHost":   h.cfg.AdminPanelHost,
		"adminPanelPort":   h.cfg.AdminPanelPort,
		"adminSshEnabled":  h.cfg.AdminSSHEnabled,
		"publicSshEnabled": h.cfg.PublicSSHEnabled,
		"sshPort":          h.cfg.SSHPort,
		"sshUser":          h.cfg.SSHUser,
		"sshPassword":      h.cfg.SSHPassword,
	}

	logger.Success("[ADMIN/SETTINGS] Server configuration retrieved successfully")
	return api.Success(c, res)
}

func (h *backupHandler) updateServerConfig(c *fiber.Ctx) error {
	logger.Info("[ADMIN/SETTINGS] Initiating server configuration update...")

	var body struct {
		PublicAPIHost    string `json:"publicApiHost"`
		PublicAPIPort    string `json:"publicApiPort"`
		AdminPanelHost   string `json:"adminPanelHost"`
		AdminPanelPort   string `json:"adminPanelPort"`
		AdminSSHEnabled  bool   `json:"adminSshEnabled"`
		PublicSSHEnabled bool   `json:"publicSshEnabled"`
		SSHPort          string `json:"sshPort"`
		SSHUser          string `json:"sshUser"`
		SSHPassword      string `json:"sshPassword"`
	}

	if err := c.BodyParser(&body); err != nil {
		logger.Error("[ADMIN/SETTINGS] Failed to parse request body: %v", err)
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	logger.Info("[ADMIN/SETTINGS] Updating environment variables...")
	updates := map[string]string{
		"PUBLIC_API_HOST":    body.PublicAPIHost,
		"PUBLIC_API_PORT":    body.PublicAPIPort,
		"ADMIN_PANEL_HOST":   body.AdminPanelHost,
		"ADMIN_PANEL_PORT":   body.AdminPanelPort,
		"ADMIN_SSH_ENABLED":  fmt.Sprintf("%v", body.AdminSSHEnabled),
		"PUBLIC_SSH_ENABLED": fmt.Sprintf("%v", body.PublicSSHEnabled),
		"SSH_PORT":           body.SSHPort,
		"SSH_USER":           body.SSHUser,
		"SSH_PASS":           body.SSHPassword,
	}

	if err := config.UpdateEnvVariables(updates); err != nil {
		logger.Error("[ADMIN/SETTINGS] Environment update failed: %v", err)
		return api.SendError(c, api.ErrEnvUpdateFailed, 500)
	}

	logger.Info("[ADMIN/SETTINGS] Synchronizing in-memory configuration...")
	h.cfg.PublicAPIHost = body.PublicAPIHost
	h.cfg.PublicAPIPort = body.PublicAPIPort
	h.cfg.AdminPanelHost = body.AdminPanelHost
	h.cfg.AdminPanelPort = body.AdminPanelPort
	h.cfg.AdminSSHEnabled = body.AdminSSHEnabled
	h.cfg.PublicSSHEnabled = body.PublicSSHEnabled
	h.cfg.SSHPort = body.SSHPort
	h.cfg.SSHUser = body.SSHUser
	h.cfg.SSHPassword = body.SSHPassword

	logger.Success("[ADMIN/SETTINGS] Server configuration updated and saved")
	return api.Success(c, api.SuccessConfigUpdated)
}

func (h *backupHandler) testSSHConfig(c *fiber.Ctx) error {
	logger.Info("[ADMIN/SETTINGS/SSH] Testing SSH port availability...")

	var body struct {
		SSHPort string `json:"sshPort"`
	}

	if err := c.BodyParser(&body); err != nil {
		logger.Error("[ADMIN/SETTINGS/SSH] Failed to parse SSH test body")
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	logger.Info("[ADMIN/SETTINGS/SSH] Attempting to listen on port %s", body.SSHPort)
	ln, err := net.Listen("tcp", ":"+body.SSHPort)
	if err != nil {
		logger.Error("[ADMIN/SETTINGS/SSH] Port %s is unavailable: %v", body.SSHPort, err)
		return api.SendError(c, api.ErrServerPortInUse, 409)
	}
	ln.Close()

	logger.Success("[ADMIN/SETTINGS/SSH] SSH port %s is available for use", body.SSHPort)
	return api.Success(c, api.SuccessServerPortOK)
}
