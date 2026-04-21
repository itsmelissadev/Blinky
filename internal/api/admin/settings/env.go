package settings

import (
	"blinky/internal/api"
	"blinky/internal/pkg/logger"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (h *backupHandler) getEnv(c *fiber.Ctx) error {
	content, err := os.ReadFile(".env")
	if err != nil {
		logger.Error("[SETTINGS/ENV] Failed to read .env: %v", err)
		return api.SendError(c, api.ErrEnvReadFailed)
	}

	lines := strings.Split(string(content), "\n")
	var vars []EnvVar
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			vars = append(vars, EnvVar{
				Key:   parts[0],
				Value: parts[1],
			})
		}
	}

	return api.Success(c, vars)
}

func (h *backupHandler) updateEnv(c *fiber.Ctx) error {
	var body struct {
		OldKey string `json:"oldKey"`
		Key    string `json:"key"`
		Value  string `json:"value"`
	}
	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	if body.Key == "" {
		return api.SendError(c, api.ErrEnvMissingKey)
	}

	updates := map[string]string{body.Key: body.Value}
	if body.OldKey != "" && body.OldKey != body.Key {
		updates[body.OldKey] = ""
	}

	if err := updateEnvVariables(updates); err != nil {
		logger.Error("[SETTINGS/ENV] Failed to update env variable %s: %v", body.Key, err)
		return api.SendError(c, api.ErrEnvUpdateFailed)
	}

	logger.Info("[SETTINGS/ENV] Environment variable updated: %s (was %s)", body.Key, body.OldKey)
	return api.Success(c, api.SuccessEnvUpdated)
}

func (h *backupHandler) deleteEnv(c *fiber.Ctx) error {
	var body struct {
		Keys []string `json:"keys"`
	}
	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	if len(body.Keys) == 0 {
		return api.Success(c, api.SuccessEnvNoChanges)
	}

	if err := deleteEnvVariables(body.Keys); err != nil {
		logger.Error("[SETTINGS/ENV] Failed to delete env variables: %v", err)
		return api.SendError(c, api.ErrEnvDeleteFailed)
	}

	logger.Info("[SETTINGS/ENV] Deleted %d environment variables", len(body.Keys))
	return api.Success(c, api.SuccessEnvDeleted)
}

func deleteEnvVariables(keys []string) error {
	content, err := os.ReadFile(".env")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) > 0 && keyMap[parts[0]] {
			continue
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(".env", []byte(strings.Join(newLines, "\n")), 0644)
}

func updateEnvVariable(oldKey string, newKey string, value string) error {
	return updateEnvVariables(map[string]string{oldKey: value})
}

func updateEnvVariables(updates map[string]string) error {
	content, err := os.ReadFile(".env")
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	for oldKey, value := range updates {
		found := false
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, oldKey+"=") {
				lines[i] = fmt.Sprintf("%s=%s", oldKey, value)
				found = true
				break
			}
		}
		if !found {
			lines = append(lines, fmt.Sprintf("%s=%s", oldKey, value))
		}
	}

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(".env", []byte(newContent), 0644)
}
