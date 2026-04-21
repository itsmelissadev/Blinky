package settings

import (
	"blinky/internal/api"
	"blinky/internal/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

func (h *backupHandler) getServerConfig(c *fiber.Ctx) error {
	return api.Success(c, fiber.Map{
		"publicApiHost":  h.cfg.PublicAPIHost,
		"publicApiPort":  h.cfg.PublicAPIPort,
		"adminPanelHost": h.cfg.AdminPanelHost,
		"adminPanelPort": h.cfg.AdminPanelPort,
	})
}

func (h *backupHandler) updateServerConfig(c *fiber.Ctx) error {
	var body struct {
		PublicAPIHost  string `json:"publicApiHost"`
		PublicAPIPort  string `json:"publicApiPort"`
		AdminPanelHost string `json:"adminPanelHost"`
		AdminPanelPort string `json:"adminPanelPort"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	v := api.NewValidator()
	v.Required("publicApiHost", body.PublicAPIHost)
	v.Required("publicApiPort", body.PublicAPIPort)
	v.Required("adminPanelHost", body.AdminPanelHost)
	v.Required("adminPanelPort", body.AdminPanelPort)

	if v.HasErrors() {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	updates := map[string]string{
		"PUBLIC_API_HOST":  body.PublicAPIHost,
		"PUBLIC_API_PORT":  body.PublicAPIPort,
		"ADMIN_PANEL_HOST": body.AdminPanelHost,
		"ADMIN_PANEL_PORT": body.AdminPanelPort,
	}

	if err := updateEnvVariables(updates); err != nil {
		logger.Error("[SETTINGS/SERVER] Failed to update env variables: %v", err)
		return api.SendError(c, api.ErrEnvUpdateFailed)
	}

	h.cfg.PublicAPIHost = body.PublicAPIHost
	h.cfg.PublicAPIPort = body.PublicAPIPort
	h.cfg.AdminPanelHost = body.AdminPanelHost
	h.cfg.AdminPanelPort = body.AdminPanelPort

	logger.Success("[SETTINGS/SERVER] Server configuration updated")
	return api.Success(c, api.SuccessConfigUpdated)
}
