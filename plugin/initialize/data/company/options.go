package company

import (
	systemStructs "ncobase/system/structs"

	"github.com/ncobase/ncore/version"
)

// SystemDefaultOptions defines core system configuration options
var SystemDefaultOptions = []systemStructs.OptionsBody{
	// Basic system info
	{
		Name:     "system.name",
		Type:     "string",
		Value:    "Digital Company Platform",
		Autoload: true,
	},
	{
		Name:     "system.description",
		Type:     "string",
		Value:    "Multi-tenant digital company management platform",
		Autoload: true,
	},
	{
		Name:     "system.version",
		Type:     "object",
		Value:    version.GetVersionInfo().JSON(),
		Autoload: true,
	},

	// UI theme settings
	{
		Name:     "system.theme",
		Type:     "object",
		Value:    `{"primaryColor":"#1890ff","layout":"side","darkMode":false,"compactMode":false}`,
		Autoload: true,
	},

	// Security settings
	{
		Name:     "system.security",
		Type:     "object",
		Value:    `{"passwordMinLength":8,"passwordComplexity":true,"loginAttempts":5,"sessionTimeout":480}`,
		Autoload: true,
	},

	// Default settings
	{
		Name:     "system.defaults",
		Type:     "object",
		Value:    `{"language":"en-US","timezone":"UTC","dateFormat":"YYYY-MM-DD","currency":"USD"}`,
		Autoload: true,
	},

	// Notification settings
	{
		Name:     "system.notifications",
		Type:     "object",
		Value:    `{"email":true,"push":true,"in_app":true,"digest_frequency":"daily"}`,
		Autoload: true,
	},

	// Multi-tenant settings
	{
		Name:     "system.multi_tenant",
		Type:     "object",
		Value:    `{"enabled":true,"isolation_level":"strict","tenant_creation":"admin_only"}`,
		Autoload: true,
	},

	// Employee settings
	{
		Name:     "system.employee",
		Type:     "object",
		Value:    `{"auto_employee_id":true,"employee_id_prefix":"EMP","probation_period_days":90}`,
		Autoload: true,
	},

	// Dashboard settings
	{
		Name:     "dashboard.default",
		Type:     "object",
		Value:    `{"widgets":["user_stats","recent_activities","system_health"],"refresh_interval":300}`,
		Autoload: true,
	},
}
