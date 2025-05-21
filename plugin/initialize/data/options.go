package data

import "ncobase/system/structs"

// SystemDefaultOptions defines default system configuration options
var SystemDefaultOptions = []structs.OptionsBody{
	// Basic system settings
	{
		Name:     "system.name",
		Type:     "string",
		Value:    "Digital Development Platform",
		Autoload: true,
	},
	{
		Name:     "system.description",
		Type:     "string",
		Value:    "A comprehensive platform for digital transformation and development",
		Autoload: true,
	},
	{
		Name:     "system.version",
		Type:     "string",
		Value:    "1.0.0",
		Autoload: true,
	},

	// UI theme settings
	{
		Name:     "system.theme",
		Type:     "object",
		Value:    `{"primaryColor":"#1890ff","layout":"side","contentWidth":"fluid","fixedHeader":true,"fixSiderbar":true,"colorWeak":false,"title":"Digital Platform","logo":"/logo.png","darkMode":false}`,
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
		Value:    `{"passwordMinLength":8,"passwordComplexity":true,"loginAttempts":5,"lockoutDuration":30,"sessionTimeout":120,"allowedIps":[],"twoFactorAuth":false}`,
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
		Value:    `{"email":true,"sms":false,"push":false,"wechat":false}`,
		Autoload: true,
	},

	// Integration settings
	{
		Name:     "system.integrations",
		Type:     "object",
		Value:    `{"wechat":{"enabled":false,"appId":"","appSecret":""},"dingtalk":{"enabled":false,"appKey":"","appSecret":""},"slack":{"enabled":false,"token":""},"teams":{"enabled":false,"webhookUrl":""}}`,
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
		Value:    `{"enabled":true,"schedule":"0 0 * * 0","retention":10,"includeDatabases":true,"includeFiles":true,"destination":"local"}`,
		Autoload: true,
	},

	// Performance settings
	{
		Name:     "system.performance",
		Type:     "object",
		Value:    `{"cacheEnabled":true,"cacheTTL":3600,"compressResponses":true,"rateLimiting":{"enabled":true,"requestsPerMinute":60}}`,
		Autoload: true,
	},

	// Dashboard default settings
	{
		Name:     "dashboard.default",
		Type:     "object",
		Value:    `{"widgets":["activeUsers","recentActivity","systemStatus","quickActions"],"refreshInterval":300}`,
		Autoload: true,
	},
}
