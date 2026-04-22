package admin

import (
	"blinky/internal/api"
	"blinky/internal/api/admin/auth"
	"blinky/internal/api/admin/collections"
	"blinky/internal/api/admin/settings"
	"blinky/internal/api/admin/system"
	"blinky/internal/config"
	"blinky/internal/database"
	"blinky/internal/panel"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewAdminApp(db *pgxpool.Pool, cfg *config.Config) *fiber.App {
	var isInitialized bool
	if db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var count int
		sql, args, _ := database.Query().
			Select(database.FuncCountAll).
			From(database.SQLTableAdmins).
			ToSql()
		_ = db.QueryRow(ctx, sql, args...).Scan(&count)
		isInitialized = count > 0
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[ADMIN] [${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool { return true },
		AllowMethods:     "GET, POST, PATCH, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, X-CSRF-Token",
		AllowCredentials: true,
	}))

	apiGroup := app.Group("/_api")

	apiGroup.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_",
		CookieHTTPOnly: false,
		CookieSameSite: "Lax",
		CookieSecure:   false,
		Expiration:     1 * time.Hour,
		Next: func(c *fiber.Ctx) bool {
			path := c.Path()
			if path == "/_api/admins/login" || path == "/_api/admins/initialized" || strings.HasPrefix(path, "/_api/setup") {
				return true
			}
			if path == "/_api/admins/user" && c.Method() == fiber.MethodPost {
				return !isInitialized
			}

			return false
		},
	}))

	apiGroup.Get("/", func(c *fiber.Ctx) error {
		return api.Success(c, api.SuccessAdminOnline)
	})

	system.RegisterSetupRoutes(apiGroup, cfg)

	if db != nil {
		auth.RegisterRoutes(apiGroup, db, &isInitialized)
		apiGroup.Use(auth.Middleware(db))
		collections.RegisterRoutes(apiGroup, db)
		settings.RegisterRoutes(apiGroup, db, cfg)
		system.RegisterRoutes(apiGroup, db, cfg)
		system.RegisterSQLRoutes(apiGroup.Group("/system"), db)
	}

	system.RegisterFileRoutes(apiGroup)

	if cfg.Environment != "development" {
		app.Use("/", filesystem.New(filesystem.Config{
			Root:       http.FS(panel.Assets),
			PathPrefix: "dist",
			Browse:     false,
			Index:      "index.html",
		}))

		app.Get("/*", func(c *fiber.Ctx) error {
			file, err := panel.Assets.ReadFile("dist/index.html")
			if err != nil {
				return c.Status(http.StatusNotFound).SendString("Resource not found")
			}
			c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
			return c.Send(file)
		})
	}

	return app
}
