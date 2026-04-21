package collections

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"blinky/internal/api"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(router fiber.Router, db *pgxpool.Pool) {
	registerEntityRoutes(router, db)
	registerMetaRoutes(router, db)
}

func registerEntityRoutes(router fiber.Router, db *pgxpool.Pool) {
	collection := router.Group("/collection")

	collection.Get("/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		col, err := api.GetCollectionDetail(ctx, tx, name)
		if err != nil {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, col)
	})

	collection.Get("/:name/records", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		col, err := api.GetCollectionDetail(ctx, tx, name)
		if err != nil {
			logger.Warn("[ADMIN/COLLECTIONS] Meta fetch failed for %s: %v", name, err)
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		schema := col.Schema
		totalDocuments := int(col.TotalDocuments)

		limit := c.QueryInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}
		if limit < 1 {
			limit = 1
		}
		offset := c.QueryInt("offset", 0)
		if offset < 0 {
			offset = 0
		}

		sql, args, err := database.Query().
			Select(database.SQLAsterisk).
			From(database.Quote(name)).
			Limit(uint64(limit)).
			Offset(uint64(offset)).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Records query failed: %v", err)
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}
		defer rows.Close()

		fields := rows.FieldDescriptions()
		results := make([]map[string]interface{}, 0)
		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				continue
			}

			row := make(map[string]interface{})
			for i, field := range fields {
				val := values[i]

				if field.DataTypeOID == database.OIDUUID {
					val = api.FormatUUID(val)
				}
				row[string(field.Name)] = val
			}
			results = append(results, row)
		}

		PerformJoins(ctx, db, schema, results, false)

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.SuccessList(c, results, api.ListMeta{
			Total:  totalDocuments,
			Limit:  limit,
			Offset: offset,
		})
	})

	collection.Post("/:name/records/bulk-delete", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		type DeleteRequest struct {
			IDs []string `json:"ids"`
		}

		var req DeleteRequest
		if err := c.BodyParser(&req); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		v := api.NewValidator()
		v.Required("ids", req.IDs)
		if v.HasErrors() {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		logger.Info("[ADMIN/COLLECTIONS] Bulk deleting %d records from %s", len(req.IDs), name)

		sql, args, err := database.Query().
			Delete(database.Quote(name)).
			Where(sq.Eq{database.SQLRecordID: req.IDs}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		res, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Bulk delete failed: %v", err)
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}
		rowsAffected := res.RowsAffected()

		updateMetaSQL, args, err := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsTotalDocuments, sq.Expr(database.NewStatement(database.SQLTableCollectionsTotalDocuments).Add(database.SQLMinus).Add("?").String(), rowsAffected)).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err != nil {
			logger.Warn("[ADMIN/COLLECTIONS] Failed to build metadata update query: %v", err)
		} else {
			if _, err := tx.Exec(ctx, updateMetaSQL, args...); err != nil {
				logger.Warn("[ADMIN/COLLECTIONS] Failed to update metadata document count: %v", err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, fiber.Map{"deleted_count": rowsAffected})
	})

	collection.Post("/:name/records", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		var schemaData []byte
		metaSQL, args, err := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if err := tx.QueryRow(ctx, metaSQL, args...).Scan(&schemaData); err != nil {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		var schema []types.CollectionField
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return api.SendError(c, api.ErrCollectionInvalidSchema, 500)
		}

		autoIDLen := 15
		for _, f := range schema {
			nameLower := strings.ToLower(f.Name)
			if (f.Type == types.TypeID || nameLower == "id") && f.Props.ID != nil && f.Props.ID.AutoGenerateLength > 0 {
				autoIDLen = f.Props.ID.AutoGenerateLength
			}
		}

		id := api.GenerateID(autoIDLen)
		now := time.Now()
		insertBuilder := database.Query().
			Insert(database.Quote(name)).
			Columns(database.SQLRecordID)

		vals := []interface{}{id}

		for _, f := range schema {
			nameLower := strings.ToLower(f.Name)
			if nameLower == database.SQLRecordID {
				continue
			}

			val, ok := body[f.Name]
			if !ok {
				val, ok = body[nameLower]
			}

			if f.Type == types.TypeDate && (f.Props.AutoTimestamp == types.AutoTimestampCreate || f.Props.AutoTimestamp == types.AutoTimestampCreateUpdate) {
				if ok && val != nil && val != "" {
					return api.SendError(c, api.ErrCollectionAutoManagedField.WithReplacements(map[string]string{"field": f.Name}), 400)
				}
				val = now
				ok = true
			}

			if ok {
				insertBuilder = insertBuilder.Columns(database.Quote(f.Name))
				vals = append(vals, val)
			}
		}

		insertSQL, args, err := insertBuilder.Values(vals...).ToSql()
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Record insert query build failed for %s: %v", name, err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if _, err := tx.Exec(ctx, insertSQL, args...); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Record insertion failed: %v", err)
			dbErr := api.HandleDBError(err)
			return api.SendError(c, dbErr, 400)
		}

		updateMetaSQL, args, err := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsTotalDocuments, sq.Expr(database.NewStatement(database.SQLTableCollectionsTotalDocuments).Add(database.SQLPlus).Add(database.SQLOne).String())).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err != nil {
			logger.Warn("[ADMIN/COLLECTIONS] Failed to build increment query: %v", err)
		} else {
			if _, err := tx.Exec(ctx, updateMetaSQL, args...); err != nil {
				logger.Warn("[ADMIN/COLLECTIONS] Failed to increment document count: %v", err)
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, fiber.Map{"id": id}, 201)
	})

	collection.Patch("/:name/records/:id", func(c *fiber.Ctx) error {
		name := c.Params("name")
		id := c.Params("id")

		if id == "" {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		var schemaData []byte
		metaSQL, args, err := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if err := tx.QueryRow(ctx, metaSQL, args...).Scan(&schemaData); err != nil {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		var schema []types.CollectionField
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return api.SendError(c, api.ErrCollectionInvalidSchema, 500)
		}

		now := time.Now()
		updateBuilder := database.Query().
			Update(database.Quote(name)).
			Where(sq.Eq{database.SQLRecordID: id})

		hasAllowedUpdates := false

		for k, v := range body {
			kLower := strings.ToLower(k)
			if kLower == database.SQLRecordID {
				continue
			}

			var matchedField *types.CollectionField
			for _, f := range schema {
				if strings.ToLower(f.Name) == kLower {
					matchedField = &f
					break
				}
			}

			if matchedField != nil {
				if matchedField.Type == types.TypeDate && matchedField.Props.AutoTimestamp == types.AutoTimestampCreateUpdate {
					return api.SendError(c, api.ErrCollectionAutoManagedField.WithReplacements(map[string]string{"field": matchedField.Name}), 400)
				}

				if matchedField.Type == types.TypeRelation {
					if v == nil {
						v = []string{}
					} else if s, ok := v.(string); ok {
						v = []string{s}
					}
				}

				updateBuilder = updateBuilder.Set(database.Quote(matchedField.Name), v)
				hasAllowedUpdates = true
			}
		}

		for _, f := range schema {
			if f.Type == types.TypeDate && f.Props.AutoTimestamp == types.AutoTimestampCreateUpdate {
				updateBuilder = updateBuilder.Set(database.Quote(f.Name), now)
				hasAllowedUpdates = true
			}
		}

		if !hasAllowedUpdates {
			return api.Success(c, api.SuccessNoChanges)
		}

		sql, args, err := updateBuilder.ToSql()
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Record update query build failed: %v", err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		res, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Record update failed: %v", err)
			dbErr := api.HandleDBError(err)
			return api.SendError(c, dbErr, 400)
		}

		if res.RowsAffected() == 0 {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, api.SuccessRecordUpdated)
	})

	collection.Post("/", func(c *fiber.Ctx) error {
		type CreateRequest struct {
			Name   string                 `json:"name"`
			Schema types.CollectionSchema `json:"schema"`
		}

		var req CreateRequest
		if err := c.BodyParser(&req); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Body parser failed: %v", err)
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		val := api.NewValidator()
		val.Required("name", req.Name)
		val.Required("schema", req.Schema)
		if val.HasErrors() {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		req.Schema.Sanitize()

		if !api.ValidateName(req.Name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		checkSql, checkArgs, err := database.Query().Select(database.FuncCountAll).
			From(database.SQLTablePgTables).
			Where(sq.Eq{database.SQLColumnSchemaname: database.SQLSchemaPublic, database.SQLColumnTablename: req.Name}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		var count int
		err = tx.QueryRow(ctx, checkSql, checkArgs...).Scan(&count)
		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Collection existence check failed: %v", err)
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}
		if count > 0 {
			return api.SendError(c, api.ErrCollectionAlreadyExists, 400)
		}

		createSQL := database.NewTable(req.Name).
			AddColumn(database.NewColumn(database.SQLRecordID).Varchar(255).PrimaryKey().Build()).
			Build()
		if _, err := tx.Exec(ctx, createSQL); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Base Table Creation failed: %v", err)
			return api.SendError(c, api.ErrCollectionCreateError, 500)
		}

		for _, field := range req.Schema {
			if field.Type == types.TypeID || field.Name == database.SQLRecordID {
				continue
			}
			if err := addColumn(ctx, tx, req.Name, field); err != nil {
				return api.SendError(c, api.ErrCollectionQueryExecution, 400)
			}
		}

		schemaJSON, _ := json.Marshal(req.Schema)
		metaSql, metaArgs, err := database.Query().
			Insert(database.SQLTableCollections).
			Columns(database.SQLTableCollectionsName, database.SQLTableCollectionsSchema).
			Values(req.Name, schemaJSON).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if _, err := tx.Exec(ctx, metaSql, metaArgs...); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Metadata Insert failed: %v", err)
			return api.SendError(c, api.ErrCollectionCreateError, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/COLLECTIONS] Collection created successfully: %s", req.Name)
		return api.Success(c, fiber.Map{"name": req.Name}, 201)
	})

	collection.Patch("/:name", func(c *fiber.Ctx) error {
		oldName := c.Params("name")
		if !api.ValidateName(oldName) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		type UpdateRequest struct {
			Name   string                 `json:"name"`
			Schema types.CollectionSchema `json:"schema"`
		}

		var req UpdateRequest
		if err := c.BodyParser(&req); err != nil {
			return api.SendError(c, api.ErrCoreInvalidBody, 400)
		}

		req.Schema.Sanitize()

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, oldName) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		currentName := oldName
		if req.Name != "" && req.Name != oldName && api.ValidateName(req.Name) {
			renameSQL := database.AlterTable(oldName).RenameTo(req.Name).Build()
			if _, err := tx.Exec(ctx, renameSQL); err != nil {
				logger.Error("[ADMIN/COLLECTIONS] Table Rename failed: %v", err)
				return api.SendError(c, api.ErrCollectionQueryExecution, 500)
			}
			currentName = req.Name
		}

		sql, args, err := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: oldName}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		var currentSchemaJSON []byte
		err = tx.QueryRow(ctx, sql, args...).Scan(&currentSchemaJSON)
		if err != nil {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		var currentSchema types.CollectionSchema
		if err := json.Unmarshal(currentSchemaJSON, &currentSchema); err != nil {
			return api.SendError(c, api.ErrCollectionInvalidSchema, 500)
		}

		if req.Schema != nil {
			for _, field := range req.Schema {
				if !api.ValidateName(field.Name) {
					continue
				}
				var oldField *types.CollectionField
				for _, curField := range currentSchema {
					if curField.ID == field.ID {
						oldField = &curField
						break
					}
				}

				if err := syncField(ctx, tx, currentName, field, oldField); err != nil {
					return api.SendError(c, api.ErrCollectionQueryExecution, 400)
				}
			}

			for _, oldField := range currentSchema {
				removed := true
				for _, newField := range req.Schema {
					if newField.ID == oldField.ID {
						removed = false
						break
					}
				}
				if removed {
					dropSql := database.AlterTable(currentName).DropColumn(oldField.Name).Build()
					if _, err := tx.Exec(ctx, dropSql); err != nil {
						logger.Error("[ADMIN/COLLECTIONS] Schema Update DROP COLUMN failed: %v", err)
						return api.SendError(c, api.ErrCollectionQueryExecution, 500)
					}
				}
			}
		}

		schemaJSON, _ := json.Marshal(req.Schema)
		updateMetaSQL, metaArgs, err := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsName, currentName).
			Set(database.SQLTableCollectionsSchema, schemaJSON).
			Set(database.SQLTableCollectionsUpdatedAt, sq.Expr(database.SQLNow)).
			Where(sq.Eq{database.SQLTableCollectionsName: oldName}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if _, err := tx.Exec(ctx, updateMetaSQL, metaArgs...); err != nil {
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/COLLECTIONS] Collection %s updated successfully", currentName)
		return api.Success(c, fiber.Map{"name": currentName, "schema": req.Schema})
	})

	collection.Delete("/:name/truncate", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		truncateSQL := database.NewStatement(database.SQLTruncateTable).Add(database.Quote(name)).String()
		if _, err := tx.Exec(ctx, truncateSQL); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Truncate Table failed: %v", err)
			return api.SendError(c, api.ErrCollectionTruncateFailed.WithReplacements(map[string]string{"error": err.Error()}), 500)
		}

		resetMetaSQL, args, err := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsTotalDocuments, 0).
			Set(database.SQLTableCollectionsTotalBytes, 0).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if _, err := tx.Exec(ctx, resetMetaSQL, args...); err != nil {
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/COLLECTIONS] Collection truncated: %s", name)
		return api.Success(c, fiber.Map{"name": name})
	})

	collection.Delete("/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		if !api.ValidateName(name) {
			return api.SendError(c, api.ErrCollectionInvalidName, 400)
		}

		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		if api.IsSystemCollection(ctx, db, name) {
			return api.SendError(c, api.ErrCollectionSystemProtected, 403)
		}

		dropSQL := database.NewStatement(database.SQLDropTable).Add(database.Quote(name)).String()
		if _, err := tx.Exec(ctx, dropSQL); err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Drop Table failed: %v", err)
			return api.SendError(c, api.ErrCollectionDropFailed.WithReplacements(map[string]string{"error": err.Error()}), 500)
		}

		deleteMetaSQL, args, _ := database.Query().
			Delete(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if _, err := tx.Exec(ctx, deleteMetaSQL, args...); err != nil {
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[ADMIN/COLLECTIONS] Collection deleted: %s", name)
		return api.Success(c, fiber.Map{"name": name})
	})
}

func registerMetaRoutes(router fiber.Router, db *pgxpool.Pool) {
	collections := router.Group("/collections")

	collections.Get("/types", func(c *fiber.Ctx) error {
		return api.Success(c, types.SupportedTypes)
	})

	collections.Get("/", func(c *fiber.Ctx) error {
		ctx := c.Context()
		tx, err := db.Begin(ctx)
		if err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}
		defer tx.Rollback(ctx)

		sql, args, err := database.Query().Select(
			database.SQLTableCollectionsName,
			database.SQLTableCollectionsSchema,
			database.SQLTableCollectionsIsSystem,
			database.SQLTableCollectionsTotalDocuments,
			database.SQLTableCollectionsTotalBytes,
			database.SQLTableCollectionsCreatedAt,
			database.SQLTableCollectionsUpdatedAt,
		).
			From(database.SQLTableCollections).
			OrderBy(database.NewStatement(database.SQLTableCollectionsName, database.SQLAsc).String()).
			ToSql()

		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}
		defer rows.Close()

		results := make([]types.CollectionDetail, 0)
		for rows.Next() {
			var col types.CollectionDetail
			var schemaJSON []byte
			if err := rows.Scan(
				&col.Name,
				&schemaJSON,
				&col.IsSystem,
				&col.TotalDocuments,
				&col.TotalBytes,
				&col.CreatedAt,
				&col.UpdatedAt,
			); err != nil {
				return api.SendError(c, api.ErrCollectionScanData, 500)
			}

			if err := json.Unmarshal(schemaJSON, &col.Schema); err != nil {
				col.Schema = types.CollectionSchema{}
			}

			results = append(results, col)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.SuccessList(c, results, api.ListMeta{
			Total:  len(results),
			Limit:  len(results),
			Offset: 0,
		})
	})
}

func PerformJoins(ctx context.Context, db *pgxpool.Pool, schema types.CollectionSchema, results []map[string]interface{}, fetchData bool) {
	for _, field := range schema {
		if field.Type != types.TypeRelation || field.Props.Relation == nil {
			continue
		}

		relOpts := field.Props.Relation
		targetCollection := relOpts.Collection
		if targetCollection == "" {
			continue
		}

		allTargetIDs := make(map[string]bool)
		for _, row := range results {
			rawVal := row[field.Name]
			if rawVal == nil {
				continue
			}

			var ids []string
			switch v := rawVal.(type) {
			case []byte:
				json.Unmarshal(v, &ids)
			case []interface{}:
				for _, id := range v {
					if s, ok := id.(string); ok {
						ids = append(ids, s)
					}
				}
			case []string:
				ids = v
			}

			if ids == nil {
				ids = []string{}
			}

			for _, id := range ids {
				allTargetIDs[id] = true
			}
			row[field.Name] = ids
		}

		if !fetchData || len(allTargetIDs) == 0 {
			if !fetchData {
				for _, row := range results {
					ids, _ := row[field.Name].([]string)
					if relOpts.RelationMode == types.RelationModeSingle {
						if len(ids) > 0 {
							row[field.Name] = ids[0]
						} else {
							row[field.Name] = nil
						}
					} else {
						row[field.Name] = ids
					}
				}
			} else {
				for _, row := range results {
					if relOpts.RelationMode == types.RelationModeSingle {
						row[field.Name] = nil
					} else {
						row[field.Name] = []interface{}{}
					}
				}
			}
			continue
		}

		var idsToFetch []string
		for id := range allTargetIDs {
			idsToFetch = append(idsToFetch, id)
		}

		sql, args, err := database.Query().
			Select(database.SQLAsterisk).
			From(database.Quote(targetCollection)).
			Where(sq.Eq{database.SQLRecordID: idsToFetch}).
			ToSql()

		if err != nil {
			logger.Error("[ADMIN/COLLECTIONS] Join query build failed for %s: %v", targetCollection, err)
			continue
		}

		rows, err := db.Query(ctx, sql, args...)
		if err != nil {
			continue
		}

	targetFields := rows.FieldDescriptions()
		targetRecordsMap := make(map[string]map[string]interface{})
		for rows.Next() {
			vals, _ := rows.Values()
			targetRow := make(map[string]interface{})
			for i, f := range targetFields {
				val := vals[i]
				if f.DataTypeOID == database.OIDUUID && val != nil {
					val = api.FormatUUID(val)
				}
				targetRow[string(f.Name)] = val
			}
			if id, ok := targetRow["id"].(string); ok {
				targetRecordsMap[id] = targetRow
			}
		}
		rows.Close()

		for _, row := range results {
			ids, _ := row[field.Name].([]string)
			if relOpts.RelationMode == types.RelationModeSingle {
				if len(ids) > 0 {
					if tRec, ok := targetRecordsMap[ids[0]]; ok {
						row[field.Name] = tRec
					} else {
						row[field.Name] = ids[0]
					}
				} else {
					row[field.Name] = nil
				}
			} else {
				relList := []map[string]interface{}{}
				for _, id := range ids {
					if tRec, ok := targetRecordsMap[id]; ok {
						relList = append(relList, tRec)
					} else {
						relList = append(relList, map[string]interface{}{"id": id})
					}
				}
				row[field.Name] = relList
			}
		}
	}
}
