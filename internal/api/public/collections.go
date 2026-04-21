package public

import (
	"encoding/json"
	"strings"
	"time"

	"blinky/internal/api"
	"blinky/internal/api/admin/collections"
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/types"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerCollectionRoutes(router fiber.Router, db *pgxpool.Pool) {
	group := router.Group("/collections")

	group.Get("/:name", func(c *fiber.Ctx) error {
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
			logger.Warn("[PUBLIC/COLLECTIONS] Failed to fetch metadata for %s: %v", name, err)
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		schema := col.Schema
		totalDocuments := int(col.TotalDocuments)

		validFields := make(map[string]bool)
		validFields[database.SQLRecordID] = true
		for _, f := range schema {
			validFields[strings.ToLower(f.Name)] = true
		}

		queryBuilder := database.Query().
			Select(database.SQLAsterisk).
			From(database.Quote(name))

		queries := c.Queries()
		for key, value := range queries {
			if key == "limit" || key == "offset" || key == "sort" {
				continue
			}

			parts := strings.Split(key, ":")
			column := parts[0]
			columnLower := strings.ToLower(column)

			if !validFields[columnLower] {
				return api.SendError(c, api.ErrCollectionColumnNotFound.WithReplacements(map[string]string{
					"field":      column,
					"collection": name,
				}), 400)
			}

			operator := "eq"
			if len(parts) > 1 {
				operator = strings.ToLower(parts[1])
			}

			quotedColumn := database.Quote(column)

			switch operator {
			case "gt":
				queryBuilder = queryBuilder.Where(sq.Gt{quotedColumn: value})
			case "lt":
				queryBuilder = queryBuilder.Where(sq.Lt{quotedColumn: value})
			case "gte":
				queryBuilder = queryBuilder.Where(sq.GtOrEq{quotedColumn: value})
			case "lte":
				queryBuilder = queryBuilder.Where(sq.LtOrEq{quotedColumn: value})
			case "neq":
				queryBuilder = queryBuilder.Where(sq.NotEq{quotedColumn: value})
			case "like":
				queryBuilder = queryBuilder.Where(sq.Like{quotedColumn: value})
			case "in":
				vals := strings.Split(value, ",")
				queryBuilder = queryBuilder.Where(sq.Eq{quotedColumn: vals})
			case "eq":
				queryBuilder = queryBuilder.Where(sq.Eq{quotedColumn: value})
			default:
				return api.SendError(c, api.ErrCollectionInvalidOperator.WithReplacements(map[string]string{
					"operator": operator,
					"field":    column,
				}), 400)
			}
		}

		sort := c.Query("sort")
		if sort != "" {
			order := database.SQLAsc
			actualSort := sort
			if strings.HasPrefix(sort, "-") {
				order = database.SQLDesc
				actualSort = sort[1:]
			}
			if !validFields[strings.ToLower(actualSort)] {
				return api.SendError(c, api.ErrCollectionColumnNotFound.WithReplacements(map[string]string{
					"field":      actualSort,
					"collection": name,
				}), 400)
			}
			queryBuilder = queryBuilder.OrderBy(database.NewStatement(database.Quote(actualSort), order).String())
		}

		limit := c.QueryInt("limit", 100)
		if limit > 1000 {
			limit = 1000
		}
		offset := c.QueryInt("offset", 0)

		queryBuilder = queryBuilder.Limit(uint64(limit)).Offset(uint64(offset))

		sql, args, err := queryBuilder.ToSql()
		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Public list query build failed for %s: %v", name, err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Public records query failed: %v", err)
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
				if field.DataTypeOID == database.OIDUUID && val != nil {
					val = api.FormatUUID(val)
				}
				row[string(field.Name)] = val
			}
			results = append(results, row)
		}

		collections.PerformJoins(ctx, db, schema, results, true)

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.SuccessList(c, results, api.ListMeta{
			Total:  totalDocuments,
			Limit:  limit,
			Offset: offset,
		})
	})

	group.Post("/:name", func(c *fiber.Ctx) error {
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
		metaSQL, args, _ := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

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
			if (f.Type == types.TypeID || nameLower == database.SQLRecordID) && f.Props.ID != nil && f.Props.ID.AutoGenerateLength > 0 {
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

			if f.Type == types.TypeRelation {
				if val == nil {
					val = []string{}
				} else if s, ok := val.(string); ok {
					val = []string{s}
				}
			}

			if ok {
				insertBuilder = insertBuilder.Columns(database.Quote(f.Name))
				vals = append(vals, val)
			}
		}

		insertSQL, args, err := insertBuilder.Values(vals...).ToSql()
		if err != nil {
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		if _, err := tx.Exec(ctx, insertSQL, args...); err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Public record insertion failed: %v", err)
			dbErr := api.HandleDBError(err)
			return api.SendError(c, dbErr, 400)
		}

		updateMetaSQL, metaArgs, _ := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsTotalDocuments, sq.Expr(database.NewStatement(database.SQLTableCollectionsTotalDocuments).Add(database.SQLPlus).Add(database.SQLOne).String())).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if _, err := tx.Exec(ctx, updateMetaSQL, metaArgs...); err != nil {
			logger.Warn("[PUBLIC/COLLECTIONS] Failed to increment document count: %v", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[PUBLIC/COLLECTIONS] Public record created in %s: %s", name, id)
		return api.Success(c, fiber.Map{"id": id}, 201)
	})

	group.Get("/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name")
		id := c.Params("id")
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

		var schemaData []byte
		metaSQL, metaArgs, _ := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if err := tx.QueryRow(ctx, metaSQL, metaArgs...).Scan(&schemaData); err != nil {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		var schema []types.CollectionField
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return api.SendError(c, api.ErrCollectionInvalidSchema, 500)
		}

		sql, args, err := database.Query().
			Select(database.SQLAsterisk).
			From(database.Quote(name)).
			Where(sq.Eq{database.SQLRecordID: id}).
			Limit(1).
			ToSql()

		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Record fetch query build failed for %s: %v", name, err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}
		defer rows.Close()

		if !rows.Next() {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		fields := rows.FieldDescriptions()
		values, _ := rows.Values()
		row := make(map[string]interface{})
		for i, field := range fields {
			val := values[i]
			if field.DataTypeOID == database.OIDUUID && val != nil {
				val = api.FormatUUID(val)
			}
			row[string(field.Name)] = val
		}

		results := []map[string]interface{}{row}
		collections.PerformJoins(ctx, db, schema, results, true)

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		return api.Success(c, results[0])
	})

	group.Patch("/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name")
		id := c.Params("id")
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
		metaSQL, args, _ := database.Query().
			Select(database.SQLTableCollectionsSchema).
			From(database.SQLTableCollections).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

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
			if kLower == "id" {
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
			logger.Error("[PUBLIC/COLLECTIONS] Record update query build failed: %v", err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		res, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Public record update failed: %v", err)
			dbErr := api.HandleDBError(err)
			return api.SendError(c, dbErr, 400)
		}

		if res.RowsAffected() == 0 {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[PUBLIC/COLLECTIONS] Public record %s updated in %s", id, name)
		return api.Success(c, api.SuccessRecordUpdated)
	})

	group.Delete("/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name")
		id := c.Params("id")
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

		sql, args, err := database.Query().
			Delete(database.Quote(name)).
			Where(sq.Eq{database.SQLRecordID: id}).
			ToSql()

		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Record deletion query build failed: %v", err)
			return api.SendError(c, api.ErrCollectionBuildQuery, 500)
		}

		res, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			logger.Error("[PUBLIC/COLLECTIONS] Public record deletion failed: %v", err)
			return api.SendError(c, api.ErrCollectionQueryExecution, 500)
		}

		if res.RowsAffected() == 0 {
			return api.SendError(c, api.ErrCollectionNotFound, 404)
		}

		updateMetaSQL, metaArgs, _ := database.Query().
			Update(database.SQLTableCollections).
			Set(database.SQLTableCollectionsTotalDocuments, sq.Expr(database.NewStatement(database.SQLTableCollectionsTotalDocuments).Add(database.SQLMinus).Add(database.SQLOne).String())).
			Where(sq.Eq{database.SQLTableCollectionsName: name}).
			ToSql()

		if _, err := tx.Exec(ctx, updateMetaSQL, metaArgs...); err != nil {
			logger.Warn("[PUBLIC/COLLECTIONS] Failed to decrement document count: %v", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return api.SendError(c, api.ErrCoreInternalServer, 500)
		}

		logger.Success("[PUBLIC/COLLECTIONS] Public record %s deleted from %s", id, name)
		return api.Success(c, api.SuccessRecordDeleted)
	})
}
