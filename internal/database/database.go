package database

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"blinky/internal/config"
	"blinky/internal/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DBConnString)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	err = pool.Ping(ctx)
	if err == nil {
		logger.Success("[DATABASE/CONNECTION] Verified connection to %s", cfg.DBName)
		return pool, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "3D000" {
		if !AskForApproval(fmt.Sprintf("Database '%s' does not exist. Create it? (y/N): ", cfg.DBName)) {
			return nil, fmt.Errorf("database creation aborted by user")
		}

		pool.Close()
		defaultConnString := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable",
			cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort)

		defaultPool, err := pgxpool.New(ctx, defaultConnString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to default postgres db: %w", err)
		}
		defer defaultPool.Close()

		sql := NewStatement(SQLCreateDatabase).Add(Quote(cfg.DBName)).String()

		_, err = defaultPool.Exec(ctx, sql)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		logger.Success("[DATABASE/CREATE] Database '%s' created successfully", cfg.DBName)

		return pgxpool.New(ctx, cfg.DBConnString)
	}

	return nil, fmt.Errorf("database is not responding: %w", err)
}

func AskForApproval(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func InitSystemTables(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start migration transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, GetCollectionsTableSQL()); err != nil {
		return fmt.Errorf("failed to create _collections table: %w", err)
	}

	var existingCorrectColumns int
	checkSQL, args, err := Query().
		Select(FuncCountAll).
		From(SQLTableInformationColumns).
		Where(squirrel.Eq{
			SQLColumnTableName:  SQLTableCollections,
			SQLColumnIsNullable: "NO",
		}).
		Where(squirrel.Eq{SQLColumnColumnName: []string{SQLTableCollectionsTotalDocuments, SQLTableCollectionsTotalBytes, SQLTableCollectionsIsSystem}}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build migration check query: %w", err)
	}

	if err := tx.QueryRow(ctx, checkSQL, args...).Scan(&existingCorrectColumns); err != nil {
		return fmt.Errorf("failed to check existing columns: %w", err)
	}

	if existingCorrectColumns < 3 {
		logger.Warn("[DATABASE/MIGRATION] Schema changes detected. Upgrade required.")
		if !AskForApproval("Do you want to proceed with the migration? (y/N): ") {
			return fmt.Errorf("migration aborted by user")
		}

		alterSQL := AlterTable(SQLTableCollections).
			AddColumn(NewColumn(SQLTableCollectionsTotalDocuments).BigInt().NotNull().Default(0).Build()).
			SetDefault(SQLTableCollectionsTotalDocuments, 0).
			SetNotNull(SQLTableCollectionsTotalDocuments).
			AddColumn(NewColumn(SQLTableCollectionsTotalBytes).BigInt().NotNull().Default(0).Build()).
			SetDefault(SQLTableCollectionsTotalBytes, 0).
			SetNotNull(SQLTableCollectionsTotalBytes).
			AddColumn(NewColumn(SQLTableCollectionsIsSystem).Boolean(true).NotNull().Build()).
			SetDefault(SQLTableCollectionsIsSystem, SQLFalse).
			SetNotNull(SQLTableCollectionsIsSystem).
			Build()

		if _, err := tx.Exec(ctx, alterSQL); err != nil {
			return fmt.Errorf("failed to alter _collections: %w", err)
		}

		logger.Success("[DATABASE/MIGRATION] Upgrade completed successfully")
	}

	return tx.Commit(ctx)
}
