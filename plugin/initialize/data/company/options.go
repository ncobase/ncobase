package company

import (
	systemStructs "ncobase/core/system/structs"

	"ncobase/internal/version"
)

// SystemDefaultOptions defines core system configuration options
var SystemDefaultOptions = []systemStructs.OptionBody{
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
		Value:    "Multi-space digital company management platform",
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

	// Multi-space settings
	{
		Name:     "system.multi_space",
		Type:     "object",
		Value:    `{"enabled":true,"isolation_level":"strict","space_creation":"admin_only"}`,
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
	{
		Name:     "ai.provider",
		Type:     "object",
		Value:    `{"enabled":false,"providers":[{"name":"openai","type":"openai","enabled":true,"base_url":"https://api.openai.com","api_key_env":"OPENAI_API_KEY"},{"name":"anthropic","type":"anthropic","enabled":false,"base_url":"https://api.anthropic.com","api_key_env":"ANTHROPIC_API_KEY"},{"name":"gemini","type":"gemini","enabled":false,"base_url":"https://generativelanguage.googleapis.com","api_key_env":"GEMINI_API_KEY"},{"name":"ollama","type":"ollama","enabled":false,"base_url":"http://localhost:11434"},{"name":"cohere","type":"cohere","enabled":false,"base_url":"https://api.cohere.ai","api_key_env":"COHERE_API_KEY"}]}`,
		Autoload: true,
	},
	{
		Name:     "ai.model",
		Type:     "object",
		Value:    `{"primary":"openai/gpt-4o-mini","fallbacks":[],"default_max_output_tokens":1024,"default_temperature":0.2}`,
		Autoload: true,
	},
	{
		Name:     "ai.policy",
		Type:     "object",
		Value:    `{"enabled":true,"allowed_actions":["builder.api","builder.form","builder.menu","builder.schema","builder.tests","content.review","content.seo","content.summary","content.tags","content.title","content.translation","payment.summary","realtime.summary","resource.description","resource.summary","resource.tags","system.summary"],"allowed_provider_types":["anthropic","cohere","gemini","ollama","openai"],"max_prompt_chars":20000,"max_messages":32,"max_input_items":64,"max_output_tokens":4096,"timeout_seconds":45,"retry":2,"rate_limit_per_second":0,"circuit_breaker":{"max_failures":5,"reset_seconds":60},"store_raw_output":false,"require_configured_model":true}`,
		Autoload: true,
	},
	{
		Name:     "ai.safety",
		Type:     "object",
		Value:    `{"redact_prompts":true,"store_request_hash":true,"max_error_chars":500,"blocked_phrases":[],"allow_system_prompts":true}`,
		Autoload: true,
	},
	{
		Name:     "ai.cost",
		Type:     "object",
		Value:    `{"currency":"USD","models":{}}`,
		Autoload: true,
	},
	{
		Name:     "ai.embedding",
		Type:     "object",
		Value:    `{"enabled":true,"model":"openai/text-embedding-3-small","input_type":"search_document","max_items":64}`,
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
