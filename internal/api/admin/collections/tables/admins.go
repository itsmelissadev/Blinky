package tables

import (
	"blinky/internal/database"
	"blinky/internal/types"
)

func GetAdminsSchema() types.CollectionSchema {
	t := true
	f := false
	min3 := 3
	max50 := 50

	return types.CollectionSchema{
		{
			ID:   types.GenerateFieldID("_admins_id"),
			Name: database.SQLTableAdminsID,
			Type: types.TypeID,
			Props: types.FieldProps{
				IsNullable: &f,
				IsUnique:   &t,
				ID: &types.IDConfig{
					AutoGenerateLength: 16,
				},
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_nickname"),
			Name: database.SQLTableAdminsNickname,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &f,
				Text: &types.TextOptions{
					Min: &min3,
					Max: &max50,
				},
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_username"),
			Name: database.SQLTableAdminsUsername,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &f,
				IsUnique:   &t,
				Text: &types.TextOptions{
					Min: &min3,
					Max: &max50,
				},
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_avatar"),
			Name: database.SQLTableAdminsAvatar,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &t,
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_email"),
			Name: database.SQLTableAdminsEmail,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &f,
				IsUnique:   &t,
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_password_hash"),
			Name: database.SQLTableAdminsPasswordHash,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &f,
				IsHidden:   &t,
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_token"),
			Name: database.SQLTableAdminsToken,
			Type: types.TypeText,
			Props: types.FieldProps{
				IsNullable: &t,
				IsUnique:   &t,
				IsHidden:   &t,
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_updated_at"),
			Name: database.SQLTableAdminsUpdatedAt,
			Type: types.TypeDate,
			Props: types.FieldProps{
				IsNullable:    &f,
				AutoTimestamp: types.AutoTimestampCreateUpdate,
			},
		},
		{
			ID:   types.GenerateFieldID("_admins_created_at"),
			Name: database.SQLTableAdminsCreatedAt,
			Type: types.TypeDate,
			Props: types.FieldProps{
				IsNullable:    &f,
				AutoTimestamp: types.AutoTimestampCreate,
			},
		},
	}
}
