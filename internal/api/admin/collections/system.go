package collections

import (
	"context"
	"encoding/json"
	"fmt"

	"blinky/internal/api/admin/collections/tables"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitSystemCollections(ctx context.Context, db *pgxpool.Pool) error {
	var count int
	sql, args, err := database.Query().
		Select(database.FuncCountAll).
		From(database.SQLTableCollections).
		Where(sq.Eq{database.SQLTableCollectionsName: database.SQLTableAdmins}).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build _admins existence check: %w", err)
	}

	err = db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check _admins existence: %w", err)
	}

	if count > 0 {
		return nil
	}

	logger.Info("[ENGINE/INIT] Initializing _admins system collection...")

	schema := tables.GetAdminsSchema()

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	createSQL := database.GetAdminsTableSQL()
	if _, err := tx.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("failed to create %s base table: %w", database.SQLTableAdmins, err)
	}

	for _, field := range schema {
		if field.Type == types.TypeID || field.Name == database.SQLTableAdminsID {
			continue
		}
		if err := addColumn(ctx, tx, database.SQLTableAdmins, field); err != nil {
			return fmt.Errorf("failed to add field %s to %s: %w", field.Name, database.SQLTableAdmins, err)
		}
	}

	schemaJSON, _ := json.Marshal(schema)
	metaSql, metaArgs, err := database.Query().
		Insert(database.SQLTableCollections).
		Columns(database.SQLTableCollectionsName, database.SQLTableCollectionsSchema, database.SQLTableCollectionsIsSystem).
		Values(database.SQLTableAdmins, schemaJSON, true).
		ToSql()

	if err != nil {
		return fmt.Errorf("failed to build metadata insert for %s: %w", database.SQLTableAdmins, err)
	}

	if _, err := tx.Exec(ctx, metaSql, metaArgs...); err != nil {
		return fmt.Errorf("failed to insert metadata for %s: %w", database.SQLTableAdmins, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit %s creation: %w", database.SQLTableAdmins, err)
	}

	logger.Success("[ENGINE/INIT] System collection %s initialized", database.SQLTableAdmins)
	return nil
}
