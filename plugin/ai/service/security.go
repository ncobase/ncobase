package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"ncobase/plugin/ai/structs"
	"sort"
	"strings"

	"github.com/ncobase/deebus/providers"
)

const redactedValue = "[REDACTED]"

var sensitiveMetadataKeys = []string{
	"api_key",
	"apikey",
	"authorization",
	"bearer_token",
	"client_secret",
	"cookie",
	"credential",
	"password",
	"private_key",
	"refresh_token",
	"secret",
	"secret_key",
	"signature",
	"token",
}

func hashAny(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		raw = []byte(fmt.Sprintf("%#v", value))
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func sanitizeError(err error, max int) (string, string) {
	if err == nil {
		return "", ""
	}
	code := "provider_error"
	var providerErr *providers.ProviderError
	if errors.As(err, &providerErr) {
		code = string(providerErr.Type)
	}
	message := strings.TrimSpace(err.Error())
	if max <= 0 {
		max = 500
	}
	if len(message) > max {
		message = message[:max] + "..."
	}
	return code, message
}

func sanitizeMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}
	clean := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if isSensitiveKey(key) {
			clean[key] = redactedValue
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			clean[key] = sanitizeMetadata(typed)
		default:
			clean[key] = value
		}
	}
	return clean
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	for _, sensitive := range sensitiveMetadataKeys {
		candidate := strings.ReplaceAll(strings.ReplaceAll(sensitive, "_", ""), "-", "")
		if normalized == candidate || strings.Contains(normalized, candidate) {
			return true
		}
	}
	return false
}

func estimateCost(cfg *structs.Config, model string, inputTokens, outputTokens, cacheCreated, cacheRead, reasoningTokens int) (float64, string) {
	if cfg == nil {
		return 0, ""
	}
	currency := cfg.Cost.Currency
	if currency == "" {
		currency = "USD"
	}
	price, ok := cfg.Cost.Models[model]
	if !ok {
		return 0, currency
	}
	if price.CurrencyOverride != "" {
		currency = price.CurrencyOverride
	}
	cost := (float64(inputTokens) / 1000 * price.InputPer1K) +
		(float64(outputTokens) / 1000 * price.OutputPer1K) +
		(float64(cacheCreated) / 1000 * price.CacheWritePer1K) +
		(float64(cacheRead) / 1000 * price.CacheReadPer1K) +
		(float64(reasoningTokens) / 1000 * price.ReasoningPer1K)
	return cost, currency
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(strings.ToLower(value))
		if trimmed != "" {
			result[trimmed] = struct{}{}
		}
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
