package database

import (
	"strings"
)

type IndexBuilder struct {
	tableName    string
	name         string
	columns      []string
	isUnique     bool
	isConcurrent bool
}

func NewIndex(tableName string, columns ...string) *IndexBuilder {
	return &IndexBuilder{
		tableName: tableName,
		columns:   columns,
	}
}

func (b *IndexBuilder) Unique() *IndexBuilder {
	b.isUnique = true
	return b
}

func (b *IndexBuilder) Concurrently() *IndexBuilder {
	b.isConcurrent = true
	return b
}

func (b *IndexBuilder) Name(name string) *IndexBuilder {
	b.name = name
	return b
}

func (b *IndexBuilder) Build() string {
	cmd := SQLCreateIndex
	if b.isUnique {
		cmd = SQLCreateUniqueIndex
	}

	idxName := b.name
	if idxName == "" {
		idxName = NewStatement(PrefixIndex, b.tableName, strings.Join(b.columns, SQLUnderscore)).Join(SQLUnderscore)
	}

	var quotedCols []string
	for _, col := range b.columns {
		quotedCols = append(quotedCols, Quote(col))
	}
	cols := NewStatement(strings.Join(quotedCols, SQLComma+SQLSpace)).
		Wrap(SQLOpenParen, SQLCloseParen).
		String()

	stmt := NewStatement(cmd)

	if b.isConcurrent {
		stmt.Add("CONCURRENTLY")
	}

	return stmt.
		Add(Quote(idxName)).
		Add(SQLOn).
		Add(Quote(b.tableName)).
		Add(cols).
		String() + SQLSemicolon
}

func (b *IndexBuilder) BuildDrop() string {
	idxName := b.name
	if idxName == "" {
		idxName = NewStatement(PrefixIndex, b.tableName, strings.Join(b.columns, SQLUnderscore)).Join(SQLUnderscore)
	}

	stmt := NewStatement(SQLDropIndex)

	if b.isConcurrent {
		stmt.Add("CONCURRENTLY")
	}

	return stmt.
		Add(Quote(idxName)).
		String() + SQLSemicolon
}
