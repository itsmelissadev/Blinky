package settings

import (
	"blinky/internal/api"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/postgresql"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (h *backupHandler) getPostgresConfig(c *fiber.Ctx) error {
	return api.Success(c, fiber.Map{
		"host":             h.cfg.DBHost,
		"port":             h.cfg.DBPort,
		"user":             h.cfg.DBUser,
		"password":         h.cfg.DBPass,
		"database":         h.cfg.DBName,
		"postgresPath":     h.cfg.PostgresPath,
		"postgresDataPath": h.cfg.PostgresDataPath,
	})
}

func (h *backupHandler) testPostgresConnection(c *fiber.Ctx) error {
	var body struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	val := api.NewValidator()
	val.Required("host", body.Host)
	val.Required("port", body.Port)
	val.Required("user", body.User)
	val.Required("database", body.Database)
	if val.HasErrors() {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		body.User, body.Password, body.Host, body.Port, body.Database)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return api.SendError(c, api.ErrCoreInternalServer.WithReplacements(map[string]string{
			"error": err.Error(),
		}))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return api.SendError(c, api.ErrCoreInternalServer.WithReplacements(map[string]string{
			"error": fmt.Sprintf("Ping failed: %v", err),
		}))
	}

	logger.Success("[SETTINGS/POSTGRES] Connection test successful for %s:%s", body.Host, body.Port)
	return api.Success(c, api.SuccessPostgresTestOK)
}

func (h *backupHandler) updatePostgresConfig(c *fiber.Ctx) error {
	var body struct {
		Host             string `json:"host"`
		Port             string `json:"port"`
		User             string `json:"user"`
		Password         string `json:"password"`
		Database         string `json:"database"`
		PostgresPath     string `json:"postgresPath"`
		PostgresDataPath string `json:"postgresDataPath"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	v := api.NewValidator()
	v.Required("host", body.Host)
	v.Required("port", body.Port)
	v.Required("user", body.User)
	v.Required("database", body.Database)
	v.Required("postgresPath", body.PostgresPath)
	v.Required("postgresDataPath", body.PostgresDataPath)
	if v.HasErrors() {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	updates := map[string]string{
		"POSTGRESQL_DB_HOST":     body.Host,
		"POSTGRESQL_DB_PORT":     body.Port,
		"POSTGRESQL_DB_USER":     body.User,
		"POSTGRESQL_DB_PASSWORD": body.Password,
		"POSTGRESQL_DB_NAME":     body.Database,
		"POSTGRESQL_FOLDER_PATH": body.PostgresPath,
		"POSTGRESQL_DATA_PATH":   body.PostgresDataPath,
	}

	if err := updateEnvVariables(updates); err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to update env variables: %v", err)
		return api.SendError(c, api.ErrEnvUpdateFailed)
	}

	h.cfg.DBHost = body.Host
	h.cfg.DBPort = body.Port
	h.cfg.DBUser = body.User
	h.cfg.DBPass = body.Password
	h.cfg.DBName = body.Database
	h.cfg.PostgresPath = body.PostgresPath
	h.cfg.PostgresDataPath = body.PostgresDataPath
	h.cfg.DBConnString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		body.User, body.Password, body.Host, body.Port, body.Database)

	logger.Success("[SETTINGS/POSTGRES] PostgreSQL configuration updated")
	return api.Success(c, api.SuccessPostgresConfigUpdated)
}

func (h *backupHandler) findPostgresConf() (string, error) {
	if h.cfg.PostgresPath == "" {
		return "", fmt.Errorf("PostgresPath root is not configured")
	}

	var possiblePaths []string
	if h.cfg.PostgresDataPath != "" {
		possiblePaths = append(possiblePaths, filepath.Join(h.cfg.PostgresDataPath, "postgresql.conf"))
	}

	possiblePaths = append(possiblePaths,
		filepath.Join(h.cfg.PostgresPath, "postgresql.conf"),
		filepath.Join(h.cfg.PostgresPath, "data", "postgresql.conf"),
	)

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	entries, _ := os.ReadDir(h.cfg.PostgresPath)
	for _, entry := range entries {
		if entry.IsDir() && (entry.Name() == "data" || entry.Name() == "16" || entry.Name() == "15" || entry.Name() == "17" || entry.Name() == "18") {
			p := filepath.Join(h.cfg.PostgresPath, entry.Name(), "postgresql.conf")
			if _, err := os.Stat(p); err == nil {
				return p, nil
			}
		}
	}

	return "", fmt.Errorf("postgresql.conf not found in the specified root path")
}

func (h *backupHandler) getPostgresConf(c *fiber.Ctx) error {
	confPath, err := h.findPostgresConf()
	if err != nil {
		logger.Warn("[SETTINGS/POSTGRES] postgresql.conf not found")
		return api.SendError(c, api.ErrPostgresConfNotFound)
	}

	content, err := os.ReadFile(confPath)
	if err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to read postgresql.conf: %v", err)
		return api.SendError(c, api.ErrCoreInternalServer)
	}

	sections := postgresql.ParseConf(string(content))

	return api.Success(c, fiber.Map{
		"path":     confPath,
		"sections": sections,
	})
}

func (h *backupHandler) updatePostgresConf(c *fiber.Ctx) error {
	var body struct {
		Sections []postgresql.ConfSection `json:"sections"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody)
	}

	content := postgresql.GenerateConf(body.Sections)

	confPath, err := h.findPostgresConf()
	if err != nil {
		return api.SendError(c, api.ErrPostgresConfNotFound)
	}

	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to update postgresql.conf: %v", err)
		return api.SendError(c, api.ErrCoreInternalServer)
	}

	logger.Success("[SETTINGS/POSTGRES] postgresql.conf updated successfully")
	return api.Success(c, api.SuccessPostgresConfUpdated)
}
