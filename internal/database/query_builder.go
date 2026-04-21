package database

import (
	"github.com/Masterminds/squirrel"
)

var PBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type NewQuery struct{}

func Query() *NewQuery {
	return &NewQuery{}
}

func (q *NewQuery) Select(columns ...string) squirrel.SelectBuilder {
	return PBuilder.Select(columns...)
}

func (q *NewQuery) Insert(table string) squirrel.InsertBuilder {
	return PBuilder.Insert(table)
}

func (q *NewQuery) Update(table string) squirrel.UpdateBuilder {
	return PBuilder.Update(table)
}

func (q *NewQuery) Delete(table string) squirrel.DeleteBuilder {
	return PBuilder.Delete(table)
}

func (q *NewQuery) And(conditions ...squirrel.Sqlizer) squirrel.And {
	return squirrel.And(conditions)
}

func (q *NewQuery) Or(conditions ...squirrel.Sqlizer) squirrel.Or {
	return squirrel.Or(conditions)
}
