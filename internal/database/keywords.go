package database

const (
	SQLCreateTable       = "CREATE TABLE IF NOT EXISTS"
	SQLAlterTable        = "ALTER TABLE"
	SQLCreateIndex       = "CREATE INDEX IF NOT EXISTS"
	SQLCreateUniqueIndex = "CREATE UNIQUE INDEX IF NOT EXISTS"
	SQLDropIndex         = "DROP INDEX IF EXISTS"
	SQLDropTable         = "DROP TABLE IF EXISTS"
	SQLCreateDatabase    = "CREATE DATABASE"
	SQLTruncateTable     = "TRUNCATE TABLE"
	SQLAsc               = "ASC"
	SQLDesc              = "DESC"
)

const (
	TypeText      = "TEXT"
	TypeVarchar   = "VARCHAR"
	TypeInteger   = "INTEGER"
	TypeBigInt    = "BIGINT"
	TypeNumeric   = "NUMERIC"
	TypeBoolean   = "BOOLEAN"
	TypeJSONB     = "JSONB"
	TypeTimestamp = "TIMESTAMP WITH TIME ZONE"
	TypeUUID      = "UUID"
)

const (
	OIDUUID = 2950
)

const (
	SQLPrimaryKey = "PRIMARY KEY"
	SQLNotNull    = "NOT NULL"
	SQLNull       = "NULL"
	SQLUnique     = "UNIQUE"
	SQLCheck      = "CHECK"
	SQLConstraint = "CONSTRAINT"
	SQLReferences = "REFERENCES"
	SQLOnDelete   = "ON DELETE"
	SQLCascade    = "CASCADE"
	SQLSetNull    = "SET NULL"
	SQLDefault    = "DEFAULT"
	SQLOn         = "ON"
	SQLNow        = "CURRENT_TIMESTAMP"
	SQLForeignKey = "FOREIGN KEY"
)

const (
	OpAddColumn      = "ADD COLUMN IF NOT EXISTS"
	OpDropColumn     = "DROP COLUMN IF EXISTS"
	OpRenameTo       = "RENAME TO"
	OpRenameColumn   = "RENAME COLUMN"
	OpAlterColumn    = "ALTER COLUMN"
	OpSetNotNull     = "SET NOT NULL"
	OpDropNotNull    = "DROP NOT NULL"
	OpSetDefault     = "SET DEFAULT"
	OpDropDefault    = "DROP DEFAULT"
	OpAddConstraint  = "ADD CONSTRAINT"
	OpDropConstraint = "DROP CONSTRAINT IF EXISTS"
)

const (
	PrefixIndex      = "idx"
	PrefixForeignKey = "fk"
	PrefixCheck      = "chk"
)

const (
	SQLTrue  = "true"
	SQLFalse = "false"
)

const (
	FuncTrim       = "trim"
	FuncCharLength = "char_length"
	FuncCountAll   = "COUNT(*)"
	FuncCount      = "COUNT"
	FuncCoalesce   = "COALESCE"
)

const (
	SQLTablePgTables           = "pg_catalog.pg_tables"
	SQLTableInformationColumns = "information_schema.columns"
	SQLColumnSchemaname        = "schemaname"
	SQLColumnTablename         = "tablename"
	SQLColumnTableName         = "table_name"
	SQLColumnColumnName        = "column_name"
	SQLColumnIsNullable        = "is_nullable"
	SQLSchemaPublic            = "public"
)

const (
	SQLTableCollections               = "_collections"
	SQLTableCollectionsID             = "id"
	SQLTableCollectionsName           = "name"
	SQLTableCollectionsSchema         = "schema"
	SQLTableCollectionsIsSystem       = "is_system"
	SQLTableCollectionsTotalDocuments = "total_documents"
	SQLTableCollectionsTotalBytes     = "total_bytes"
	SQLTableCollectionsCreatedAt      = "created_at"
	SQLTableCollectionsUpdatedAt      = "updated_at"
)

const (
	SQLTableAdmins             = "_admins"
	SQLTableAdminsID           = "id"
	SQLTableAdminsNickname     = "nickname"
	SQLTableAdminsUsername     = "username"
	SQLTableAdminsEmail        = "email"
	SQLTableAdminsPasswordHash = "password_hash"
	SQLTableAdminsAvatar       = "avatar"
	SQLTableAdminsToken        = "token"
	SQLTableAdminsCreatedAt    = "created_at"
	SQLTableAdminsUpdatedAt    = "updated_at"
)

const (
	SQLRecordID        = "id"
	SQLRecordCreatedAt = "created_at"
	SQLRecordUpdatedAt = "updated_at"
)

const (
	SQLSpace       = " "
	SQLComma       = ","
	SQLUnderscore  = "_"
	SQLQuote       = "\""
	SQLSingleQuote = "'"
	SQLOpenParen   = "("
	SQLCloseParen  = ")"
	SQLSemicolon   = ";"
	SQLGte         = ">="
	SQLLte         = "<="
	SQLEq          = "="
	SQLAnd         = "AND"
	SQLOr          = "OR"
	SQLNot         = "NOT"
	SQLIs          = "IS"
	SQLIn          = "IN"
	SQLType        = "TYPE"
	SQLUsing       = "USING"
	SQLCast        = "::"
	SQLToJsonb     = "to_jsonb"
	SQLNotEq       = "!="
	SQLZero        = "0"
	SQLNewline     = "\n"
	SQLTab         = "\t"
	SQLAsterisk    = "*"
	SQLEmptyString = "''"
	SQLPlus        = "+"
	SQLMinus       = "-"
	SQLOne         = "1"
	SQLTo          = "TO"
)
