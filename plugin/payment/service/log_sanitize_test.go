package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizePaymentLogTextMasksJSONPayloads(t *testing.T) {
	raw := `{"order_id":"ord_1","client_secret":"pi_secret","nested":{"authorization":"Bearer token-value","amount":100},"items":[{"card_number":"4242424242424242"}]}`

	sanitized := sanitizePaymentLogText(raw)

	var payload map[string]any
	if err := json.Unmarshal([]byte(sanitized), &payload); err != nil {
		t.Fatalf("expected sanitized payload to remain valid JSON: %v", err)
	}

	if payload["client_secret"] != paymentLogRedactedValue {
		t.Fatalf("expected client_secret to be redacted, got %#v", payload["client_secret"])
	}
	if strings.Contains(sanitized, "pi_secret") || strings.Contains(sanitized, "token-value") || strings.Contains(sanitized, "424242") {
		t.Fatalf("expected sensitive values to be removed, got %s", sanitized)
	}
	if !strings.Contains(sanitized, `"amount":100`) {
		t.Fatalf("expected non-sensitive fields to remain, got %s", sanitized)
	}
}

func TestSanitizePaymentLogTextMasksFormAndHeaderPayloads(t *testing.T) {
	raw := "authorization=Bearer abc.def; account=user@example.com&amount=99&signature=signed-value"

	sanitized := sanitizePaymentLogText(raw)

	for _, secret := range []string{"abc.def", "user@example.com", "signed-value"} {
		if strings.Contains(sanitized, secret) {
			t.Fatalf("expected %q to be redacted from %q", secret, sanitized)
		}
	}
	if !strings.Contains(sanitized, "amount=99") {
		t.Fatalf("expected non-sensitive values to remain, got %s", sanitized)
	}
}

func TestSanitizePaymentLogMetadataMasksNestedValues(t *testing.T) {
	metadata := map[string]any{
		"provider": "stripe",
		"headers": map[string]any{
			"Stripe-Signature": "secret-signature",
		},
		"events": []any{
			map[string]any{"refresh_token": "refresh-secret"},
		},
	}

	sanitized := sanitizePaymentLogMetadata(metadata)

	if sanitized["provider"] != "stripe" {
		t.Fatalf("expected provider to remain, got %#v", sanitized["provider"])
	}
	encoded, _ := json.Marshal(sanitized)
	if strings.Contains(string(encoded), "secret-signature") || strings.Contains(string(encoded), "refresh-secret") {
		t.Fatalf("expected nested secrets to be redacted, got %s", encoded)
	}
}
