package types

import (
	"time"
	"blinky/internal/database"
	"github.com/google/uuid"
)

func GenerateFieldID(name string) string {
	ns := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	return uuid.NewSHA1(ns, []byte(name)).String()
}

type FieldType string

const (
	TypeID       FieldType = "id"
	TypeText     FieldType = "text"
	TypeNumber   FieldType = "number"
	TypeBoolean  FieldType = "boolean"
	TypeJSON     FieldType = "json"
	TypeDate     FieldType = "date"
	TypeRelation FieldType = "relation"
)

const (
	RelationModeSingle   = "single"
	RelationModeMultiple = "multiple"
)

const (
	AutoTimestampCreate       = "create"
	AutoTimestampCreateUpdate = "create_update"
)

type FieldTypeInfo struct {
	Type  FieldType `json:"type"`
	Label string    `json:"label"`
}

var SupportedTypes = []FieldTypeInfo{
	{Type: TypeID, Label: "ID"},
	{Type: TypeText, Label: "Text"},
	{Type: TypeNumber, Label: "Number"},
	{Type: TypeBoolean, Label: "Boolean"},
	{Type: TypeJSON, Label: "JSON"},
	{Type: TypeDate, Label: "Date"},
	{Type: TypeRelation, Label: "Relation"},
}

type CollectionField struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Type  FieldType  `json:"type"`
	Props FieldProps `json:"props"`
}

type FieldProps struct {
	ID            *IDConfig        `json:"id,omitempty"`
	IsNullable    *bool            `json:"is_nullable,omitempty"`
	IsUnique      *bool            `json:"is_unique,omitempty"`
	IsHidden      *bool            `json:"is_hidden,omitempty"`
	DefaultVal    interface{}      `json:"default,omitempty"`
	DefaultBool   *bool            `json:"default_bool,omitempty"`
	DefaultNow    bool             `json:"default_now,omitempty"`
	AutoTimestamp string           `json:"auto_timestamp,omitempty"`
	Text          *TextOptions     `json:"text,omitempty"`
	Number        *NumberOptions   `json:"number,omitempty"`
	Relation      *RelationOptions `json:"relation,omitempty"`
}

type IDConfig struct {
	AutoGenerateLength int `json:"auto_len,omitempty"`
}

type TextOptions struct {
	Min   *int   `json:"min,omitempty"`
	Max   *int   `json:"max,omitempty"`
	Regex string `json:"regex,omitempty"`
}

type NumberOptions struct {
	Min        *float64 `json:"min,omitempty"`
	Max        *float64 `json:"max,omitempty"`
	NoZero     bool     `json:"no_zero,omitempty"`
	NoDecimals bool     `json:"no_decimals,omitempty"`
}

type RelationOptions struct {
	Collection    string `json:"collection"`
	RelationMode  string `json:"relation_mode"`
	CascadeDelete bool   `json:"cascade_delete"`
}

type CollectionSchema []CollectionField

type CollectionDetail struct {
	Name           string           `json:"name"`
	Schema         CollectionSchema `json:"schema"`
	IsSystem       bool             `json:"is_system"`
	TotalDocuments int64            `json:"total_documents"`
	TotalBytes     int64            `json:"total_bytes"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

func (f CollectionField) GetDefault() interface{} {
	if f.Type == TypeBoolean && f.Props.DefaultBool != nil {
		return *f.Props.DefaultBool
	}
	if f.Type == TypeDate {
		if f.Props.AutoTimestamp == "create" || f.Props.AutoTimestamp == "create_update" {
			return database.SQLNow
		}
		if f.Props.AutoTimestamp == "none" || f.Props.AutoTimestamp == "" {
			if f.Props.IsNullable != nil && *f.Props.IsNullable {
				return database.SQLNull
			}
		}
	}
	if f.Type == TypeDate && f.Props.DefaultNow {
		return database.SQLNow
	}
	if f.Type == TypeText {
		isUnique := f.Props.IsUnique != nil && *f.Props.IsUnique
		isNullable := f.Props.IsNullable == nil || *f.Props.IsNullable
		defaultVal, _ := f.Props.DefaultVal.(string)

		if isUnique {
			if isNullable {
				return database.SQLNull
			}
			return nil
		}

		if isNullable {
			if defaultVal != "" {
				return defaultVal
			}
			return database.SQLNull
		}

		if defaultVal != "" {
			return defaultVal
		}
		return nil
	}
	if f.Type == TypeRelation {
		return []string{}
	}
	if s, ok := f.Props.DefaultVal.(string); ok && s == "" {
		if f.Type != TypeText {
			return nil
		}
	}

	return f.Props.DefaultVal
}

func (s CollectionSchema) Sanitize() {
	for i := range s {
		f := &s[i]
		if f.Type == TypeRelation {
			f.Props.IsNullable = boolPtr(false)
			f.Props.IsUnique = boolPtr(false)
			f.Props.DefaultVal = []string{}

			if f.Props.Relation != nil {
				if f.Props.Relation.RelationMode == "" {
					f.Props.Relation.RelationMode = RelationModeSingle
				}
			}
		}
	}
}

func boolPtr(b bool) *bool {
	return &b
}
