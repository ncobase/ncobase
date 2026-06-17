package service

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/plugin/ai/structs"
	"sort"
	"strings"
)

const (
	optionAIProvider  = "ai.provider"
	optionAIModel     = "ai.model"
	optionAIPolicy    = "ai.policy"
	optionAISafety    = "ai.safety"
	optionAICost      = "ai.cost"
	optionAIEmbedding = "ai.embedding"
)

type ConfigProvider interface {
	Get(ctx context.Context) (*structs.Config, error)
}

type OptionObjectService interface {
	GetObjectByName(ctx context.Context, name string) (map[string]any, error)
}

type systemOptionConfigProvider struct {
	option OptionObjectService
}

func NewSystemOptionConfigProvider(option OptionObjectService) ConfigProvider {
	return &systemOptionConfigProvider{option: option}
}

func NewDefaultConfigProvider() ConfigProvider {
	return &systemOptionConfigProvider{}
}

func (p *systemOptionConfigProvider) Get(ctx context.Context) (*structs.Config, error) {
	cfg := defaultConfig()
	if p == nil || p.option == nil {
		return cfg, nil
	}

	if values, err := p.option.GetObjectByName(ctx, optionAIProvider); err == nil {
		if v, ok := boolFromAny(values["enabled"]); ok {
			cfg.Enabled = v
		}
		if rawProviders, ok := values["providers"]; ok {
			var providers []structs.ProviderConfig
			if err := decodeAny(rawProviders, &providers); err == nil {
				cfg.Providers = providers
			}
		}
	}

	if values, err := p.option.GetObjectByName(ctx, optionAIModel); err == nil {
		_ = decodeMap(values, &cfg.Model)
	}
	if values, err := p.option.GetObjectByName(ctx, optionAIPolicy); err == nil {
		_ = decodeMap(values, &cfg.Policy)
	}
	if values, err := p.option.GetObjectByName(ctx, optionAISafety); err == nil {
		_ = decodeMap(values, &cfg.Safety)
	}
	if values, err := p.option.GetObjectByName(ctx, optionAICost); err == nil {
		_ = decodeMap(values, &cfg.Cost)
	}
	if values, err := p.option.GetObjectByName(ctx, optionAIEmbedding); err == nil {
		_ = decodeMap(values, &cfg.Embedding)
	}

	return normalizeConfig(cfg), nil
}

func defaultConfig() *structs.Config {
	return &structs.Config{
		Enabled: false,
		Providers: []structs.ProviderConfig{
			{
				Name:      "openai",
				Type:      "openai",
				Enabled:   true,
				BaseURL:   "https://api.openai.com",
				APIKeyEnv: "OPENAI_API_KEY",
			},
			{
				Name:      "anthropic",
				Type:      "anthropic",
				Enabled:   false,
				BaseURL:   "https://api.anthropic.com",
				APIKeyEnv: "ANTHROPIC_API_KEY",
			},
			{
				Name:      "gemini",
				Type:      "gemini",
				Enabled:   false,
				BaseURL:   "https://generativelanguage.googleapis.com",
				APIKeyEnv: "GEMINI_API_KEY",
			},
			{
				Name:    "ollama",
				Type:    "ollama",
				Enabled: false,
				BaseURL: "http://localhost:11434",
			},
			{
				Name:      "cohere",
				Type:      "cohere",
				Enabled:   false,
				BaseURL:   "https://api.cohere.ai",
				APIKeyEnv: "COHERE_API_KEY",
			},
		},
		Model: structs.ModelConfig{
			Primary:                "openai/gpt-4o-mini",
			Fallbacks:              []string{},
			DefaultMaxOutputTokens: 1024,
			DefaultTemperature:     0.2,
		},
		Embedding: structs.EmbeddingConfig{
			Enabled:   true,
			Model:     "openai/text-embedding-3-small",
			InputType: "search_document",
			MaxItems:  64,
		},
		Policy: structs.PolicyConfig{
			Enabled:                true,
			AllowedActions:         defaultActionKeys(),
			AllowedProviderTypes:   []string{"openai", "anthropic", "gemini", "ollama", "cohere"},
			MaxPromptChars:         20000,
			MaxMessages:            32,
			MaxInputItems:          64,
			MaxOutputTokens:        4096,
			TimeoutSeconds:         45,
			Retry:                  2,
			RateLimitPerSecond:     0,
			CircuitBreaker:         structs.CircuitBreakerConfig{MaxFailures: 5, ResetSeconds: 60},
			StoreRawOutput:         false,
			RequireConfiguredModel: true,
		},
		Safety: structs.SafetyConfig{
			RedactPrompts:      true,
			StoreRequestHash:   true,
			MaxErrorChars:      500,
			BlockedPhrases:     []string{},
			AllowSystemPrompts: true,
		},
		Cost: structs.CostConfig{
			Currency: "USD",
			Models:   map[string]structs.CostModelConfig{},
		},
	}
}

func normalizeConfig(cfg *structs.Config) *structs.Config {
	defaults := defaultConfig()
	if cfg == nil {
		return defaults
	}

	if len(cfg.Providers) == 0 {
		cfg.Providers = defaults.Providers
	}
	for i := range cfg.Providers {
		cfg.Providers[i].Name = strings.TrimSpace(strings.ToLower(cfg.Providers[i].Name))
		cfg.Providers[i].Type = strings.TrimSpace(strings.ToLower(cfg.Providers[i].Type))
		cfg.Providers[i].BaseURL = strings.TrimSpace(cfg.Providers[i].BaseURL)
		cfg.Providers[i].APIKeyEnv = strings.TrimSpace(cfg.Providers[i].APIKeyEnv)
		cfg.Providers[i].BearerTokenEnv = strings.TrimSpace(cfg.Providers[i].BearerTokenEnv)
		if cfg.Providers[i].Name == "" {
			cfg.Providers[i].Name = cfg.Providers[i].Type
		}
	}

	if cfg.Model.Primary == "" {
		cfg.Model.Primary = defaults.Model.Primary
	}
	if cfg.Model.DefaultMaxOutputTokens <= 0 {
		cfg.Model.DefaultMaxOutputTokens = defaults.Model.DefaultMaxOutputTokens
	}
	if cfg.Model.DefaultTemperature < 0 {
		cfg.Model.DefaultTemperature = defaults.Model.DefaultTemperature
	}

	if cfg.Embedding.Model == "" {
		cfg.Embedding.Model = defaults.Embedding.Model
	}
	if cfg.Embedding.InputType == "" {
		cfg.Embedding.InputType = defaults.Embedding.InputType
	}
	if cfg.Embedding.MaxItems <= 0 {
		cfg.Embedding.MaxItems = defaults.Embedding.MaxItems
	}

	if len(cfg.Policy.AllowedActions) == 0 {
		cfg.Policy.AllowedActions = defaults.Policy.AllowedActions
	}
	if len(cfg.Policy.AllowedProviderTypes) == 0 {
		cfg.Policy.AllowedProviderTypes = defaults.Policy.AllowedProviderTypes
	}
	if cfg.Policy.MaxPromptChars <= 0 {
		cfg.Policy.MaxPromptChars = defaults.Policy.MaxPromptChars
	}
	if cfg.Policy.MaxMessages <= 0 {
		cfg.Policy.MaxMessages = defaults.Policy.MaxMessages
	}
	if cfg.Policy.MaxInputItems <= 0 {
		cfg.Policy.MaxInputItems = defaults.Policy.MaxInputItems
	}
	if cfg.Policy.MaxOutputTokens <= 0 {
		cfg.Policy.MaxOutputTokens = defaults.Policy.MaxOutputTokens
	}
	if cfg.Policy.TimeoutSeconds <= 0 {
		cfg.Policy.TimeoutSeconds = defaults.Policy.TimeoutSeconds
	}
	if cfg.Policy.CircuitBreaker.ResetSeconds <= 0 {
		cfg.Policy.CircuitBreaker.ResetSeconds = defaults.Policy.CircuitBreaker.ResetSeconds
	}
	if cfg.Safety.MaxErrorChars <= 0 {
		cfg.Safety.MaxErrorChars = defaults.Safety.MaxErrorChars
	}
	if cfg.Cost.Currency == "" {
		cfg.Cost.Currency = defaults.Cost.Currency
	}
	if cfg.Cost.Models == nil {
		cfg.Cost.Models = map[string]structs.CostModelConfig{}
	}

	sort.Strings(cfg.Policy.AllowedActions)
	sort.Strings(cfg.Policy.AllowedProviderTypes)
	return cfg
}

func decodeMap(values map[string]any, target any) error {
	if values == nil {
		return nil
	}
	return decodeAny(values, target)
}

func decodeAny(value any, target any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal option: %w", err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode option: %w", err)
	}
	return nil
}

func boolFromAny(value any) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "on":
			return true, true
		case "false", "0", "no", "off":
			return false, true
		default:
			return false, false
		}
	default:
		return false, false
	}
}
