package database

import (
	"fmt"
	"strconv"
	"strings"
)

type Statement struct {
	parts []string
}

func (s *Statement) ToSql() (any, error) {
	panic("unimplemented")
}

func NewStatement(parts ...string) *Statement {
	cleanParts := []string{}
	for _, p := range parts {
		if p != "" {
			cleanParts = append(cleanParts, p)
		}
	}
	return &Statement{parts: cleanParts}
}

func (s *Statement) Add(part string) *Statement {
	if part != "" {
		s.parts = append(s.parts, part)
	}
	return s
}

func (s *Statement) AddInt(val int) *Statement {
	return s.Add(strconv.Itoa(val))
}

func (s *Statement) Space() *Statement {
	return s
}

func (s *Statement) Newline() *Statement {
	return s.Add(SQLNewline)
}

func (s *Statement) Tab() *Statement {
	return s.Add(SQLTab)
}

func (s *Statement) Join(sep string) string {
	return strings.Join(s.parts, sep)
}

func (s *Statement) JoinAnd() string {
	return strings.Join(s.parts, SQLSpace+SQLAnd+SQLSpace)
}

func (s *Statement) String() string {
	return strings.Join(s.parts, SQLSpace)
}

func Quote(name string) string {

	escaped := strings.ReplaceAll(name, SQLQuote, SQLQuote+SQLQuote)
	return SQLQuote + escaped + SQLQuote
}

func Coalesce(col, fallback string) string {

	return NewStatement(FuncCoalesce).Add(SQLOpenParen + col + SQLComma + SQLSpace + fallback + SQLCloseParen).String()
}

func Cast(val, targetType string) string {
	return val + SQLCast + targetType
}

func (s *Statement) Wrap(left, right string) *Statement {
	if len(s.parts) > 0 {
		last := len(s.parts) - 1
		s.parts[last] = left + s.parts[last] + right
	}
	return s
}

func ToVal(v interface{}) string {
	if v == nil {
		return SQLNull
	}

	if s, ok := v.(string); ok {
		escaped := strings.ReplaceAll(s, SQLSingleQuote, SQLSingleQuote+SQLSingleQuote)
		return SQLSingleQuote + escaped + SQLSingleQuote
	}
	return fmt.Sprintf("%v", v)
}
