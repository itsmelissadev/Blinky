package auth

import (
	"blinky/internal/api"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Middleware(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("admin_token")
		if token == "" {
			return api.SendError(c, api.ErrAuthRequired, 401)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Failed to start middleware transaction: %v", err)
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		var count int
		sqlStr, args, err := database.Query().
			Select(database.FuncCountAll).
			From(database.SQLTableAdmins).
			Where(squirrel.Eq{database.SQLTableAdminsToken: token}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		err = tx.QueryRow(ctx, sqlStr, args...).Scan(&count)
		if err != nil || count == 0 {
			return api.SendError(c, api.ErrAuthInvalidSession, 401)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		c.Locals("admin_token", token)
		return c.Next()
	}
}
