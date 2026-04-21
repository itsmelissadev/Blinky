package database

func GetCollectionsTableSQL() string {
	return NewTable(SQLTableCollections).
		AddColumn(NewColumn(SQLTableCollectionsName).Varchar(255).PrimaryKey().Build()).
		AddColumn(NewColumn(SQLTableCollectionsSchema).JSONB().NotNull().Build()).
		AddColumn(NewColumn(SQLTableCollectionsTotalDocuments).BigInt().NotNull().Default(0).Build()).
		AddColumn(NewColumn(SQLTableCollectionsTotalBytes).BigInt().NotNull().Default(0).Build()).
		AddColumn(NewColumn(SQLTableCollectionsIsSystem).Boolean(false).NotNull().Default(SQLFalse).Build()).
		AddColumn(NewColumn(SQLTableCollectionsCreatedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		AddColumn(NewColumn(SQLTableCollectionsUpdatedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		Build()
}

func GetAdminsTableSQL() string {
	return NewTable(SQLTableAdmins).
		AddColumn(NewColumn(SQLTableAdminsID).Varchar(255).PrimaryKey().Build()).
		Build()
}
