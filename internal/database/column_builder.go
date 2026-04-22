package database

import (
	"encoding/json"
	"strings"
)

type ColumnBuilder struct {
	name       string
	dataType   string
	nullable   bool
	unique     bool
	primaryKey bool
	defaultVal string
	check      []string
	references string
	onDelete   string
}

func NewColumn(name string) *ColumnBuilder {
	return &ColumnBuilder{
		name:     name,
		nullable: true,
	}
}

func (b *ColumnBuilder) PrimaryKey() *ColumnBuilder {
	b.primaryKey = true
	b.nullable = false
	return b
}

func (b *ColumnBuilder) ID() *ColumnBuilder {
	b.dataType = TypeVarchar
	b.unique = true
	b.nullable = false
	return b
}

func (b *ColumnBuilder) Text() *ColumnBuilder {
	b.dataType = TypeText
	return b
}

func (b *ColumnBuilder) Varchar(length int) *ColumnBuilder {
	b.dataType = NewStatement(TypeVarchar).Add(NewStatement().AddInt(length).Wrap(SQLOpenParen, SQLCloseParen).String()).String()
	return b
}

func (b *ColumnBuilder) Integer() *ColumnBuilder {
	b.dataType = TypeInteger
	return b
}

func (b *ColumnBuilder) BigInt() *ColumnBuilder {
	b.dataType = TypeBigInt
	return b
}

func (b *ColumnBuilder) Numeric(p, s *int) *ColumnBuilder {
	if p != nil && s != nil {
		precisionStr := NewStatement().AddInt(*p).Add(SQLComma).AddInt(*s).String()
		b.dataType = NewStatement(TypeNumeric).Add(NewStatement(precisionStr).Wrap(SQLOpenParen, SQLCloseParen).String()).String()
	} else if p != nil {
		b.dataType = NewStatement(TypeNumeric).Add(NewStatement().AddInt(*p).Wrap(SQLOpenParen, SQLCloseParen).String()).String()
	} else {
		b.dataType = TypeNumeric
	}
	return b
}

func (b *ColumnBuilder) Boolean(defaultVal bool) *ColumnBuilder {
	b.dataType = TypeBoolean
	if defaultVal {
		b.defaultVal = SQLTrue
	} else {
		b.defaultVal = SQLFalse
	}
	return b
}

func (b *ColumnBuilder) JSONB() *ColumnBuilder {
	b.dataType = TypeJSONB
	return b
}

func (b *ColumnBuilder) Timestamp() *ColumnBuilder {
	b.dataType = TypeTimestamp
	return b
}

func (b *ColumnBuilder) UUID() *ColumnBuilder {
	b.dataType = TypeUUID
	return b
}

func (b *ColumnBuilder) NotNull() *ColumnBuilder {
	b.nullable = false
	return b
}

func (b *ColumnBuilder) Unique() *ColumnBuilder {
	b.unique = true
	return b
}

func (b *ColumnBuilder) DefaultNow() *ColumnBuilder {
	b.defaultVal = SQLNow
	return b
}

func (b *ColumnBuilder) DefaultNull() *ColumnBuilder {
	b.defaultVal = SQLNull
	return b
}

func (b *ColumnBuilder) Default(val interface{}) *ColumnBuilder {
	if val == nil {
		return b
	}
	var finalVal string
	switch v := val.(type) {
	case string:
		upperV := strings.ToUpper(v)
		if v == SQLNow || upperV == SQLNull || strings.Contains(v, SQLOpenParen) || strings.Contains(v, SQLCast) {
			finalVal = v
		} else {
			finalVal = SQLSingleQuote + strings.ReplaceAll(v, SQLSingleQuote, SQLSingleQuote+SQLSingleQuote) + SQLSingleQuote
		}
	case bool:
		if v {
			finalVal = SQLTrue
		} else {
			finalVal = SQLFalse
		}
	case map[string]interface{}, []interface{}, []string:
		jsonData, _ := json.Marshal(v)
		finalVal = SQLSingleQuote + strings.ReplaceAll(string(jsonData), SQLSingleQuote, SQLSingleQuote+SQLSingleQuote) + SQLSingleQuote + SQLCast + TypeJSONB
	default:
		finalVal = ToVal(v)
	}
	b.defaultVal = finalVal
	return b
}

func (b *ColumnBuilder) SetType(dataType string) *ColumnBuilder {
	b.dataType = dataType
	return b
}

type SQLExpr struct {
	expr string
}

var SQL = struct {
	Col        func(string) *SQLExpr
	Trim       func(string) *SQLExpr
	CharLength func(*SQLExpr) *SQLExpr
}{
	Col: func(col string) *SQLExpr {
		return &SQLExpr{expr: Quote(col)}
	},
	Trim: func(col string) *SQLExpr {
		return &SQLExpr{expr: NewStatement(FuncTrim).Add(NewStatement(Quote(col)).Wrap(SQLOpenParen, SQLCloseParen).String()).String()}
	},
	CharLength: func(ex *SQLExpr) *SQLExpr {
		return &SQLExpr{expr: NewStatement(FuncCharLength).Add(NewStatement(ex.expr).Wrap(SQLOpenParen, SQLCloseParen).String()).String()}
	},
}

func (e *SQLExpr) Gte(val interface{}) string {
	return NewStatement(e.expr).Add(SQLGte).Add(ToVal(val)).String()
}
func (e *SQLExpr) Lte(val interface{}) string {
	return NewStatement(e.expr).Add(SQLLte).Add(ToVal(val)).String()
}

func (b *ColumnBuilder) Min(val interface{}) *ColumnBuilder {
	return b.AddCheck(SQL.Col(b.name).Gte(val))
}

func (b *ColumnBuilder) Max(val interface{}) *ColumnBuilder {
	return b.AddCheck(SQL.Col(b.name).Lte(val))
}

func (b *ColumnBuilder) MinLength(n int) *ColumnBuilder {
	return b.AddCheck(SQL.CharLength(SQL.Trim(b.name)).Gte(n))
}

func (b *ColumnBuilder) MaxLength(n int) *ColumnBuilder {
	return b.AddCheck(SQL.CharLength(SQL.Trim(b.name)).Lte(n))
}

func (b *ColumnBuilder) AddCheck(condition string) *ColumnBuilder {
	b.check = append(b.check, condition)
	return b
}

func (b *ColumnBuilder) References(table, column string) *ColumnBuilder {
	b.references = NewStatement(Quote(table)).
		Add(NewStatement(Quote(column)).Wrap(SQLOpenParen, SQLCloseParen).String()).
		String()
	return b
}

func (b *ColumnBuilder) OnDeleteCascade() *ColumnBuilder {
	b.onDelete = SQLCascade
	return b
}

func (b *ColumnBuilder) OnDeleteSetNull() *ColumnBuilder {
	b.onDelete = SQLSetNull
	return b
}

func (b *ColumnBuilder) Build() string {
	stmt := NewStatement().
		Add(Quote(b.name)).
		Add(b.dataType)

	if b.primaryKey {
		stmt.Add(SQLPrimaryKey)
	}

	if !b.nullable {
		stmt.Add(SQLNotNull)
	} else if !b.primaryKey {
		stmt.Add(SQLNull)
	}

	if b.unique && !b.primaryKey {
		stmt.Add(SQLUnique)
	}

	if b.defaultVal != "" {
		stmt.Add(SQLDefault).Add(b.defaultVal)
	}

	if len(b.check) > 0 {
		checkExpr := NewStatement(NewStatement(b.check...).JoinAnd()).
			Wrap(SQLOpenParen, SQLCloseParen).
			String()
		stmt.Add(SQLCheck).Add(checkExpr)
	}

	if b.references != "" {
		refStmt := NewStatement(SQLReferences).Add(b.references)
		if b.onDelete != "" {
			refStmt.Add(SQLOnDelete).Add(b.onDelete)
		}
		stmt.Add(refStmt.String())
	}

	return stmt.String()
}
