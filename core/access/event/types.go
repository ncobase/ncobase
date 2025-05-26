package event

// Event categories for activity logging
const (
	CategoryAuth     = "auth"
	CategoryUser     = "user"
	CategorySystem   = "system"
	CategorySecurity = "security"
	CategoryData     = "data"
)

// Event types
const (
	// Auth events

	TypeLogin         = "login"
	TypeLogout        = "logout"
	TypeTokenRefresh  = "token_refresh"
	TypePasswordReset = "password_reset"

	// User events

	TypeUserCreated   = "created"
	TypeUserUpdated   = "updated"
	TypeUserDeleted   = "deleted"
	TypeProfileUpdate = "profile_updated"
	TypeStatusChange  = "status_changed"
	TypeApiKeyGen     = "api_key_generated"
	TypeApiKeyDel     = "api_key_deleted"

	// System events

	TypeSystemStart   = "started"
	TypeSystemStop    = "stopped"
	TypeSystemRestart = "restarted"
	TypeSystemUpgrade = "upgraded"
	TypeConfigChange  = "config_changed"

	// Security events

	TypeSecurityAlert   = "alert"
	TypeSecurityBlock   = "blocked"
	TypeSecurityViolate = "violation"
	TypeSecurityUnlock  = "unlocked"

	// Data events

	TypeDataAccess   = "accessed"
	TypeDataExport   = "exported"
	TypeDataImport   = "imported"
	TypeDataBackup   = "backup"
	TypeDataRestore  = "restore"
	TypeDataShare    = "shared"
	TypeDataDownload = "downloaded"
)

// Event names (category.type format)
const (
	UserLogin           = "user.login"
	UserCreated         = "user.created"
	UserUpdated         = "user.updated"
	UserDeleted         = "user.deleted"
	UserPasswordChanged = "user.password_changed"
	UserPasswordReset   = "user.password_reset"
	UserProfileUpdated  = "user.profile_updated"
	UserStatusUpdated   = "user.status_updated"
	UserApiKeyGen       = "user.apikey_generated"
	UserApiKeyDel       = "user.apikey_deleted"
	UserAuthCodeSent    = "user.auth_code_sent"

	SystemModified  = "system.modified"
	SystemStarted   = "system.started"
	SystemStopped   = "system.stopped"
	SystemRestarted = "system.restarted"
	SystemUpgraded  = "system.upgraded"

	SecurityIncident  = "security.incident"
	SecurityViolation = "security.violation"
	SecurityGranted   = "security.granted"
	SecurityBlocked   = "security.blocked"

	SecurityAccessGranted   = "security.access_granted"
	SecurityAccessDenied    = "security.access_denied"
	SecurityAccessRevoked   = "security.access_revoked"
	SecurityAccessRequested = "security.access_requested"
	SecurityAccessModified  = "security.access_modified"

	DataAccessed   = "data.accessed"
	DataModified   = "data.modified"
	DataExported   = "data.exported"
	DataImported   = "data.imported"
	DataShared     = "data.shared"
	DataDownloaded = "data.downloaded"
)
