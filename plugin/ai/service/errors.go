package service

import "errors"

var (
	ErrAIDisabled     = errors.New("AI service is disabled")
	ErrAIUnconfigured = errors.New("AI provider is not configured")
	ErrPolicyDenied   = errors.New("AI request is denied by policy")
	ErrActionNotFound = errors.New("AI action is not registered")
)
