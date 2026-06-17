package service

import (
	"context"
	"testing"
)

type fakeOptionService map[string]map[string]any

func (f fakeOptionService) GetObjectByName(_ context.Context, name string) (map[string]any, error) {
	if value, ok := f[name]; ok {
		return value, nil
	}
	return nil, ErrAIUnconfigured
}

func TestConfigProviderMergesRuntimeOptions(t *testing.T) {
	provider := NewSystemOptionConfigProvider(fakeOptionService{
		optionAIProvider: {
			"enabled": true,
			"providers": []map[string]any{
				{
					"name":        "openai",
					"type":        "openai",
					"enabled":     true,
					"base_url":    "https://api.openai.com",
					"api_key_env": "OPENAI_API_KEY",
				},
			},
		},
		optionAIModel: {
			"primary":                   "openai/gpt-4o",
			"default_max_output_tokens": 2048,
			"default_temperature":       0.1,
		},
		optionAIPolicy: {
			"enabled":                true,
			"allowed_actions":        []any{"content.summary"},
			"max_prompt_chars":       1000,
			"max_messages":           8,
			"max_output_tokens":      512,
			"timeout_seconds":        15,
			"retry":                  1,
			"allowed_provider_types": []any{"openai"},
		},
	})

	cfg, err := provider.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if !cfg.Enabled {
		t.Fatal("expected AI provider config to be enabled")
	}
	if cfg.Model.Primary != "openai/gpt-4o" {
		t.Fatalf("primary model = %q", cfg.Model.Primary)
	}
	if cfg.Policy.MaxPromptChars != 1000 {
		t.Fatalf("max prompt chars = %d", cfg.Policy.MaxPromptChars)
	}
	if len(cfg.Policy.AllowedActions) != 1 || cfg.Policy.AllowedActions[0] != "content.summary" {
		t.Fatalf("allowed actions = %#v", cfg.Policy.AllowedActions)
	}
}

func TestProviderStatusDoesNotExposeSecretValue(t *testing.T) {
	t.Setenv("NCOBASE_AI_TEST_KEY", "secret-value")
	cfg := defaultConfig()
	cfg.Enabled = true
	cfg.Providers = cfg.Providers[:1]
	cfg.Providers[0].APIKeyEnv = "NCOBASE_AI_TEST_KEY"

	manager := newClientManager(NewDefaultConfigProvider())
	statuses := manager.providerStatuses(cfg)
	if len(statuses) != 1 {
		t.Fatalf("statuses len = %d", len(statuses))
	}
	if !statuses[0].Configured {
		t.Fatal("expected provider to be configured")
	}
	if len(statuses[0].Secrets) != 1 || !statuses[0].Secrets[0].Present {
		t.Fatalf("secret status = %#v", statuses[0].Secrets)
	}
	if statuses[0].Secrets[0].Env != "NCOBASE_AI_TEST_KEY" {
		t.Fatalf("secret env = %q", statuses[0].Secrets[0].Env)
	}
}
