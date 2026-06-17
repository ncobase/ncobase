package structs

type ProviderConfig struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Enabled        bool              `json:"enabled"`
	BaseURL        string            `json:"base_url,omitempty"`
	APIMode        string            `json:"api_mode,omitempty"`
	APIKeyEnv      string            `json:"api_key_env,omitempty"`
	BearerTokenEnv string            `json:"bearer_token_env,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	HeaderEnvs     map[string]string `json:"header_envs,omitempty"`
	Organization   string            `json:"organization,omitempty"`
	Project        string            `json:"project,omitempty"`
	UserProject    string            `json:"user_project,omitempty"`
}

type ModelConfig struct {
	Primary                string   `json:"primary"`
	Fallbacks              []string `json:"fallbacks,omitempty"`
	DefaultMaxOutputTokens int      `json:"default_max_output_tokens"`
	DefaultTemperature     float64  `json:"default_temperature"`
}

type EmbeddingConfig struct {
	Enabled   bool   `json:"enabled"`
	Model     string `json:"model,omitempty"`
	InputType string `json:"input_type,omitempty"`
	MaxItems  int    `json:"max_items"`
}

type CircuitBreakerConfig struct {
	MaxFailures  int `json:"max_failures"`
	ResetSeconds int `json:"reset_seconds"`
}

type PolicyConfig struct {
	Enabled                bool                 `json:"enabled"`
	AllowedActions         []string             `json:"allowed_actions,omitempty"`
	AllowedProviderTypes   []string             `json:"allowed_provider_types,omitempty"`
	MaxPromptChars         int                  `json:"max_prompt_chars"`
	MaxMessages            int                  `json:"max_messages"`
	MaxInputItems          int                  `json:"max_input_items"`
	MaxOutputTokens        int                  `json:"max_output_tokens"`
	TimeoutSeconds         int                  `json:"timeout_seconds"`
	Retry                  int                  `json:"retry"`
	RateLimitPerSecond     int                  `json:"rate_limit_per_second"`
	CircuitBreaker         CircuitBreakerConfig `json:"circuit_breaker"`
	StoreRawOutput         bool                 `json:"store_raw_output"`
	RequireConfiguredModel bool                 `json:"require_configured_model"`
}

type SafetyConfig struct {
	RedactPrompts      bool     `json:"redact_prompts"`
	StoreRequestHash   bool     `json:"store_request_hash"`
	MaxErrorChars      int      `json:"max_error_chars"`
	BlockedPhrases     []string `json:"blocked_phrases,omitempty"`
	AllowSystemPrompts bool     `json:"allow_system_prompts"`
}

type CostModelConfig struct {
	InputPer1K       float64 `json:"input_per_1k"`
	OutputPer1K      float64 `json:"output_per_1k"`
	CacheWritePer1K  float64 `json:"cache_write_per_1k,omitempty"`
	CacheReadPer1K   float64 `json:"cache_read_per_1k,omitempty"`
	ReasoningPer1K   float64 `json:"reasoning_per_1k,omitempty"`
	CurrencyOverride string  `json:"currency_override,omitempty"`
}

type CostConfig struct {
	Currency string                     `json:"currency"`
	Models   map[string]CostModelConfig `json:"models,omitempty"`
}

type Config struct {
	Enabled   bool             `json:"enabled"`
	Providers []ProviderConfig `json:"providers"`
	Model     ModelConfig      `json:"model"`
	Embedding EmbeddingConfig  `json:"embedding"`
	Policy    PolicyConfig     `json:"policy"`
	Safety    SafetyConfig     `json:"safety"`
	Cost      CostConfig       `json:"cost"`
}

type SecretStatus struct {
	Name    string `json:"name"`
	Env     string `json:"env"`
	Present bool   `json:"present"`
}

type ProviderStatus struct {
	Name                   string         `json:"name"`
	Type                   string         `json:"type"`
	Enabled                bool           `json:"enabled"`
	Configured             bool           `json:"configured"`
	BaseURL                string         `json:"base_url,omitempty"`
	APIMode                string         `json:"api_mode,omitempty"`
	Secrets                []SecretStatus `json:"secrets,omitempty"`
	SupportsAuthentication bool           `json:"supports_authentication"`
}

type StatusResponse struct {
	Enabled        bool             `json:"enabled"`
	Ready          bool             `json:"ready"`
	Configured     bool             `json:"configured"`
	Primary        string           `json:"primary,omitempty"`
	Fallbacks      []string         `json:"fallbacks,omitempty"`
	EmbeddingModel string           `json:"embedding_model,omitempty"`
	Providers      []ProviderStatus `json:"providers"`
	AllowedActions []string         `json:"allowed_actions"`
	Policy         PolicyConfig     `json:"policy"`
	Safety         SafetyConfig     `json:"safety"`
	Stats          StatsSnapshot    `json:"stats"`
	Message        string           `json:"message,omitempty"`
}

type ProviderHealth struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Healthy   bool   `json:"healthy"`
	Error     string `json:"error,omitempty"`
	CheckedAt int64  `json:"checked_at"`
}

type StatsSnapshot struct {
	TotalRequests      int64 `json:"total_requests"`
	SuccessRequests    int64 `json:"success_requests"`
	FailedRequests     int64 `json:"failed_requests"`
	InputTokens        int64 `json:"input_tokens"`
	OutputTokens       int64 `json:"output_tokens"`
	TotalTokens        int64 `json:"total_tokens"`
	CacheCreatedTokens int64 `json:"cache_created_tokens"`
	CacheReadTokens    int64 `json:"cache_read_tokens"`
}
