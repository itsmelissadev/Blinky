package database

import (
	"strings"
)

type IndexBuilder struct {
	tableName string
	name      string
	columns   []string
	isUnique  bool
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

	cols := NewStatement(strings.Join(b.columns, SQLQuote+SQLComma+SQLSpace+SQLQuote)).
		Wrap(SQLQuote, SQLQuote).
		Wrap(SQLOpenParen, SQLCloseParen).
		String()

	return NewStatement(cmd).
		Space().
		Add(SQLQuote+idxName+SQLQuote).
		Space().
		Add(SQLOn).
		Space().
		Add(SQLQuote+b.tableName+SQLQuote).
		Space().
		Add(cols).
		String() + SQLSemicolon
}

func (b *IndexBuilder) BuildDrop() string {
	idxName := b.name
	if idxName == "" {
		idxName = NewStatement(PrefixIndex, b.tableName, strings.Join(b.columns, SQLUnderscore)).Join(SQLUnderscore)
	}
	return NewStatement(SQLDropIndex).
		Space().
		Add(SQLQuote+idxName+SQLQuote).
		String() + SQLSemicolon
}
