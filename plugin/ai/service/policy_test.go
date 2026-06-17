package service

import (
	"strings"
	"testing"

	"ncobase/plugin/ai/structs"
)

func TestValidateCompletionPolicyRejectsOversizedPrompt(t *testing.T) {
	cfg := defaultConfig()
	cfg.Enabled = true
	cfg.Policy.MaxPromptChars = 12

	svc := &aiService{}
	err := svc.validateCompletionPolicy(cfg, &structs.CompleteRequest{Prompt: strings.Repeat("a", 13)})
	if err == nil || !strings.Contains(err.Error(), "prompt length") {
		t.Fatalf("expected prompt length error, got %v", err)
	}
}

func TestValidateCompletionPolicyRejectsDisallowedAction(t *testing.T) {
	cfg := defaultConfig()
	cfg.Enabled = true
	cfg.Policy.AllowedActions = []string{"content.summary"}

	svc := &aiService{}
	err := svc.validateCompletionPolicy(cfg, &structs.CompleteRequest{
		Action: "payment.summary",
		Prompt: "Summarize this payment context.",
	})
	if err == nil || !strings.Contains(err.Error(), "action payment.summary") {
		t.Fatalf("expected disallowed action error, got %v", err)
	}
}
