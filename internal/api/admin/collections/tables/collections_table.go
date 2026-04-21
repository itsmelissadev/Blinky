package tables

import (
	"blinky/internal/database"
	"blinky/internal/types"
)

func GetCollectionsSchema() types.CollectionSchema {
	f := false
	t := true

	return types.CollectionSchema{
		{
			ID:   types.GenerateFieldID("_collections_name"),
			Name: database.SQLTableCollectionsName,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &f,
				IsUnique:   &t,
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_schema"),
			Name: database.SQLTableCollectionsSchema,
			Type: types.TypeJSON,
			Props: types.FieldProps{
				IsNullable: &f,
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_is_system"),
			Name: database.SQLTableCollectionsIsSystem,
			Type: types.TypeBoolean,
			Props: types.FieldProps{
				IsNullable: &f,
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_total_documents"),
			Name: database.SQLTableCollectionsTotalDocuments,
			Type: types.TypeNumber,
			Props: types.FieldProps{
				IsNullable: &f,
				Number: &types.NumberOptions{
					NoDecimals: true,
				},
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_total_bytes"),
			Name: database.SQLTableCollectionsTotalBytes,
			Type: types.TypeNumber,
			Props: types.FieldProps{
				IsNullable: &f,
				Number: &types.NumberOptions{
					NoDecimals: true,
				},
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_created_at"),
			Name: database.SQLTableCollectionsCreatedAt,
			Type: types.TypeDate,
			Props: types.FieldProps{
				IsNullable:    &f,
				AutoTimestamp: types.AutoTimestampCreate,
			},
		},
		{
			ID:   types.GenerateFieldID("_collections_updated_at"),
			Name: database.SQLTableCollectionsUpdatedAt,
			Type: types.TypeDate,
			Props: types.FieldProps{
				IsNullable:    &f,
				AutoTimestamp: types.AutoTimestampCreateUpdate,
			},
		},
	}
}
