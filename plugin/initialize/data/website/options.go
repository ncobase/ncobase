package website

import (
	systemStructs "ncobase/core/system/structs"

	"ncobase/internal/version"
)

// SystemDefaultOptions for regular websites
var SystemDefaultOptions = []systemStructs.OptionBody{
	// Basic system info
	{
		Name:     "system.name",
		Type:     "string",
		Value:    "Website Platform",
		Autoload: true,
	},
	{
		Name:     "system.description",
		Type:     "string",
		Value:    "Simple website management platform",
		Autoload: true,
	},
	{
		Name:     "system.version",
		Type:     "object",
		Value:    version.GetVersionInfo().JSON(),
		Autoload: true,
	},
	{
		Name:     "system.frontend",
		Type:     "object",
		Value:    `{"sign_in_url":"http://localhost:3000/login","sign_up_url":"http://localhost:3000/register"}`,
		Autoload: true,
	},

	// UI theme
	{
		Name:     "system.theme",
		Type:     "object",
		Value:    `{"primaryColor":"#2563eb","layout":"side","darkMode":false}`,
		Autoload: true,
	},

	// Security
	{
		Name:     "system.security",
		Type:     "object",
		Value:    `{"passwordMinLength":6,"loginAttempts":5,"sessionTimeout":720}`,
		Autoload: true,
	},
	{
		Name:     "auth.token",
		Type:     "object",
		Value:    `{"access_token_expiry":"2h","refresh_token_expiry":"7d","register_token_expiry":"30m","mfa_token_expiry":"5m"}`,
		Autoload: true,
	},
	{
		Name:     "auth.session",
		Type:     "object",
		Value:    `{"max_sessions":10,"session_expiry":"7d","cleanup_interval":"1h"}`,
		Autoload: true,
	},

	// Defaults
	{
		Name:     "system.defaults",
		Type:     "object",
		Value:    `{"language":"en-US","timezone":"UTC","dateFormat":"YYYY-MM-DD"}`,
		Autoload: true,
	},

	// Notifications
	{
		Name:     "system.notifications",
		Type:     "object",
		Value:    `{"email":true,"in_app":true}`,
		Autoload: true,
	},
	{
		Name:     "system.storage_policy",
		Type:     "object",
		Value:    `{"default_provider":"configured","allow_public_links":true,"require_owner_scope":true,"audit_downloads":true}`,
		Autoload: true,
	},
	{
		Name:     "system.email_policy",
		Type:     "object",
		Value:    `{"enabled":true,"sender_name":"System Admin","allow_auth_email":true,"allow_password_reset":true,"digest_frequency":"daily"}`,
		Autoload: true,
	},
	{
		Name:     "resource.upload",
		Type:     "object",
		Value:    `{"max_upload_size":5368709120,"allowed_types":["*"],"default_storage":"configured"}`,
		Autoload: true,
	},
	{
		Name:     "resource.image",
		Type:     "object",
		Value:    `{"enable_thumbnails":true,"default_thumbnail_width":300,"default_thumbnail_height":300,"enable_resizing":true,"max_image_width":2048,"max_image_height":2048,"compression_quality":85}`,
		Autoload: true,
	},
	{
		Name:     "resource.quota",
		Type:     "object",
		Value:    `{"enable_quotas":true,"enable_enforcement":true,"default_quota":10737418240,"warning_threshold":0.8,"quota_check_interval":"24h"}`,
		Autoload: true,
	},

	// Dashboard
	{
		Name:     "dashboard.default",
		Type:     "object",
		Value:    `{"widgets":["user_stats","content_stats","recent_activities"]}`,
		Autoload: true,
	},
}
