package database

import (
	"encoding/json"
	"strings"
)

type TableBuilder struct {
	name        string
	columns     []string
	indexes     []string
	constraints []string
}

func NewTable(name string) *TableBuilder {
	return &TableBuilder{name: name}
}

func (t *TableBuilder) AddColumn(col string) *TableBuilder {
	t.columns = append(t.columns, col)
	return t
}

func (t *TableBuilder) AddColumns(cols ...string) *TableBuilder {
	t.columns = append(t.columns, cols...)
	return t
}

func (t *TableBuilder) Index(columns ...string) *IndexBuilder {
	return NewIndex(t.name, columns...)
}

func (t *TableBuilder) AddIndex(idx *IndexBuilder) *TableBuilder {
	t.indexes = append(t.indexes, idx.Build())
	return t
}

func (t *TableBuilder) Constraint(name string) *ConstraintBuilder {
	return NewConstraint(t.name, name)
}

func (t *TableBuilder) AddConstraint(cb *ConstraintBuilder) *TableBuilder {
	t.constraints = append(t.constraints, cb.Build())
	return t
}

func (t *TableBuilder) AddForeignKey(column, refTable, refColumn, onDelete string) *TableBuilder {
	constraintName := NewStatement(PrefixForeignKey, t.name, column).Join(SQLUnderscore)

	fk := NewStatement(SQLConstraint).
		Add(Quote(constraintName)).
		Add(SQLForeignKey).
		Add(NewStatement(Quote(column)).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Add(SQLReferences).
		Add(Quote(refTable)).
		Add(NewStatement(Quote(refColumn)).Wrap(SQLOpenParen, SQLCloseParen).String())

	if onDelete != "" {
		fk.Add(SQLOnDelete).Add(onDelete)
	}
	t.constraints = append(t.constraints, fk.String())
	return t
}

func (t *TableBuilder) Build() string {
	allParts := append([]string{}, t.columns...)
	allParts = append(allParts, t.constraints...)

	sep := NewStatement(SQLComma, SQLNewline, SQLTab).Join("")
	bodyStart := NewStatement(SQLOpenParen, SQLNewline, SQLTab).Join("")
	bodyEnd := NewStatement(SQLNewline, SQLCloseParen).Join("")

	body := NewStatement(strings.Join(allParts, sep)).
		Wrap(bodyStart, bodyEnd).
		String()

	sql := NewStatement(SQLCreateTable).
		Add(Quote(t.name)).
		Add(body).
		String()

	sql += SQLSemicolon

	if len(t.indexes) > 0 {
		sql += SQLNewline + strings.Join(t.indexes, SQLNewline)
	}

	return sql
}

type AlterTableBuilder struct {
	name string
	ops  []string
	pre  []string
}

func AlterTable(name string) *AlterTableBuilder {
	return &AlterTableBuilder{name: name}
}

func (a *AlterTableBuilder) AddColumn(col string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpAddColumn).Add(col).String())
	return a
}

func (a *AlterTableBuilder) SafeAddColumn(colName string, colBuilder *ColumnBuilder, defaultValue interface{}) *AlterTableBuilder {
	originalNullable := colBuilder.nullable
	colBuilder.nullable = true
	a.AddColumn(colBuilder.Build())

	a.SetDefault(colName, defaultValue)

	if !originalNullable {
		a.SetNotNull(colName)
	}
	return a
}

func (a *AlterTableBuilder) DropColumn(name string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpDropColumn).Add(Quote(name)).String())
	return a
}

func (a *AlterTableBuilder) RenameColumn(oldName, newName string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpRenameColumn).Add(Quote(oldName)).Add(SQLTo).Add(Quote(newName)).String())
	return a
}

func (a *AlterTableBuilder) RenameTo(newName string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpRenameTo).Add(Quote(newName)).String())
	return a
}

func (a *AlterTableBuilder) AlterColumn(col, operation string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpAlterColumn).Add(Quote(col)).Add(operation).String())
	return a
}

func (a *AlterTableBuilder) SetNotNull(col string) *AlterTableBuilder {
	return a.AlterColumn(col, OpSetNotNull)
}

func (a *AlterTableBuilder) DropNotNull(col string) *AlterTableBuilder {
	return a.AlterColumn(col, OpDropNotNull)
}

func (a *AlterTableBuilder) DropDefault(col string) *AlterTableBuilder {
	return a.AlterColumn(col, OpDropDefault)
}

func (a *AlterTableBuilder) SetType(col, dataType string) *AlterTableBuilder {
	return a.AlterColumn(col, SQLType+SQLSpace+dataType)
}

func (a *AlterTableBuilder) SetDefault(col string, val interface{}) *AlterTableBuilder {
	var defaultVal string
	if val == nil {
		defaultVal = SQLNull
	} else {
		switch v := val.(type) {
		case string:
			upperV := strings.ToUpper(v)
			if v == SQLNow || upperV == SQLNull || strings.Contains(v, SQLOpenParen) || strings.Contains(v, SQLCast) {
				defaultVal = v
			} else {
				defaultVal = SQLSingleQuote + strings.ReplaceAll(v, SQLSingleQuote, SQLSingleQuote+SQLSingleQuote) + SQLSingleQuote
			}
		case bool:
			if v {
				defaultVal = SQLTrue
			} else {
				defaultVal = SQLFalse
			}
		case map[string]interface{}, []interface{}:
			jsonData, _ := json.Marshal(v)
			defaultVal = SQLSingleQuote + strings.ReplaceAll(string(jsonData), SQLSingleQuote, SQLSingleQuote+SQLSingleQuote) + SQLSingleQuote + SQLCast + TypeJSONB
		default:
			defaultVal = ToVal(v)
		}
	}
	return a.AlterColumn(col, OpSetDefault+SQLSpace+defaultVal)
}

func (a *AlterTableBuilder) Index(columns ...string) *IndexBuilder {
	return NewIndex(a.name, columns...)
}

func (a *AlterTableBuilder) AddIndex(idx *IndexBuilder) *AlterTableBuilder {
	a.pre = append(a.pre, idx.Build())
	return a
}

func (a *AlterTableBuilder) DropIndex(columns []string) *AlterTableBuilder {
	idx := NewIndex(a.name, columns...)
	a.pre = append(a.pre, idx.BuildDrop())
	return a
}

func (a *AlterTableBuilder) Constraint(name string) *ConstraintBuilder {
	return NewConstraint(a.name, name)
}

func (a *AlterTableBuilder) AddConstraint(cb *ConstraintBuilder) *AlterTableBuilder {
	a.ops = append(a.ops, cb.BuildAdd())
	return a
}

func (a *AlterTableBuilder) DropConstraint(name string) *AlterTableBuilder {
	a.ops = append(a.ops, NewStatement(OpDropConstraint).Add(Quote(name)).String())
	return a
}

func (a *AlterTableBuilder) AddForeignKey(column, refTable, refColumn, onDelete string) *AlterTableBuilder {
	constraintName := NewStatement(PrefixForeignKey, a.name, column).Join(SQLUnderscore)
	fk := NewStatement(OpAddConstraint).
		Add(NewStatement(SQLConstraint).Add(Quote(constraintName)).String()).
		Add(SQLForeignKey).
		Add(NewStatement(Quote(column)).Wrap(SQLOpenParen, SQLCloseParen).String()).
		Add(SQLReferences).
		Add(Quote(refTable)).
		Add(NewStatement(Quote(refColumn)).Wrap(SQLOpenParen, SQLCloseParen).String())

	if onDelete != "" {
		fk.Add(SQLOnDelete).Add(onDelete)
	}
	a.ops = append(a.ops, fk.String())
	return a
}

func (a *AlterTableBuilder) DropForeignKey(column string) *AlterTableBuilder {
	constraintName := NewStatement(PrefixForeignKey, a.name, column).Join(SQLUnderscore)
	a.ops = append(a.ops, NewStatement(OpDropConstraint).Add(Quote(constraintName)).String())
	return a
}

func (a *AlterTableBuilder) Build() string {
	stmt := NewStatement()
	if len(a.pre) > 0 {
		stmt.Add(strings.Join(a.pre, SQLNewline) + SQLNewline)
	}

	if len(a.ops) > 0 {
		alterPart := NewStatement(SQLAlterTable).
			Add(Quote(a.name)).
			Add(strings.Join(a.ops, SQLComma+SQLSpace)).
			String()
		stmt.Add(alterPart + SQLSemicolon)
	}
	return stmt.String()
}
