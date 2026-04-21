package api

import "strings"

type ApiError struct {
	Number  int    `json:"number"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e ApiError) WithReplacements(reps map[string]string) ApiError {
	newMsg := e.Message
	for k, v := range reps {
		newMsg = strings.ReplaceAll(newMsg, "{"+k+"}", v)
	}
	return ApiError{
		Number:  e.Number,
		Code:    e.Code,
		Message: newMsg,
	}
}

var (
	ErrCoreInternalServer = ApiError{Number: 1000, Code: "CORE_INTERNAL_SERVER_ERROR", Message: "An unexpected error occurred"}
	ErrCoreInvalidBody    = ApiError{Number: 1001, Code: "CORE_INVALID_BODY", Message: "Invalid request body"}
	ErrCoreUnauthorized   = ApiError{Number: 1002, Code: "CORE_UNAUTHORIZED", Message: "Unauthorized access: API Key required"}
	ErrCoreForbidden      = ApiError{Number: 1003, Code: "CORE_FORBIDDEN", Message: "Forbidden: You do not have permission to access this resource"}

	ErrCollectionBuildQuery          = ApiError{Number: 2000, Code: "COLLECTION_BUILD_ERROR", Message: "Failed to build database query"}
	ErrCollectionQueryExecution      = ApiError{Number: 2001, Code: "COLLECTION_QUERY_ERROR", Message: "Failed to execute database query"}
	ErrCollectionScanData            = ApiError{Number: 2002, Code: "COLLECTION_SCAN_ERROR", Message: "Failed to scan database result"}
	ErrCollectionNameRequired        = ApiError{Number: 2003, Code: "COLLECTION_NAME_REQUIRED", Message: "Name is required"}
	ErrCollectionInvalidName         = ApiError{Number: 2004, Code: "COLLECTION_INVALID_NAME", Message: "Invalid collection name. Use lowercase, numbers and underscores only."}
	ErrCollectionAlreadyExists       = ApiError{Number: 2005, Code: "COLLECTION_ALREADY_EXISTS", Message: "Collection already exists"}
	ErrCollectionCreateError         = ApiError{Number: 2006, Code: "COLLECTION_CREATE_ERROR", Message: "Failed to create collection"}
	ErrCollectionInvalidSchema       = ApiError{Number: 2007, Code: "COLLECTION_INVALID_SCHEMA", Message: "The provided collection schema is invalid"}
	ErrCollectionUnknownType         = ApiError{Number: 2008, Code: "COLLECTION_UNKNOWN_TYPE", Message: "The provided field type is unknown"}
	ErrCollectionNotFound            = ApiError{Number: 2009, Code: "COLLECTION_NOT_FOUND", Message: "The specified collection was not found"}
	ErrCollectionInvalidLength       = ApiError{Number: 2010, Code: "COLLECTION_INVALID_LENGTH", Message: "Field '{field}': {reason}"}
	ErrCollectionAutoManagedField    = ApiError{Number: 2011, Code: "COLLECTION_AUTO_MANAGED_FIELD", Message: "Field '{field}' is auto-generated, you cannot save your own time here."}
	ErrCollectionColumnNotFound      = ApiError{Number: 2012, Code: "COLLECTION_COLUMN_NOT_FOUND", Message: "Field '{field}' was not found in collection '{collection}'"}
	ErrCollectionInvalidOperator     = ApiError{Number: 2013, Code: "COLLECTION_INVALID_OPERATOR", Message: "Invalid operator '{operator}' for field '{field}'"}
	ErrCollectionUniqueViolation     = ApiError{Number: 2014, Code: "COLLECTION_UNIQUE_VIOLATION", Message: "A record with this value already exists."}
	ErrCollectionNotNullViolation    = ApiError{Number: 2015, Code: "COLLECTION_NOT_NULL_VIOLATION", Message: "A required field is missing."}
	ErrCollectionCheckViolation      = ApiError{Number: 2016, Code: "COLLECTION_CHECK_VIOLATION", Message: "A field value does not meet the requirements."}
	ErrCollectionForeignKeyViolation = ApiError{Number: 2017, Code: "COLLECTION_FOREIGN_KEY_VIOLATION", Message: "A related record does not exist or cannot be modified."}
	ErrCollectionSystemProtected     = ApiError{Number: 2018, Code: "COLLECTION_SYSTEM_PROTECTED", Message: "System collections cannot be modified"}
	ErrCollectionDropFailed          = ApiError{Number: 2019, Code: "COLLECTION_DROP_FAILED", Message: "Failed to drop collection table: {error}"}
	ErrCollectionTruncateFailed      = ApiError{Number: 2020, Code: "COLLECTION_TRUNCATE_FAILED", Message: "Failed to truncate collection table: {error}"}

	ErrAuthRequired                 = ApiError{Number: 3000, Code: "AUTH_REQUIRED", Message: "Admin authentication required"}
	ErrAuthInvalidSession           = ApiError{Number: 3001, Code: "AUTH_INVALID_SESSION", Message: "Invalid or expired session"}
	ErrAuthInvalidCredentials       = ApiError{Number: 3002, Code: "AUTH_INVALID_CREDENTIALS", Message: "Invalid username/email or password"}
	ErrAuthAdminNotFound            = ApiError{Number: 3003, Code: "AUTH_ADMIN_NOT_FOUND", Message: "Admin account not found"}
	ErrAuthDeleteSelfForbidden      = ApiError{Number: 3004, Code: "AUTH_DELETE_SELF_FORBIDDEN", Message: "You cannot delete your own account"}
	ErrAuthDeleteLastAdminForbidden = ApiError{Number: 3005, Code: "AUTH_DELETE_LAST_ADMIN_FORBIDDEN", Message: "You cannot delete the only remaining admin account"}
	ErrAuthInvalidFields            = ApiError{Number: 3006, Code: "AUTH_INVALID_FIELDS", Message: "Required fields are missing or too short"}
	ErrAuthConflict                 = ApiError{Number: 3007, Code: "AUTH_CONFLICT", Message: "Username or email is already taken"}
	ErrAuthUpdateFailed             = ApiError{Number: 3008, Code: "AUTH_UPDATE_FAILED", Message: "Failed to update admin account"}
	ErrAuthDeleteFailed             = ApiError{Number: 3009, Code: "AUTH_DELETE_FAILED", Message: "Failed to delete admin account"}

	ErrBackupDirReadFailed   = ApiError{Number: 4000, Code: "BACKUP_DIR_READ_FAILED", Message: "Failed to read backup directory"}
	ErrBackupDirCreateFailed = ApiError{Number: 4001, Code: "BACKUP_DIR_CREATE_FAILED", Message: "Failed to create backup directory"}
	ErrBackupPgDumpFailed    = ApiError{Number: 4002, Code: "BACKUP_PG_DUMP_FAILED", Message: "Database dump failed. (Command: {command})"}
	ErrBackupZipFailed       = ApiError{Number: 4003, Code: "BACKUP_ZIP_FAILED", Message: "Failed to compress backup file"}
	ErrBackupMissingFilename = ApiError{Number: 4004, Code: "BACKUP_MISSING_FILENAME", Message: "Filename is required"}
	ErrBackupInvalidFilename = ApiError{Number: 4005, Code: "BACKUP_INVALID_FILENAME", Message: "Invalid backup filename"}
	ErrBackupNotFound        = ApiError{Number: 4006, Code: "BACKUP_NOT_FOUND", Message: "Backup file not found"}
	ErrBackupDeleteFailed    = ApiError{Number: 4007, Code: "BACKUP_DELETE_FAILED", Message: "Failed to delete backup file"}
	ErrBackupBrowseFailed    = ApiError{Number: 4008, Code: "BROWSE_FAILED", Message: "Failed to read directory"}

	ErrEnvReadFailed        = ApiError{Number: 5000, Code: "ENV_READ_FAILED", Message: "Failed to read .env file"}
	ErrEnvUpdateFailed      = ApiError{Number: 5001, Code: "ENV_UPDATE_FAILED", Message: "Failed to update env variable"}
	ErrEnvMissingKey        = ApiError{Number: 5002, Code: "ENV_MISSING_KEY", Message: "Key is required"}
	ErrEnvDeleteFailed      = ApiError{Number: 5003, Code: "ENV_DELETE_FAILED", Message: "Failed to delete environment variables"}
	ErrPostgresConfNotFound = ApiError{Number: 5004, Code: "POSTGRES_CONF_NOT_FOUND", Message: "postgresql.conf not found in the specified root path"}
	ErrServerPortInUse      = ApiError{Number: 6000, Code: "SERVER_PORT_IN_USE", Message: "Specified port is already in use or restricted"}
)
