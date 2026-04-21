package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blinky/internal/api/admin"
	"blinky/internal/api/admin/collections"
	"blinky/internal/api/admin/system"
	"blinky/internal/api/public"
	"blinky/internal/config"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/worker"

	"golang.org/x/sync/errgroup"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Engine struct {
	Config  *config.Config
	DB      *pgxpool.Pool
	Workers *worker.Manager
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logger.Init(*debug)
	logger.Info("[ENGINE] Blinky starting...")

	cfg := config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var dbPool *pgxpool.Pool
	if cfg.IsEnvExist {
		var err error
		dbPool, err = database.Connect(ctx, cfg)
		if err != nil {
			logger.Error("[ENGINE] Fatal: Database connection failed: %v", err)
			os.Exit(1)
		}

		if err := database.InitSystemTables(ctx, dbPool); err != nil {
			logger.Error("[ENGINE] Fatal: System table initialization failed: %v", err)
			os.Exit(1)
		}

		if err := collections.InitSystemCollections(ctx, dbPool); err != nil {
			logger.Error("[ENGINE] Fatal: System collections initialization failed: %v", err)
			os.Exit(1)
		}
	} else {
		logger.Warn("[ENGINE] No .env found. Starting in Setup Mode...")
	}

	engine := &Engine{
		Config:  cfg,
		DB:      dbPool,
		Workers: worker.NewManager(),
	}

	if err := engine.Run(ctx); err != nil {
		logger.Error("[ENGINE] Failure: %v", err)
		os.Exit(1)
	}

	logger.Info("[ENGINE] Blinky has exited gracefully")
}

func (e *Engine) Run(ctx context.Context) error {
	publicApp := public.NewPublicApp(e.DB)
	adminApp := admin.NewAdminApp(e.DB, e.Config)

	system.SetPublicApp(publicApp)

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

	e.Workers.Start(ctx)

	logger.Success("[ENGINE] All services are active and healthy")

	g.Go(func() error {
		<-ctx.Done()
		logger.Warn("[ENGINE] Termination signal received. Initiating graceful shutdown...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var shutdownErr error
		if err := adminApp.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("[ENGINE] Admin API shutdown failed: %v", err)
			shutdownErr = err
		}
		if err := publicApp.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("[ENGINE] Public API shutdown failed: %v", err)
			shutdownErr = err
		}

		e.Workers.Wait()
		if e.DB != nil {
			e.DB.Close()
			logger.Info("[ENGINE] Database connection pool closed")
		}

		return shutdownErr
	})

	return g.Wait()
}
