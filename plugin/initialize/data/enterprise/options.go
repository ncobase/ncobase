package enterprise

import (
	"ncobase/core/system/structs"

	"ncobase/internal/version"
)

// SystemDefaultOptions defines default system configuration options
var SystemDefaultOptions = []structs.OptionBody{
	// Basic system settings
	{
		Name:     "system.name",
		Type:     "string",
		Value:    "Digital Enterprise Platform",
		Autoload: true,
	},
	{
		Name:     "system.description",
		Type:     "string",
		Value:    "Multi-space digital enterprise management and collaboration platform",
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
		Value:    `{"primaryColor":"#1890ff","layout":"side","contentWidth":"fluid","fixedHeader":true,"fixSiderbar":true,"colorWeak":false,"title":"Enterprise Platform","logo":"/logo.png","darkMode":false,"compactMode":false}`,
		Autoload: true,
	},

	// Storage settings
	{
		Name:     "system.storage",
		Type:     "object",
		Value:    `{"type":"local","local":{"directory":"uploads"},"oss":{"endpoint":"","accessKeyId":"","accessKeySecret":"","bucket":""},"s3":{"endpoint":"","accessKeyId":"","accessKeySecret":"","bucket":""}}`,
		Autoload: true,
	},

	// Security settings
	{
		Name:     "system.security",
		Type:     "object",
		Value:    `{"passwordMinLength":8,"passwordComplexity":true,"loginAttempts":5,"lockoutDuration":30,"sessionTimeout":480,"mfaRequired":false,"ipWhitelist":[],"auditLogging":true}`,
		Autoload: true,
	},

	// Email settings
	{
		Name:     "system.email",
		Type:     "object",
		Value:    `{"fromName":"System Admin","fromEmail":"admin@example.com","smtp":{"host":"smtp.example.com","port":587,"secure":true,"auth":{"user":"","pass":""}}}`,
		Autoload: true,
	},

	// Default settings
	{
		Name:     "system.defaults",
		Type:     "object",
		Value:    `{"language":"en-US","timezone":"UTC","dateFormat":"YYYY-MM-DD","timeFormat":"HH:mm:ss","currency":"USD"}`,
		Autoload: true,
	},

	// Notification settings
	{
		Name:     "system.notifications",
		Type:     "object",
		Value:    `{"email":true,"sms":false,"push":true,"in_app":true,"digest_frequency":"daily","channels":{"hr":"email","finance":"email","system":"push"}}`,
		Autoload: true,
	},

	// Integration settings
	{
		Name:     "system.integrations",
		Type:     "object",
		Value:    `{"ldap":{"enabled":false,"server":"","domain":""},"sso":{"enabled":false,"provider":"","config":{}},"hr_system":{"enabled":false,"api_endpoint":"","sync_frequency":"daily"}}`,
		Autoload: true,
	},

	// Audit settings
	{
		Name:     "system.audit",
		Type:     "object",
		Value:    `{"enabled":true,"logLogin":true,"logOperations":true,"retention":90}`,
		Autoload: true,
	},

	// Backup settings
	{
		Name:     "system.backup",
		Type:     "object",
		Value:    `{"enabled":true,"schedule":"0 2 * * *","retention_days":30,"include_databases":true,"include_files":true,"offsite_backup":false,"encryption":true}`,
		Autoload: true,
	},

	// Performance settings
	{
		Name:     "system.performance",
		Type:     "object",
		Value:    `{"cacheEnabled":true,"cacheTTL":3600,"compressResponses":true,"rateLimiting":{"enabled":true,"requestsPerMinute":120},"database_pooling":{"max_connections":100}}`,
		Autoload: true,
	},

	// Multi-space settings
	{
		Name:     "system.multi_space",
		Type:     "object",
		Value:    `{"enabled":true,"isolation_level":"strict","shared_resources":["system","menu","dictionary"],"space_creation":"admin_only"}`,
		Autoload: true,
	},

	// Employee management settings
	{
		Name:     "system.employee",
		Type:     "object",
		Value:    `{"auto_employee_id":true,"employee_id_prefix":"EMP","probation_period_days":90,"annual_leave_days":21,"sick_leave_days":10}`,
		Autoload: true,
	},

	// Organization settings
	{
		Name:     "system.organization",
		Type:     "object",
		Value:    `{"max_hierarchy_levels":5,"allow_cross_company_assignment":true,"require_manager_approval":true,"auto_org_chart":true}`,
		Autoload: true,
	},

	// Workflow settings
	{
		Name:     "system.workflow",
		Type:     "object",
		Value:    `{"approval_required_for":["employee_creation","role_assignment","department_transfer"],"auto_notifications":true,"escalation_timeout_hours":24}`,
		Autoload: true,
	},

	// Reporting and analytics
	{
		Name:     "system.analytics",
		Type:     "object",
		Value:    `{"enabled":true,"retention_days":365,"anonymize_pii":true,"dashboard_refresh_interval":300,"export_formats":["pdf","excel","csv"]}`,
		Autoload: true,
	},

	// Compliance and legal
	{
		Name:     "system.compliance",
		Type:     "object",
		Value:    `{"gdpr_enabled":true,"data_retention_days":2555,"audit_trail":true,"encryption_at_rest":true,"anonymization_rules":{"employee_data":365,"financial_data":2555}}`,
		Autoload: true,
	},

	// Dashboard settings
	{
		Name:     "dashboard.enterprise",
		Type:     "object",
		Value:    `{"widgets":["employee_count","active_projects","department_overview","financial_summary","recent_activities","system_health"],"refresh_interval":300,"layout":"grid"}`,
		Autoload: true,
	},

	// HR specific settings
	{
		Name:     "hr.settings",
		Type:     "object",
		Value:    `{"performance_review_cycle":"annual","goal_setting":"quarterly","skill_assessment":true,"career_path_planning":true,"succession_planning":false}`,
		Autoload: true,
	},

	// Finance specific settings
	{
		Name:     "finance.settings",
		Type:     "object",
		Value:    `{"budget_approval_workflow":true,"expense_categories":["travel","equipment","training","marketing"],"currency":"USD","fiscal_year_start":"2024-01-01"}`,
		Autoload: true,
	},
}
