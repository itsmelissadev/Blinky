package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"blinky/internal/database"
	"blinky/internal/types"
	"encoding/json"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetCollectionDetail(ctx context.Context, tx pgx.Tx, name string) (*types.CollectionDetail, error) {
	sql, args, err := database.Query().
		Select(
			database.SQLTableCollectionsName,
			database.SQLTableCollectionsSchema,
			database.SQLTableCollectionsIsSystem,
			database.SQLTableCollectionsTotalDocuments,
			database.SQLTableCollectionsTotalBytes,
			database.SQLTableCollectionsCreatedAt,
			database.SQLTableCollectionsUpdatedAt,
		).
		From(database.SQLTableCollections).
		Where(sq.Eq{database.SQLTableCollectionsName: name}).ToSql()

	if err != nil {
		return nil, err
	}

	var col types.CollectionDetail
	var schemaJSON []byte
	err = tx.QueryRow(ctx, sql, args...).Scan(
		&col.Name,
		&schemaJSON,
		&col.IsSystem,
		&col.TotalDocuments,
		&col.TotalBytes,
		&col.CreatedAt,
		&col.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(schemaJSON, &col.Schema); err != nil {
		col.Schema = types.CollectionSchema{}
	}

	return &col, nil
}

func ValidateName(name string) bool {
	match, _ := regexp.MatchString(`^[a-z0-9_]+$`, name)
	return match
}

func GenerateID(length int) string {
	if length <= 0 {
		length = 16
	}
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {

			panic(fmt.Sprintf("failed to generate secure random number: %v", err))
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func FormatUUID(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	if b, ok := val.([16]byte); ok {
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	} else if b, ok := val.([]byte); ok && len(b) == 16 {
		return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	}
	return val
}

func HandleDBError(err error) ApiError {
	if err == nil {
		return ErrCoreInternalServer
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "SQLSTATE 23505") {
		return ErrCollectionUniqueViolation
	} else if strings.Contains(errMsg, "SQLSTATE 23502") {
		return ErrCollectionNotNullViolation
	} else if strings.Contains(errMsg, "SQLSTATE 23514") {
		return ErrCollectionCheckViolation
	} else if strings.Contains(errMsg, "SQLSTATE 23503") {
		return ErrCollectionForeignKeyViolation
	}

	return ErrCollectionQueryExecution
}
func IsSystemCollection(ctx context.Context, db *pgxpool.Pool, name string) bool {
	var isSystem bool
	sql, args, err := database.Query().
		Select(database.SQLTableCollectionsIsSystem).
		From(database.SQLTableCollections).
		Where(sq.Eq{database.SQLTableCollectionsName: name}).
		ToSql()
	if err != nil {
		return false
	}
	err = db.QueryRow(ctx, sql, args...).Scan(&isSystem)
	return err == nil && isSystem
}
