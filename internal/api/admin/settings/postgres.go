package settings

import (
	"blinky/internal/api"
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/pathutil"
	"blinky/internal/pkg/postgresql"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (h *backupHandler) getPostgresConfig(c *fiber.Ctx) error {
	return api.Success(c, fiber.Map{
		"host":     h.cfg.PostgresDBHost,
		"port":     h.cfg.PostgresDBPort,
		"user":     h.cfg.PostgresDBUser,
		"password": h.cfg.PostgresDBPassword,
		"database": h.cfg.PostgresDBName,
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
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	val := api.NewValidator()
	val.Required("host", body.Host)
	val.Required("port", body.Port)
	val.Required("user", body.User)
	val.Required("database", body.Database)
	if val.HasErrors() {
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		body.User, body.Password, body.Host, body.Port, body.Database)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return api.SendError(c, api.ErrCoreInternalServer.WithReplacements(map[string]string{
			"error": err.Error(),
		}), 500)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return api.SendError(c, api.ErrCoreInternalServer.WithReplacements(map[string]string{
			"error": fmt.Sprintf("Ping failed: %v", err),
		}), 500)
	}

	logger.Success("[SETTINGS/POSTGRES] Connection test successful for %s:%s", body.Host, body.Port)
	return api.Success(c, api.SuccessPostgresTestOK)
}

func (h *backupHandler) updatePostgresConfig(c *fiber.Ctx) error {
	var body struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database"`
	}

	if err := c.BodyParser(&body); err != nil {
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	v := api.NewValidator()
	v.Required("host", body.Host)
	v.Required("port", body.Port)
	v.Required("user", body.User)
	v.Required("database", body.Database)
	if v.HasErrors() {
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	updates := map[string]string{
		"POSTGRESQL_DB_HOST":     body.Host,
		"POSTGRESQL_DB_PORT":     body.Port,
		"POSTGRESQL_DB_USER":     body.User,
		"POSTGRESQL_DB_PASSWORD": body.Password,
		"POSTGRESQL_DB_NAME":     body.Database,
	}

	if err := config.UpdateEnvVariables(updates); err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to update env variables: %v", err)
		return api.SendError(c, api.ErrEnvUpdateFailed, 500)
	}

	h.cfg.PostgresDBHost = body.Host
	h.cfg.PostgresDBPort = body.Port
	h.cfg.PostgresDBUser = body.User
	h.cfg.PostgresDBPassword = body.Password
	h.cfg.PostgresDBName = body.Database
	h.cfg.UpdateDBConnString()

	logger.Success("[SETTINGS/POSTGRES] PostgreSQL configuration updated")
	return api.Success(c, api.SuccessPostgresConfigUpdated)
}

func (h *backupHandler) findPostgresConf() (string, error) {
	path := pathutil.Join(pathutil.GetPostgresDataPath(), "postgresql.conf")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("postgresql.conf not found at %s", path)
}

func (h *backupHandler) getPostgresConf(c *fiber.Ctx) error {
	confPath, err := h.findPostgresConf()
	if err != nil {
		logger.Warn("[SETTINGS/POSTGRES] postgresql.conf not found")
		return api.SendError(c, api.ErrPostgresConfNotFound, 404)
	}

	content, err := os.ReadFile(confPath)
	if err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to read postgresql.conf: %v", err)
		return api.SendError(c, api.ErrCoreInternalServer, 500)
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
		return api.SendError(c, api.ErrCoreInvalidBody, 400)
	}

	content := postgresql.GenerateConf(body.Sections)

	confPath, err := h.findPostgresConf()
	if err != nil {
		return api.SendError(c, api.ErrPostgresConfNotFound, 404)
	}

	if err := os.WriteFile(confPath, []byte(content), 0644); err != nil {
		logger.Error("[SETTINGS/POSTGRES] Failed to update postgresql.conf: %v", err)
		return api.SendError(c, api.ErrCoreInternalServer, 500)
	}

	logger.Success("[SETTINGS/POSTGRES] postgresql.conf updated successfully")
	return api.Success(c, api.SuccessPostgresConfUpdated)
}
