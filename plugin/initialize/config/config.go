package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Initialization *InitConfig `json:"initialization"`
	Security       *SecConfig  `json:"security"`
	Logging        *LogConfig  `json:"logging"`
}

type InitConfig struct {
	AllowReinitialization bool   `json:"allow_reinitialization"`
	InitToken             string `json:"init_token"`
	TokenExpiry           string `json:"token_expiry"`
	PersistState          bool   `json:"persist_state"`
	DataMode              string `json:"data_mode"` // "enterprise", "company", "website"
}

type SecConfig struct {
	DefaultPasswordPolicy *PasswordPolicy `json:"default_password_policy"`
}

type LogConfig struct {
	EnableDebug bool `json:"enable_debug"`
}

type PasswordPolicy struct {
	MinLength          int  `json:"min_length"`
	RequireUppercase   bool `json:"require_uppercase"`
	RequireLowercase   bool `json:"require_lowercase"`
	RequireDigits      bool `json:"require_digits"`
	RequireSpecial     bool `json:"require_special"`
	ExpirePasswordDays int  `json:"expire_password_days"`
}

func GetDefaultConfig() *Config {
	return &Config{
		Initialization: &InitConfig{
			AllowReinitialization: false,
			InitToken:             "Ac231",
			TokenExpiry:           "24h",
			PersistState:          true,
			DataMode:              "website",
		},
		Security: &SecConfig{
			DefaultPasswordPolicy: &PasswordPolicy{
				MinLength:          8,
				RequireUppercase:   false,
				RequireLowercase:   true,
				RequireDigits:      true,
				RequireSpecial:     false,
				ExpirePasswordDays: 0,
			},
		},
		Logging: &LogConfig{
			EnableDebug: false,
		},
	}
}

func GetConfigFromFile(config *Config, viper *viper.Viper) *Config {
	if viper == nil {
		return config
	}

	if config == nil {
		config = GetDefaultConfig()
	}

	if config.Initialization == nil {
		config.Initialization = &InitConfig{}
	}

	if config.Security == nil {
		config.Security = &SecConfig{}
	}

	if config.Logging == nil {
		config.Logging = &LogConfig{}
	}

	// Load configuration from viper
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

	if viper.IsSet("system.initialization.data_mode") {
		config.Initialization.DataMode = viper.GetString("system.initialization.data_mode")
	}

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
