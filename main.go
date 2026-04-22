package main

import (
	"blinky/internal/api/admin"
	"blinky/internal/api/admin/settings"
	"blinky/internal/api/admin/system"
	"blinky/internal/api/public"
	"blinky/internal/config"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/pkg/pathutil"
	"blinky/internal/pkg/postgresql"
	"blinky/internal/pkg/ssh"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	Config *config.Config
	DB     *pgxpool.Pool
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg := config.LoadConfig()

	pgManager := postgresql.NewManager(cfg)

	if _, err := os.Stat(pathutil.GetPostgresDataPath()); os.IsNotExist(err) {
		if cfg.PostgresDBPassword == "" || cfg.PostgresDBPassword == "postgres" {
			fmt.Println("\n[SETUP] Blinky needs to initialize a new managed database.")
			fmt.Println("[SETUP] Please provide secure credentials for the 'postgres' superuser.")

			var user, pass, dbName string
			fmt.Print("Enter Database Username (default: postgres): ")
			fmt.Scanln(&user)
			if strings.TrimSpace(user) == "" {
				user = "postgres"
			}

			for {
				fmt.Print("Enter Secure Password: ")
				fmt.Scanln(&pass)
				pass = strings.TrimSpace(pass)
				if len(pass) >= 8 {
					break
				}
				fmt.Println("[ERROR] Password must be at least 8 characters long.")
			}

			fmt.Print("Enter Database Name (default: blinky_db): ")
			fmt.Scanln(&dbName)
			if strings.TrimSpace(dbName) == "" {
				dbName = "blinky_db"
			}

			updates := map[string]string{
				"POSTGRESQL_DB_USER":     user,
				"POSTGRESQL_DB_PASSWORD": pass,
				"POSTGRESQL_DB_NAME":     dbName,
			}

			if err := config.UpdateEnvVariables(updates); err != nil {
				logger.Error("[ENGINE] Failed to save credentials: %v", err)
				return err
			}

			cfg.PostgresDBUser = user
			cfg.PostgresDBPassword = pass
			cfg.PostgresDBName = dbName
			cfg.IsEnvExist = true
			cfg.UpdateDBConnString()
			logger.Success("[ENGINE] Configuration saved. Proceeding with database initialization...")
		}
	}

	if err := pgManager.Start(); err != nil {
		logger.Error("[ENGINE] Failed to start managed PostgreSQL: %v", err)
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ENGINE] Recovered from panic: %v", r)
		}
		pgManager.Stop()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var db *pgxpool.Pool
	if cfg.IsEnvExist {
		var err error
		db, err = database.Connect(ctx, cfg)
		if err != nil {
			logger.Error("[ENGINE] Failed to connect to database: %v", err)
			return err
		}

		if err := database.Migrate(ctx, db, cfg, settings.CreateBackup); err != nil {
			logger.Error("[ENGINE] Database migration failed: %v", err)
			return err
		}
	} else {
		logger.Warn("[ENGINE] No .env file found. Entering setup mode...")
	}

	e := &Engine{
		Config: cfg,
		DB:     db,
	}

	if err := e.Run(ctx); err != nil {
		logger.Error("[ENGINE] Runtime error: %v", err)
		return err
	}

	logger.Warn("[ENGINE] Termination signal received. Initiating graceful shutdown...")
	if e.DB != nil {
		e.DB.Close()
		logger.Info("[ENGINE] Database connection pool closed")
	}

	logger.Info("[ENGINE] Blinky has exited gracefully")
	return nil
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
