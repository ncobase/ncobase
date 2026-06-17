package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"ncobase/plugin/ai/structs"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ncobase/deebus"
)

type clientManager struct {
	config ConfigProvider
	mu     sync.RWMutex
	cache  map[string]*deebus.Client
}

func newClientManager(config ConfigProvider) *clientManager {
	if config == nil {
		config = NewDefaultConfigProvider()
	}
	return &clientManager{
		config: config,
		cache:  make(map[string]*deebus.Client),
	}
}

func (m *clientManager) status(ctx context.Context) (*structs.StatusResponse, *structs.Config, error) {
	cfg, err := m.config.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	statuses := m.providerStatuses(cfg)
	configuredProviders := make(map[string]struct{})
	for _, status := range statuses {
		if status.Enabled && status.Configured {
			configuredProviders[status.Name] = struct{}{}
		}
	}

	primaryProvider, _, _ := parseProviderModel(cfg.Model.Primary)
	_, primaryConfigured := configuredProviders[primaryProvider]
	ready := cfg.Enabled && cfg.Policy.Enabled && primaryConfigured
	message := ""
	switch {
	case !cfg.Enabled:
		message = "AI service is disabled by ai.provider."
	case !cfg.Policy.Enabled:
		message = "AI policy is disabled by ai.policy."
	case !primaryConfigured:
		message = "Primary AI provider is not configured or its secret environment variables are missing."
	}

	resp := &structs.StatusResponse{
		Enabled:        cfg.Enabled,
		Ready:          ready,
		Configured:     len(configuredProviders) > 0,
		Primary:        cfg.Model.Primary,
		Fallbacks:      cfg.Model.Fallbacks,
		EmbeddingModel: cfg.Embedding.Model,
		Providers:      statuses,
		AllowedActions: cfg.Policy.AllowedActions,
		Policy:         cfg.Policy,
		Safety:         cfg.Safety,
		Message:        message,
	}

	if ready {
		if client, err := m.get(ctx, cfg, "", nil); err == nil && client != nil {
			resp.Stats = statsSnapshot(client)
		}
	}

	return resp, cfg, nil
}

func (m *clientManager) get(ctx context.Context, cfg *structs.Config, primaryOverride string, fallbackOverride []string) (*deebus.Client, error) {
	if cfg == nil {
		var err error
		cfg, err = m.config.Get(ctx)
		if err != nil {
			return nil, err
		}
	}
	if !cfg.Enabled || !cfg.Policy.Enabled {
		return nil, ErrAIDisabled
	}

	deebusConfig, err := m.deebusConfig(cfg, primaryOverride, fallbackOverride)
	if err != nil {
		return nil, err
	}

	fingerprint := configFingerprint(deebusConfig)
	m.mu.RLock()
	if client, ok := m.cache[fingerprint]; ok {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	client, err := deebus.NewClient(deebusConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIUnconfigured, err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if cached, ok := m.cache[fingerprint]; ok {
		return cached, nil
	}
	m.cache[fingerprint] = client
	return client, nil
}

func (m *clientManager) deebusConfig(cfg *structs.Config, primaryOverride string, fallbackOverride []string) (deebus.Config, error) {
	allowedProviderTypes := stringSet(cfg.Policy.AllowedProviderTypes)
	providers := make(map[string]deebus.ProviderConfig)
	for _, provider := range cfg.Providers {
		if !provider.Enabled {
			continue
		}
		if _, ok := allowedProviderTypes[provider.Type]; !ok && len(allowedProviderTypes) > 0 {
			continue
		}
		resolved, configured := resolveProviderConfig(provider)
		if !configured {
			continue
		}
		providers[provider.Name] = resolved
	}

	primary := strings.TrimSpace(primaryOverride)
	if primary == "" {
		primary = cfg.Model.Primary
	}
	if primary == "" {
		return deebus.Config{}, fmt.Errorf("%w: primary model is empty", ErrAIUnconfigured)
	}

	primaryProvider, _, err := parseProviderModel(primary)
	if err != nil {
		return deebus.Config{}, err
	}
	if _, ok := providers[primaryProvider]; !ok {
		return deebus.Config{}, fmt.Errorf("%w: provider %s is not configured", ErrAIUnconfigured, primaryProvider)
	}

	fallbacks := cfg.Model.Fallbacks
	if fallbackOverride != nil {
		fallbacks = fallbackOverride
	}
	filteredFallbacks := make([]string, 0, len(fallbacks))
	for _, fallback := range fallbacks {
		providerName, _, err := parseProviderModel(fallback)
		if err != nil {
			continue
		}
		if _, ok := providers[providerName]; ok {
			filteredFallbacks = append(filteredFallbacks, fallback)
		}
	}

	return deebus.Config{
		Providers: providers,
		Primary:   primary,
		Fallbacks: filteredFallbacks,
		Timeout:   cfg.Policy.TimeoutSeconds,
		Retry:     cfg.Policy.Retry,
		RateLimit: cfg.Policy.RateLimitPerSecond,
		CircuitBreaker: deebus.CircuitBreakerConfig{
			MaxFailures:  cfg.Policy.CircuitBreaker.MaxFailures,
			ResetTimeout: cfg.Policy.CircuitBreaker.ResetSeconds,
		},
	}, nil
}

func (m *clientManager) providerStatuses(cfg *structs.Config) []structs.ProviderStatus {
	statuses := make([]structs.ProviderStatus, 0, len(cfg.Providers))
	for _, provider := range cfg.Providers {
		_, configured := resolveProviderConfig(provider)
		status := structs.ProviderStatus{
			Name:                   provider.Name,
			Type:                   provider.Type,
			Enabled:                provider.Enabled,
			Configured:             configured,
			BaseURL:                provider.BaseURL,
			APIMode:                provider.APIMode,
			Secrets:                providerSecrets(provider),
			SupportsAuthentication: provider.Type != "ollama",
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func resolveProviderConfig(provider structs.ProviderConfig) (deebus.ProviderConfig, bool) {
	headers := make(map[string]string, len(provider.Headers)+len(provider.HeaderEnvs))
	for key, value := range provider.Headers {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			headers[key] = value
		}
	}
	for headerName, envName := range provider.HeaderEnvs {
		if headerName == "" || envName == "" {
			continue
		}
		if value, ok := os.LookupEnv(envName); ok && strings.TrimSpace(value) != "" {
			headers[headerName] = value
		}
	}

	apiKey := ""
	if provider.APIKeyEnv != "" {
		apiKey = strings.TrimSpace(os.Getenv(provider.APIKeyEnv))
	}
	bearerToken := ""
	if provider.BearerTokenEnv != "" {
		bearerToken = strings.TrimSpace(os.Getenv(provider.BearerTokenEnv))
	}

	configured := provider.Type == "ollama" || apiKey != "" || bearerToken != "" || len(headers) > 0
	return deebus.ProviderConfig{
		Type:         provider.Type,
		APIKey:       apiKey,
		BearerToken:  bearerToken,
		BaseURL:      provider.BaseURL,
		APIMode:      provider.APIMode,
		Headers:      headers,
		Organization: provider.Organization,
		Project:      provider.Project,
		UserProject:  provider.UserProject,
	}, configured
}

func providerSecrets(provider structs.ProviderConfig) []structs.SecretStatus {
	secrets := make([]structs.SecretStatus, 0, 2+len(provider.HeaderEnvs))
	if provider.APIKeyEnv != "" {
		_, present := os.LookupEnv(provider.APIKeyEnv)
		secrets = append(secrets, structs.SecretStatus{Name: "api_key", Env: provider.APIKeyEnv, Present: present})
	}
	if provider.BearerTokenEnv != "" {
		_, present := os.LookupEnv(provider.BearerTokenEnv)
		secrets = append(secrets, structs.SecretStatus{Name: "bearer_token", Env: provider.BearerTokenEnv, Present: present})
	}
	for headerName, envName := range provider.HeaderEnvs {
		_, present := os.LookupEnv(envName)
		secrets = append(secrets, structs.SecretStatus{Name: "header:" + headerName, Env: envName, Present: present})
	}
	return secrets
}

func statsSnapshot(client *deebus.Client) structs.StatsSnapshot {
	if client == nil || client.Stats == nil {
		return structs.StatsSnapshot{}
	}
	requests, inputTokens, outputTokens, success, failed := client.Stats.Get()
	return structs.StatsSnapshot{
		TotalRequests:      requests,
		SuccessRequests:    success,
		FailedRequests:     failed,
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
		TotalTokens:        client.Stats.TotalTokens.Load(),
		CacheCreatedTokens: client.Stats.CacheCreatedTokens.Load(),
		CacheReadTokens:    client.Stats.CacheReadTokens.Load(),
	}
}

func configFingerprint(cfg deebus.Config) string {
	raw, err := json.Marshal(cfg)
	if err != nil {
		raw = []byte(fmt.Sprintf("%#v", cfg))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func parseProviderModel(model string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(model), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(model, "..") {
		return "", "", fmt.Errorf("invalid model %q: expected provider/model", model)
	}
	return parts[0], parts[1], nil
}

func withPolicyTimeout(ctx context.Context, cfg *structs.Config) (context.Context, context.CancelFunc) {
	timeout := 45 * time.Second
	if cfg != nil && cfg.Policy.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.Policy.TimeoutSeconds) * time.Second
	}
	return context.WithTimeout(ctx, timeout)
}
