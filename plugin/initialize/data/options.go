package data

import "ncobase/system/structs"

// SystemDefaultOptions defines the default system configuration options
var SystemDefaultOptions = []structs.OptionsBody{
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
	{
		Name:     "system.theme",
		Type:     "object",
		Value:    `{"primaryColor":"#1890ff","layout":"side","contentWidth":"fluid","fixedHeader":true,"fixSiderbar":true,"colorWeak":false,"title":"Digital Platform","logo":"/logo.png"}`,
		Autoload: true,
	},
	{
		Name:     "system.storage",
		Type:     "object",
		Value:    `{"type":"local","local":{"directory":"uploads"}}`,
		Autoload: true,
	},
	{
		Name:     "system.security",
		Type:     "object",
		Value:    `{"passwordMinLength":8,"passwordComplexity":true,"loginAttempts":5,"lockoutDuration":30,"sessionTimeout":120}`,
		Autoload: true,
	},
	{
		Name:     "system.email",
		Type:     "object",
		Value:    `{"fromName":"System Admin","fromEmail":"admin@example.com","smtp":{"host":"smtp.example.com","port":587,"secure":true}}`,
		Autoload: true,
	},
	{
		Name:     "system.defaults",
		Type:     "object",
		Value:    `{"language":"en-US","timezone":"UTC","dateFormat":"YYYY-MM-DD","timeFormat":"HH:mm:ss"}`,
		Autoload: true,
	},
}
