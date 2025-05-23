package config

import (
	"github.com/spf13/viper"
)

// Config holds configuration for the system module
type Config struct {
	Initialization *InitConfig `json:"initialization"`
	Security       *SecConfig  `json:"security"`
	Logging        *LogConfig  `json:"logging"`
}

// InitConfig holds configuration for system initialization
type InitConfig struct {
	AllowReinitialization bool   `json:"allow_reinitialization"`
	InitToken             string `json:"init_token"`
	TokenExpiry           string `json:"token_expiry"`
	PersistState          bool   `json:"persist_state"`
}

// SecConfig holds security configuration for the system
type SecConfig struct {
	DefaultPasswordPolicy *PasswordPolicy `json:"default_password_policy"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	EnableDebug bool `json:"enable_debug"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength          int  `json:"min_length"`
	RequireUppercase   bool `json:"require_uppercase"`
	RequireLowercase   bool `json:"require_lowercase"`
	RequireDigits      bool `json:"require_digits"`
	RequireSpecial     bool `json:"require_special"`
	ExpirePasswordDays int  `json:"expire_password_days"`
}

// GetDefaultConfig returns the default configuration for the system module
func GetDefaultConfig() *Config {
	return &Config{
		Initialization: &InitConfig{
			AllowReinitialization: false,
			InitToken:             "Ac231", // Empty means no token required
			TokenExpiry:           "24h",
			PersistState:          true,
		},
		Security: &SecConfig{
			DefaultPasswordPolicy: &PasswordPolicy{
				MinLength:          8,
				RequireUppercase:   true,
				RequireLowercase:   true,
				RequireDigits:      true,
				RequireSpecial:     false,
				ExpirePasswordDays: 90,
			},
		},
		Logging: &LogConfig{
			EnableDebug: false,
		},
	}
}

// GetConfigFromFile loads configuration from Viper into the system configuration
func GetConfigFromFile(config *Config, viper *viper.Viper) *Config {
	if viper == nil {
		return config
	}

	// If config is nil, start with default config
	if config == nil {
		config = GetDefaultConfig()
	}

	// Initialize config sections if needed
	if config.Initialization == nil {
		config.Initialization = &InitConfig{}
	}

	if config.Security == nil {
		config.Security = &SecConfig{}
	}

	if config.Logging == nil {
		config.Logging = &LogConfig{}
	}

	// Initialization settings
	if viper.IsSet("system.initialization.allow_reinitialization") {
		config.Initialization.AllowReinitialization = viper.GetBool("system.initialization.allow_reinitialization")
	}

	if viper.IsSet("system.initialization.init_token") {
		config.Initialization.InitToken = viper.GetString("system.initialization.init_token")
	}

	if viper.IsSet("system.initialization.token_expiry") {
		config.Initialization.TokenExpiry = viper.GetString("system.initialization.token_expiry")
	}

	if viper.IsSet("system.initialization.persist_state") {
		config.Initialization.PersistState = viper.GetBool("system.initialization.persist_state")
	}

	// Logging settings
	if viper.IsSet("system.logging.enable_debug") {
		config.Logging.EnableDebug = viper.GetBool("system.logging.enable_debug")
	}

	// Security settings
	if config.Security.DefaultPasswordPolicy == nil {
		config.Security.DefaultPasswordPolicy = &PasswordPolicy{}
	}

	if viper.IsSet("system.security.default_password_policy.min_length") {
		config.Security.DefaultPasswordPolicy.MinLength = viper.GetInt("system.security.default_password_policy.min_length")
	}

	if viper.IsSet("system.security.default_password_policy.require_uppercase") {
		config.Security.DefaultPasswordPolicy.RequireUppercase = viper.GetBool("system.security.default_password_policy.require_uppercase")
	}

	if viper.IsSet("system.security.default_password_policy.require_lowercase") {
		config.Security.DefaultPasswordPolicy.RequireLowercase = viper.GetBool("system.security.default_password_policy.require_lowercase")
	}

	if viper.IsSet("system.security.default_password_policy.require_digits") {
		config.Security.DefaultPasswordPolicy.RequireDigits = viper.GetBool("system.security.default_password_policy.require_digits")
	}

	if viper.IsSet("system.security.default_password_policy.require_special") {
		config.Security.DefaultPasswordPolicy.RequireSpecial = viper.GetBool("system.security.default_password_policy.require_special")
	}

	if viper.IsSet("system.security.default_password_policy.expire_password_days") {
		config.Security.DefaultPasswordPolicy.ExpirePasswordDays = viper.GetInt("system.security.default_password_policy.expire_password_days")
	}

	return config
}
