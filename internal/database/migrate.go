package database

import (
	"blinky/internal"
	"blinky/internal/config"
	"blinky/internal/pkg/logger"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MigrationFunc func(ctx context.Context, tx pgx.Tx) error

type Migration struct {
	Version       int
	Name          string
	Up            MigrationFunc
	Down          MigrationFunc
	NoTransaction bool
}

type BackupFunc func(cfg *config.Config) (string, error)

var migrations = []Migration{
	{
		Version: 1,
		Name:    "Initialize system tables",
		Up: func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, GetCollectionsTableSQL()); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, GetAdminsTableSQL()); err != nil {
				return err
			}
			return nil
		},
		Down: func(ctx context.Context, tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, NewStatement(SQLDropTable, Quote(SQLTableCollections)).String()); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, NewStatement(SQLDropTable, Quote(SQLTableAdmins)).String()); err != nil {
				return err
			}
			return nil
		},
	},
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config, backupFn BackupFunc) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, GetMigrationsTableSQL()); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, GetMigrationLogsTableSQL()); err != nil {
		return err
	}

	var currentVersion int
	query, args, _ := Query().
		Select(Coalesce(Max(SQLTableMigrationsVersion), SQLZero)).
		From(SQLTableMigrations).
		ToSql()

	if err := tx.QueryRow(ctx, query, args...).Scan(&currentVersion); err != nil {
		return err
	}

	var pending []Migration
	for _, m := range migrations {
		if m.Version > currentVersion && m.Version <= internal.AppDatabaseVersion {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		return tx.Commit(ctx)
	}

	if currentVersion > 0 {
		logger.Warn("[DATABASE/MIGRATION] %d pending migrations found (Current Version: v%d)", len(pending), currentVersion)
		if !AskForApproval("Would you like to apply these migrations sequentially? (y/N): ") {
			return fmt.Errorf("migration aborted by user")
		}

		logger.Info("[DATABASE/MIGRATION] Security check: Creating automatic backup before proceeding...")
		if _, err := backupFn(cfg); err != nil {
			logger.Error("[DATABASE/MIGRATION] Fatal: Automatic backup failed: %v", err)
			logger.Error("[DATABASE/MIGRATION] Migration aborted to prevent data loss. Exiting...")
			os.Exit(1)
		} else {
			logger.Success("[DATABASE/MIGRATION] Safety backup created successfully.")
		}
	} else {
		logger.Info("[DATABASE/SETUP] Initializing system schema (Version: v%d)...", internal.AppDatabaseVersion)
	}

	for _, m := range pending {
		startTime := time.Now()
		logger.Info("[DATABASE/PROCESS] Applying Version %d: %s...", m.Version, m.Name)

		status := "SUCCESS"
		message := "Applied successfully"
		var migrationErr error

		if m.NoTransaction {

			migrationErr = m.Up(ctx, nil)
		} else {
			migrationErr = m.Up(ctx, tx)
		}

		if migrationErr != nil {
			status = "ERROR"
			message = migrationErr.Error()
			logger.Error("[DATABASE/PROCESS] Failed at Version %d: %v", m.Version, migrationErr)
		}

		duration := time.Since(startTime).Milliseconds()

		logSQL, lArgs, _ := Query().
			Insert(SQLTableMigrationLogs).
			Columns(SQLTableMigrationLogsID, SQLTableMigrationLogsVersion, SQLTableMigrationLogsStatus, SQLTableMigrationLogsMessage, SQLTableMigrationLogsDuration).
			Values(uuid.New().String(), m.Version, status, message, duration).
			ToSql()

		if _, err := tx.Exec(ctx, logSQL, lArgs...); err != nil {
			return fmt.Errorf("failed to record migration log: %w", err)
		}

		if migrationErr != nil {
			return migrationErr
		}

		insertSQL, iArgs, _ := Query().
			Insert(SQLTableMigrations).
			Columns(SQLTableMigrationsID, SQLTableMigrationsVersion).
			Values(uuid.New().String(), m.Version).
			ToSql()

		if _, err := tx.Exec(ctx, insertSQL, iArgs...); err != nil {
			return err
		}
		logger.Success("[DATABASE/PROCESS] Version %d applied successfully", m.Version)
	}

	return tx.Commit(ctx)
}
