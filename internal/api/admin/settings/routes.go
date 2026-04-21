package settings

import (
	"blinky/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router fiber.Router, db *pgxpool.Pool, cfg *config.Config) {
	h := &backupHandler{
		db:  db,
		cfg: cfg,
	}

	g := router.Group("/settings/backup")
	g.Get("/", h.list)
	g.Post("/", h.create)
	g.Delete("/:filename", h.delete)
	g.Get("/download/:filename", h.download)
	g.Get("/config", h.getConfig)
	g.Patch("/config", h.updateConfig)

	e := router.Group("/settings/environments")
	e.Get("/", h.getEnv)
	e.Patch("/", h.updateEnv)
	e.Delete("/", h.deleteEnv)

	p := router.Group("/settings/postgresql")
	p.Get("/", h.getPostgresConfig)
	p.Post("/test", h.testPostgresConnection)
	p.Patch("/", h.updatePostgresConfig)
	p.Get("/conf", h.getPostgresConf)
	p.Patch("/conf", h.updatePostgresConf)

	s := router.Group("/settings/server")
	s.Get("/", h.getServerConfig)
	s.Patch("/", h.updateServerConfig)
}

type backupHandler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}
