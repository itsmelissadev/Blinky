package panel

import (
	"blinky/internal/panel"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewPanelApp() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[PANEL] [${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

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

	return app
}
