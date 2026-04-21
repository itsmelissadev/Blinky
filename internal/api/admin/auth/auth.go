package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"blinky/internal/api"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logger.Error("[ADMIN/AUTH] Critical: Failed to generate secure token: %v", err)
		panic("secure random generation failed")
	}
	return hex.EncodeToString(b)
}

func RegisterRoutes(router fiber.Router, db *pgxpool.Pool, isInitialized *bool) {
	admins := router.Group("/admins")

	admins.Get("/initialized", func(c *fiber.Ctx) error {
		var count int
		sqlStr, args, err := database.Query().Select(database.FuncCountAll).From(database.SQLTableAdmins).ToSql()
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		err = tx.QueryRow(ctx, sqlStr, args...).Scan(&count)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Setup check failed: %v", err)
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if count > 0 && isInitialized != nil {
			*isInitialized = true
		}

		return api.Success(c, fiber.Map{"success": count > 0})
	})

	admins.Post("/login", func(c *fiber.Ctx) error {
		type LoginReq struct {
			Identifier string `json:"identifier"`
			Password   string `json:"password"`
		}
		var req LoginReq
		if err := c.BodyParser(&req); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		val := api.NewValidator()
		val.Required("identifier", req.Identifier)
		val.Required("password", req.Password)
		if val.HasErrors() {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		var id, hash string
		sqlStr, args, err := database.Query().
			Select(database.SQLTableAdminsID, database.SQLTableAdminsPasswordHash).
			From(database.SQLTableAdmins).
			Where(sq.Or{
				sq.Eq{database.SQLTableAdminsUsername: req.Identifier},
				sq.Eq{database.SQLTableAdminsEmail: req.Identifier},
			}).
			Limit(1).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		err = tx.QueryRow(ctx, sqlStr, args...).Scan(&id, &hash)
		if err != nil {
			logger.Warn("[ADMIN/AUTH] Login failed: invalid credentials for %s", req.Identifier)
			return api.SendError(c, api.ErrAuthInvalidCredentials, 401)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			logger.Warn("[ADMIN/AUTH] Login failed: password mismatch for %s", id)
			return api.SendError(c, api.ErrAuthInvalidCredentials, 401)
		}

		token := generateToken()
		sqlStrUpd, argsUpd, err := database.Query().
			Update(database.SQLTableAdmins).
			Set(database.SQLTableAdminsToken, token).
			Where(sq.Eq{database.SQLTableAdminsID: id}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		_, err = tx.Exec(ctx, sqlStrUpd, argsUpd...)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Failed to update admin token: %v", err)
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "admin_token",
			Value:    token,
			Path:     "/",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			MaxAge:   86400 * 30,
		})

		logger.Success("[ADMIN/AUTH] Admin logged in successfully: %s", id)
		return api.Success(c, api.SuccessAdminLoggedIn)
	})

	admins.Post("/user", func(c *fiber.Ctx) error {
		var count int
		ctx := c.Context()
		sqlStrCount, argsCount, err := database.Query().Select(database.FuncCountAll).From(database.SQLTableAdmins).ToSql()
		if err == nil {
			db.QueryRow(ctx, sqlStrCount, argsCount...).Scan(&count)
		}

		if count > 0 {
			return Middleware(db)(c)
		}
		return c.Next()
	}, func(c *fiber.Ctx) error {
		type CreateReq struct {
			Nickname string `json:"nickname"`
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
			Avatar   string `json:"avatar"`
		}
		var req CreateReq
		if err := c.BodyParser(&req); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		v := api.NewValidator()
		v.Required("nickname", req.Nickname).MinLength("nickname", req.Nickname, 3)
		v.Required("username", req.Username).MinLength("username", req.Username, 3)
		v.Required("password", req.Password).MinLength("password", req.Password, 6)
		v.Required("email", req.Email).Email("email", req.Email)

		if v.HasErrors() {
			return api.SendError(c, api.ErrAuthInvalidFields, 400)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Password hashing failed: %v", err)
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		id := api.GenerateID(16)
		now := time.Now()

		sqlStrIns, argsIns, _ := database.Query().
			Insert(database.SQLTableAdmins).
			Columns(
				database.SQLTableAdminsID,
				database.SQLTableAdminsNickname,
				database.SQLTableAdminsUsername,
				database.SQLTableAdminsAvatar,
				database.SQLTableAdminsEmail,
				database.SQLTableAdminsPasswordHash,
				database.SQLTableAdminsCreatedAt,
				database.SQLTableAdminsUpdatedAt,
			).
			Values(id, req.Nickname, req.Username, req.Avatar, req.Email, string(hash), now, now).
			ToSql()

		_, err = tx.Exec(ctx, sqlStrIns, argsIns...)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Failed to create admin: %v", err)
			dbErr := api.HandleDBError(err)
			if dbErr.Code == api.ErrCollectionUniqueViolation.Code {
				return api.SendError(c, api.ErrAuthConflict, 409)
			}
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if isInitialized != nil {
			*isInitialized = true
		}

		logger.Success("[ADMIN/AUTH] Admin created: %s (%s)", req.Username, id)
		return api.Success(c, fiber.Map{"id": id}, 201)
	})

	admins.Get("/me", Middleware(db), func(c *fiber.Ctx) error {
		token := c.Cookies("admin_token")
		var user struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
			Email    string `json:"email"`
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		sqlStr, args, err := database.Query().
			Select(
				database.SQLTableAdminsID,
				database.SQLTableAdminsNickname,
				database.SQLTableAdminsUsername,
				database.Coalesce(database.SQLTableAdminsAvatar, database.SQLEmptyString),
				database.SQLTableAdminsEmail,
			).
			From(database.SQLTableAdmins).
			Where(sq.Eq{database.SQLTableAdminsToken: token}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		err = tx.QueryRow(ctx, sqlStr, args...).Scan(&user.ID, &user.Nickname, &user.Username, &user.Avatar, &user.Email)
		if err != nil {
			logger.Warn("[ADMIN/AUTH] Me route failed: user not found for token")
			return api.SendError(c, api.ErrAuthAdminNotFound, 404)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, user)
	})

	admins.Get("/user/:username", Middleware(db), func(c *fiber.Ctx) error {
		username := c.Params("username")
		var user struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
			Email    string `json:"email"`
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		sqlStr, args, err := database.Query().
			Select(
				database.SQLTableAdminsID,
				database.SQLTableAdminsNickname,
				database.SQLTableAdminsUsername,
				database.Coalesce(database.SQLTableAdminsAvatar, database.SQLEmptyString),
				database.SQLTableAdminsEmail,
			).
			From(database.SQLTableAdmins).
			Where(sq.Eq{database.SQLTableAdminsUsername: username}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		err = tx.QueryRow(ctx, sqlStr, args...).Scan(&user.ID, &user.Nickname, &user.Username, &user.Avatar, &user.Email)
		if err != nil {
			logger.Warn("[ADMIN/AUTH] Get user failed: %s not found", username)
			return api.SendError(c, api.ErrAuthAdminNotFound, 404)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, user)
	})

	admins.Patch("/user", Middleware(db), func(c *fiber.Ctx) error {
		type UpdateReq struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
			Username string `json:"username"`
			Avatar   string `json:"avatar"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var req UpdateReq
		if err := c.BodyParser(&req); err != nil || req.ID == "" {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		v := api.NewValidator()
		v.Required("id", req.ID)
		if req.Email != "" {
			v.Email("email", req.Email)
		}
		if v.HasErrors() {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		builder := database.Query().Update(database.SQLTableAdmins).Where(sq.Eq{database.SQLTableAdminsID: req.ID})
		builder = builder.Set(database.SQLTableAdminsUpdatedAt, time.Now())

		if req.Nickname != "" {
			builder = builder.Set(database.SQLTableAdminsNickname, req.Nickname)
		}
		if req.Username != "" {
			builder = builder.Set(database.SQLTableAdminsUsername, req.Username)
		}
		if req.Avatar != "" {
			builder = builder.Set(database.SQLTableAdminsAvatar, req.Avatar)
		}
		if req.Email != "" {
			builder = builder.Set(database.SQLTableAdminsEmail, req.Email)
		}
		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
			if err != nil {
				logger.Error("[ADMIN/AUTH] Update password hashing failed: %v", err)
				return api.SendError(c, api.ErrCoreInternalServer, 500)
			}
			builder = builder.Set(database.SQLTableAdminsPasswordHash, string(hash))
		}

		sql, args, err := builder.ToSql()
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Admin update failed for %s: %v", req.ID, err)
			return api.SendError(c, api.ErrAuthUpdateFailed, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/AUTH] Admin updated: %s", req.ID)
		return api.Success(c, api.SuccessAdminUpdated)
	})

	admins.Delete("/user", Middleware(db), func(c *fiber.Ctx) error {
		type DeleteReq struct {
			ID string `json:"id"`
		}
		var req DeleteReq
		if err := c.BodyParser(&req); err != nil || req.ID == "" {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		token := c.Cookies("admin_token")
		var currentUserID string
		sqlStrCur, argsCur, err := database.Query().
			Select(database.SQLTableAdminsID).
			From(database.SQLTableAdmins).
			Where(sq.Eq{database.SQLTableAdminsToken: token}).
			ToSql()

		if err == nil {
			tx.QueryRow(ctx, sqlStrCur, argsCur...).Scan(&currentUserID)
		}
		if currentUserID == req.ID {
			logger.Warn("[ADMIN/AUTH] Admin %s tried to delete themselves", currentUserID)
			return api.SendError(c, api.ErrAuthDeleteSelfForbidden, 403)
		}

		var count int
		sqlStrCount, argsCount, err := database.Query().
			Select(database.FuncCountAll).
			From(database.SQLTableAdmins).
			ToSql()

		if err == nil {
			tx.QueryRow(ctx, sqlStrCount, argsCount...).Scan(&count)
		}
		if count <= 1 {
			logger.Warn("[ADMIN/AUTH] Attempt to delete the last admin account")
			return api.SendError(c, api.ErrAuthDeleteLastAdminForbidden, 403)
		}

		sqlStrDel, argsDel, err := database.Query().
			Delete(database.SQLTableAdmins).
			Where(sq.Eq{database.SQLTableAdminsID: req.ID}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		_, err = tx.Exec(ctx, sqlStrDel, argsDel...)
		if err != nil {
			logger.Error("[ADMIN/AUTH] Failed to delete admin %s: %v", req.ID, err)
			return api.SendError(c, api.ErrAuthDeleteFailed, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/AUTH] Admin deleted: %s", req.ID)
		return api.Success(c, api.SuccessAdminDeleted)
	})

	admins.Post("/logout", Middleware(db), func(c *fiber.Ctx) error {
		token := c.Cookies("admin_token")

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if token != "" {
			sqlStrUpd, argsUpd, err := database.Query().
				Update(database.SQLTableAdmins).
				Set(database.SQLTableAdminsToken, sq.Expr(database.SQLNull)).
				Where(sq.Eq{database.SQLTableAdminsToken: token}).
				ToSql()

			if err == nil {
				tx.Exec(ctx, sqlStrUpd, argsUpd...)
			}
			logger.Info("[ADMIN/AUTH] Admin logged out")
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "admin_token",
			Value:    "",
			Path:     "/",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Lax",
			Expires:  time.Now().Add(-1 * time.Hour),
		})

		return api.Success(c, api.SuccessAdminLoggedOut)
	})

	admins.Get("/", Middleware(db), func(c *fiber.Ctx) error {
		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		sqlStrList, argsList, err := database.Query().
			Select(
				database.SQLTableAdminsID,
				database.SQLTableAdminsNickname,
				database.SQLTableAdminsUsername,
				database.SQLTableAdminsEmail,
				database.Coalesce(database.SQLTableAdminsAvatar, database.SQLEmptyString),
				database.SQLTableAdminsCreatedAt,
				database.SQLTableAdminsUpdatedAt,
			).
			From(database.SQLTableAdmins).
			OrderBy(database.NewStatement(database.SQLTableAdminsCreatedAt, database.SQLAsc).String()).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		rows, err := tx.Query(ctx, sqlStrList, argsList...)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer rows.Close()

		var results []map[string]interface{}
		for rows.Next() {
			var id, nickname, username, email, avatar string
			var createdAt, updatedAt time.Time
			if err := rows.Scan(&id, &nickname, &username, &email, &avatar, &createdAt, &updatedAt); err == nil {
				if results == nil {
					results = make([]map[string]interface{}, 0)
				}
				results = append(results, map[string]interface{}{
					"id":         id,
					"nickname":   nickname,
					"username":   username,
					"email":      email,
					"avatar":     avatar,
					"created_at": createdAt,
					"updated_at": updatedAt,
				})
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		if results == nil {
			results = make([]map[string]interface{}, 0)
		}

		return api.SuccessList(c, results, api.ListMeta{
			Total:  len(results),
			Limit:  len(results),
			Offset: 0,
		})
	})
}
