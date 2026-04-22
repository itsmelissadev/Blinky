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
		AddColumn(NewColumn(SQLTableAdminsNickname).Varchar(255).NotNull().Build()).
		AddColumn(NewColumn(SQLTableAdminsUsername).Varchar(255).NotNull().Unique().Build()).
		AddColumn(NewColumn(SQLTableAdminsEmail).Varchar(255).NotNull().Unique().Build()).
		AddColumn(NewColumn(SQLTableAdminsPasswordHash).Varchar(255).NotNull().Build()).
		AddColumn(NewColumn(SQLTableAdminsAvatar).Varchar(255).Build()).
		AddColumn(NewColumn(SQLTableAdminsToken).Varchar(255).Build()).
		AddColumn(NewColumn(SQLTableAdminsCreatedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		AddColumn(NewColumn(SQLTableAdminsUpdatedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		Build()
}

func GetMigrationsTableSQL() string {
	return NewTable(SQLTableMigrations).
		AddColumn(NewColumn(SQLTableMigrationsID).Varchar(255).PrimaryKey().Build()).
		AddColumn(NewColumn(SQLTableMigrationsVersion).Integer().NotNull().Build()).
		AddColumn(NewColumn(SQLTableMigrationsAppliedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		Build()
}

func GetMigrationLogsTableSQL() string {
	return NewTable(SQLTableMigrationLogs).
		AddColumn(NewColumn(SQLTableMigrationLogsID).Varchar(255).PrimaryKey().Build()).
		AddColumn(NewColumn(SQLTableMigrationLogsVersion).Integer().NotNull().Build()).
		AddColumn(NewColumn(SQLTableMigrationLogsStatus).Varchar(50).NotNull().Build()).
		AddColumn(NewColumn(SQLTableMigrationLogsMessage).Text().Build()).
		AddColumn(NewColumn(SQLTableMigrationLogsDuration).BigInt().Build()).
		AddColumn(NewColumn(SQLTableMigrationLogsCreatedAt).Timestamp().NotNull().Default(SQLNow).Build()).
		Build()
}
