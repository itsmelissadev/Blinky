package collections

import (
	"blinky/internal/database"
	"blinky/internal/pkg/logger"
	"blinky/internal/types"
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func syncField(ctx context.Context, tx pgx.Tx, tableName string, field types.CollectionField, oldField *types.CollectionField) error {
	logger.Info("[DATABASE/SCHEMA] Synchronizing field %s in %s", field.Name, tableName)
	switch field.Type {
	case types.TypeText:
		return handleText(ctx, tx, tableName, field, oldField)
	case types.TypeNumber:
		return handleNumber(ctx, tx, tableName, field, oldField)
	case types.TypeBoolean:
		return handleBoolean(ctx, tx, tableName, field, oldField)
	case types.TypeJSON:
		return handleJSON(ctx, tx, tableName, field, oldField)
	case types.TypeDate:
		return handleDate(ctx, tx, tableName, field, oldField)
	case types.TypeID:
		return handleID(ctx, tx, tableName, field, oldField)
	case types.TypeRelation:
		return handleRelation(ctx, tx, tableName, field, oldField)
	}
	return nil
}

func handleText(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	props := field.Props
	targetType := database.TypeText

	if def := field.GetDefault(); def != nil {
		if s, ok := def.(string); ok {
			if props.Text != nil {
				if props.Text.Min != nil && len(s) < *props.Text.Min {
					return fmt.Errorf("field '%s': default value length is less than min length (%d)", field.Name, *props.Text.Min)
				}
				if props.Text.Max != nil && len(s) > *props.Text.Max {
					return fmt.Errorf("field '%s': default value length is greater than max length (%d)", field.Name, *props.Text.Max)
				}
			}
		}
	}

	typeChanged, err := syncLifecycle(ctx, tx, table, field, oldField, targetType)
	if err != nil {
		return err
	}

	opts := props.Text
	if opts == nil {
		opts = &types.TextOptions{}
	}
	var oldOpts types.TextOptions
	if oldField != nil && oldField.Props.Text != nil {
		oldOpts = *oldField.Props.Text
	}

	if opts.Min != nil && opts.Max != nil && *opts.Min > *opts.Max {
		return fmt.Errorf("field '%s': min length cannot be greater than max length", field.Name)
	}

	if typeChanged || isIntPtrChanged(opts.Min, oldOpts.Min) {
		if err := syncCheck(ctx, tx, table, field.Name, "min", func(b *database.ConstraintBuilder) {
			if opts.Min != nil {
				b.MinLength(field.Name, *opts.Min)
			}
		}); err != nil {
			return err
		}
	}
	if typeChanged || isIntPtrChanged(opts.Max, oldOpts.Max) {
		if err := syncCheck(ctx, tx, table, field.Name, "max", func(b *database.ConstraintBuilder) {
			if opts.Max != nil {
				b.MaxLength(field.Name, *opts.Max)
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func handleNumber(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	props := field.Props
	opts := props.Number
	if opts == nil {
		opts = &types.NumberOptions{}
	}

	targetType := database.TypeNumeric
	if opts.NoDecimals {
		targetType = database.TypeBigInt
	}

	if def := field.GetDefault(); def != nil {
		var val float64
		switch v := def.(type) {
		case float64:
			val = v
		case int:
			val = float64(v)
		case int64:
			val = float64(v)
		case string:
			if v != "" {
				parsed, err := strconv.ParseFloat(v, 64)
				if err == nil {
					val = parsed
				}
			}
		}

		if opts.NoZero && val == 0 {
			return fmt.Errorf("field '%s': zero value is not allowed", field.Name)
		}

		if opts.Min != nil && val < *opts.Min {
			return fmt.Errorf("field '%s': default value %v is less than min %v", field.Name, val, *opts.Min)
		}
		if opts.Max != nil && val > *opts.Max {
			return fmt.Errorf("field '%s': default value %v is greater than max %v", field.Name, val, *opts.Max)
		}
	}

	typeChanged, err := syncLifecycle(ctx, tx, table, field, oldField, targetType)
	if err != nil {
		return err
	}

	var oldOpts types.NumberOptions
	if oldField != nil && oldField.Props.Number != nil {
		oldOpts = *oldField.Props.Number
	}

	if opts.Min != nil && opts.Max != nil && *opts.Min > *opts.Max {
		return fmt.Errorf("field '%s': min value cannot be greater than max value", field.Name)
	}

	if typeChanged || isFloatPtrChanged(opts.Min, oldOpts.Min) {
		if err := syncCheck(ctx, tx, table, field.Name, "min", func(b *database.ConstraintBuilder) {
			if opts.Min != nil {
				b.MinValue(field.Name, *opts.Min)
			}
		}); err != nil {
			return err
		}
	}
	if typeChanged || isFloatPtrChanged(opts.Max, oldOpts.Max) {
		if err := syncCheck(ctx, tx, table, field.Name, "max", func(b *database.ConstraintBuilder) {
			if opts.Max != nil {
				b.MaxValue(field.Name, *opts.Max)
			}
		}); err != nil {
			return err
		}
	}

	if typeChanged || opts.NoZero != oldOpts.NoZero {
		if err := syncCheck(ctx, tx, table, field.Name, "nozero", func(b *database.ConstraintBuilder) {
			if opts.NoZero {
				b.Check(database.NewStatement(database.Quote(field.Name)).
					Add(database.SQLNotEq).Add(database.SQLZero).String())
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func handleBoolean(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	f := false
	field.Props.IsNullable = &f
	field.Props.IsUnique = &f

	_, err := syncLifecycle(ctx, tx, table, field, oldField, database.TypeBoolean)
	return err
}

func handleJSON(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	_, err := syncLifecycle(ctx, tx, table, field, oldField, database.TypeJSONB)
	return err
}

func handleDate(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	_, err := syncLifecycle(ctx, tx, table, field, oldField, database.TypeTimestamp)
	return err
}

func handleRelation(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	_, err := syncLifecycle(ctx, tx, table, field, oldField, database.TypeJSONB)
	return err
}

func handleID(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField) error {
	if oldField == nil {
		colBuilder := database.NewColumn(field.Name).ID().PrimaryKey()
		alterSQL := database.AlterTable(table).AddColumn(colBuilder.Build()).Build()
		_, err := tx.Exec(ctx, alterSQL)
		return err
	}
	return nil
}

func syncLifecycle(ctx context.Context, tx pgx.Tx, table string, field types.CollectionField, oldField *types.CollectionField, targetType string) (bool, error) {
	props := field.Props

	if oldField == nil {
		colBuilder := database.NewColumn(field.Name).SetType(targetType)

		if props.IsNullable != nil && !*props.IsNullable {
			colBuilder.NotNull()
		}
		if def := field.GetDefault(); def != nil {
			colBuilder.Default(def)
		}

		alterSQL := database.AlterTable(table).AddColumn(colBuilder.Build()).Build()
		logger.Debug("[DATABASE/SCHEMA] Adding column %s to %s", field.Name, table)
		if _, err := tx.Exec(ctx, alterSQL); err != nil {
			return false, err
		}

		if props.IsUnique != nil && *props.IsUnique {
			idxSQL := database.AlterTable(table).Index(field.Name).Unique().Build()
			logger.Debug("[DATABASE/SCHEMA] Adding unique index for %s on %s", field.Name, table)
			if _, err := tx.Exec(ctx, idxSQL); err != nil {
				return false, err
			}
		}
		return true, nil
	}

	currentName := oldField.Name
	if field.Name != oldField.Name {
		renameSQL := database.AlterTable(table).RenameColumn(oldField.Name, field.Name).Build()
		if _, err := tx.Exec(ctx, renameSQL); err != nil {
			return false, err
		}
		currentName = field.Name
	}

	typeChanged := field.Type != oldField.Type || isNumberOptionsChanged(props.Number, oldField.Props.Number)

	newNull := props.IsNullable == nil || *props.IsNullable
	oldNull := oldField.Props.IsNullable == nil || *oldField.Props.IsNullable
	if oldNull != newNull {
		alter := database.AlterTable(table)
		if newNull {
			alter.DropNotNull(currentName)
		} else {
			alter.SetNotNull(currentName)
		}
		if _, err := tx.Exec(ctx, alter.Build()); err != nil {
			return typeChanged, err
		}
	}

	if typeChanged {
		using := getUsingClause(currentName, targetType)
		alterSQL := database.AlterTable(table).AlterColumn(currentName, database.NewStatement(database.SQLType).
			Add(targetType).Add(database.SQLUsing).Add(using).String()).Build()
		if _, err := tx.Exec(ctx, alterSQL); err != nil {
			return typeChanged, err
		}
	}

	newUnique := props.IsUnique != nil && *props.IsUnique
	oldUnique := oldField.Props.IsUnique != nil && *oldField.Props.IsUnique
	if oldUnique != newUnique {
		alter := database.AlterTable(table)
		if newUnique {
			alter.AddIndex(alter.Index(currentName).Unique())
		} else {
			alter.DropIndex([]string{currentName})
		}
		if _, err := tx.Exec(ctx, alter.Build()); err != nil {
			return typeChanged, err
		}
	}

	newDef := field.GetDefault()
	oldDef := oldField.GetDefault()
	if !reflect.DeepEqual(newDef, oldDef) {
		alter := database.AlterTable(table)
		if newDef == nil {
			alter.DropDefault(currentName)
		} else {
			alter.SetDefault(currentName, newDef)
		}
		if _, err := tx.Exec(ctx, alter.Build()); err != nil {
			return typeChanged, err
		}
	}

	return typeChanged, nil
}

func syncCheck(ctx context.Context, tx pgx.Tx, table, col, suffix string, builder func(*database.ConstraintBuilder)) error {
	alter := database.AlterTable(table)
	cName := database.NewStatement(database.PrefixCheck, table, col, suffix).Join(database.SQLUnderscore)

	alter.DropConstraint(cName)

	cb := alter.Constraint(cName)
	builder(cb)

	if cb.HasBody() {
		alter.AddConstraint(cb)
	}

	sql := alter.Build()
	if sql != "" {
		if _, err := tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("sync check %s failed: %w", cName, err)
		}
	}
	return nil
}

func getUsingClause(colName string, targetType string) string {
	quoted := database.Quote(colName)
	if targetType == database.TypeJSONB {
		return database.NewStatement(database.SQLToJsonb).Add(database.SQLOpenParen).Add(quoted).Add(database.SQLCloseParen).String()
	}
	return database.Cast(quoted, targetType)
}

func isIntPtrChanged(n, o *int) bool {
	if n == nil && o == nil {
		return false
	}
	if n == nil || o == nil {
		return true
	}
	return *n != *o
}

func isFloatPtrChanged(n, o *float64) bool {
	if n == nil && o == nil {
		return false
	}
	if n == nil || o == nil {
		return true
	}
	return *n != *o
}

func isNumberOptionsChanged(n, o *types.NumberOptions) bool {
	if n == nil && o == nil {
		return false
	}
	if n == nil || o == nil {
		return true
	}
	return isFloatPtrChanged(n.Min, o.Min) ||
		isFloatPtrChanged(n.Max, o.Max) ||
		n.NoZero != o.NoZero ||
		n.NoDecimals != o.NoDecimals
}

func addColumn(ctx context.Context, tx pgx.Tx, tableName string, field types.CollectionField) error {
	return syncField(ctx, tx, tableName, field, nil)
}
