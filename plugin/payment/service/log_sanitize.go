package service

import (
	"encoding/json"
	"regexp"
	"strings"
)

const paymentLogRedactedValue = "[REDACTED]"

var (
	paymentLogSensitiveKeys = []string{
		"access_token",
		"account",
		"api_key",
		"apikey",
		"authorization",
		"card",
		"card_number",
		"client_secret",
		"cookie",
		"credential",
		"cvc",
		"cvv",
		"password",
		"private_key",
		"refresh_token",
		"secret",
		"secret_key",
		"sign",
		"signature",
		"token",
		"webhook_secret",
	}
	paymentLogAuthorizationPattern = regexp.MustCompile(`(?i)(authorization\s*[:=]\s*)(bearer\s+)?([^\s,;]+)`)
	paymentLogPairPattern          = regexp.MustCompile(`(?i)(\b(?:access_token|account|api_key|apikey|authorization|card|card_number|client_secret|cookie|credential|cvc|cvv|password|private_key|refresh_token|secret|secret_key|sign|signature|token|webhook_secret)\b\s*[:=]\s*)("[^"]*"|'[^']*'|[^,\s;&}]+)`)
)

func sanitizePaymentLogText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return value
	}

	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		sanitized := sanitizePaymentLogValue(payload)
		if encoded, err := json.Marshal(sanitized); err == nil {
			return string(encoded)
		}
	}

	sanitized := paymentLogAuthorizationPattern.ReplaceAllString(trimmed, `${1}${2}`+paymentLogRedactedValue)
	return paymentLogPairPattern.ReplaceAllString(sanitized, `${1}`+paymentLogRedactedValue)
}

func sanitizePaymentLogMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return nil
	}

	sanitized, _ := sanitizePaymentLogValue(metadata).(map[string]any)
	return sanitized
}

func sanitizePaymentLogValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			if isPaymentLogSensitiveKey(key) {
				result[key] = paymentLogRedactedValue
				continue
			}
			result[key] = sanitizePaymentLogValue(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, child := range typed {
			result[i] = sanitizePaymentLogValue(child)
		}
		return result
	case string:
		return sanitizePaymentLogText(typed)
	default:
		return typed
	}
}

func isPaymentLogSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	for _, sensitive := range paymentLogSensitiveKeys {
		candidate := strings.ReplaceAll(strings.ReplaceAll(sensitive, "_", ""), "-", "")
		if normalized == candidate || strings.Contains(normalized, candidate) {
			return true
		}
	}
	return false
}
