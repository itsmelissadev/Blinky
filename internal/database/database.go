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
