package main

import (
	"blinky/internal/api/admin"
	"blinky/internal/api/admin/settings"
	"blinky/internal/api/admin/system"
	"blinky/internal/api/public"
	"blinky/internal/config"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/ssh"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	Config *config.Config
	DB     *pgxpool.Pool
}

func main() {
	cfg := config.LoadConfig()

	logger.Info("[ENGINE] Blinky v1.0.0-alpha (Code: 1) starting...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		logger.Error("[ENGINE] Failed to connect to database: %v", err)
		os.Exit(1)
	}

	if err := database.Migrate(ctx, db, cfg, settings.CreateBackup); err != nil {
		logger.Error("[ENGINE] Database migration failed: %v", err)
		os.Exit(1)
	}

	e := &Engine{
		Config: cfg,
		DB:     db,
	}

	if err := e.Run(ctx); err != nil {
		logger.Error("[ENGINE] Runtime error: %v", err)
		os.Exit(1)
	}

	logger.Warn("[ENGINE] Termination signal received. Initiating graceful shutdown...")
	e.DB.Close()
	logger.Info("[ENGINE] Database connection pool closed")
	logger.Info("[ENGINE] Blinky has exited gracefully")
}

func (e *Engine) Run(ctx context.Context) error {
	publicApp := public.NewPublicApp(e.DB)
	adminApp := admin.NewAdminApp(e.DB, e.Config)

	system.SetPublicApp(publicApp)

	if e.Config.AdminSSHEnabled {
		e.Config.AdminPanelHost = "127.0.0.1"
	}
	if e.Config.PublicSSHEnabled {
		e.Config.PublicAPIHost = "127.0.0.1"
	}

	if e.Config.AdminSSHEnabled || e.Config.PublicSSHEnabled {
		go ssh.StartSSHServer(e.Config)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		addr := fmt.Sprintf("%s:%s", e.Config.AdminPanelHost, e.Config.AdminPanelPort)
		logger.Info("[ENGINE] Admin Panel listening on %s", addr)
		if err := adminApp.Listen(addr); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		addr := fmt.Sprintf("%s:%s", e.Config.PublicAPIHost, e.Config.PublicAPIPort)
		logger.Info("[ENGINE] Public API listening on %s", addr)
		if err := publicApp.Listen(addr); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Warn("[ENGINE] Stopping services...")
		_ = adminApp.Shutdown()
		_ = publicApp.Shutdown()
		return nil
	})

	return g.Wait()
}
