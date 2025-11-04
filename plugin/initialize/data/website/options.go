package website

import (
	systemStructs "ncobase/system/structs"

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

	// Dashboard
	{
		Name:     "dashboard.default",
		Type:     "object",
		Value:    `{"widgets":["user_stats","content_stats","recent_activities"]}`,
		Autoload: true,
	},
}
