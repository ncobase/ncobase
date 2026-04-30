package config

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
