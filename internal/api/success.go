package api

type SuccessMessage struct {
	Message string `json:"message"`
}

var (
	SuccessOK = SuccessMessage{Message: "Operation successful"}

	SuccessAdminLoggedIn  = SuccessMessage{Message: "Admin logged in successfully"}
	SuccessAdminLoggedOut = SuccessMessage{Message: "Admin logged out successfully"}
	SuccessAdminUpdated   = SuccessMessage{Message: "Admin account updated successfully"}
	SuccessAdminDeleted   = SuccessMessage{Message: "Admin account deleted successfully"}

	SuccessCollectionCreated = SuccessMessage{Message: "Collection created successfully"}
	SuccessCollectionUpdated = SuccessMessage{Message: "Collection updated successfully"}
	SuccessCollectionDeleted = SuccessMessage{Message: "Collection deleted successfully"}
	SuccessRecordUpdated     = SuccessMessage{Message: "Record updated successfully"}
	SuccessRecordDeleted     = SuccessMessage{Message: "Record deleted successfully"}
	SuccessNoChanges         = SuccessMessage{Message: "No changes applied"}

	SuccessEnvUpdated   = SuccessMessage{Message: "Environment variable updated successfully"}
	SuccessEnvDeleted   = SuccessMessage{Message: "Variables deleted successfully"}
	SuccessEnvNoChanges = SuccessMessage{Message: "No variables selected"}

	SuccessBackupCreated = SuccessMessage{Message: "Backup created successfully"}
	SuccessBackupDeleted = SuccessMessage{Message: "Backup deleted successfully"}
	SuccessConfigUpdated = SuccessMessage{Message: "Configuration updated successfully"}

	SuccessPostgresConfigUpdated = SuccessMessage{Message: "PostgreSQL configuration updated successfully. Please restart the engine for changes to take full effect."}
	SuccessPostgresConfUpdated   = SuccessMessage{Message: "postgresql.conf updated successfully. You must restart PostgreSQL for changes to take effect."}
	SuccessPostgresTestOK        = SuccessMessage{Message: "Connection test successful!"}

	SuccessEngineRestarting     = SuccessMessage{Message: "API routes are being re-initialized and engine is restarting..."}
	SuccessEngineStopping       = SuccessMessage{Message: "Public API service is stopping..."}
	SuccessEngineStarting       = SuccessMessage{Message: "Public API service is starting..."}
	SuccessEngineAlreadyRunning = SuccessMessage{Message: "Public API service is already running"}
	SuccessEngineAlreadyStopped = SuccessMessage{Message: "API routes are already stopped"}

	SuccessSetupDBTestOK = SuccessMessage{Message: "Database connection test successful"}

	SuccessPublicOnline = SuccessMessage{Message: "Blinky Public API Online"}
	SuccessAdminOnline  = SuccessMessage{Message: "Blinky Admin API Online"}
)
