package config

import (
	"ncobase/core/workflow/engine/types"
	"time"
)

// RetryConfig defines the retry behavior
type RetryConfig struct {
	// Maximum number of retry attempts
	MaxAttempts int `json:"max_attempts"`

	// Initial delay between retries
	InitialInterval time.Duration `json:"initial_interval"`

	// Maximum delay between retries
	MaxInterval time.Duration `json:"max_interval"`

	// Multiplier for the delay after each retry
	Multiplier float64 `json:"multiplier"`

	// Jitter determines if random jitter should be added to delays
	Jitter bool `json:"jitter"`

	// RetryableErrors defines which errors should be retried
	RetryableErrors []error `json:"retryable_errors"`

	// Maximum total duration for all retries
	MaxDuration time.Duration `json:"max_duration"`

	// OnRetry is called before each retry attempt
	OnRetry func(attempt int, err error)

	// OnMaxAttemptsReached is called when max attempts are reached
	OnMaxAttemptsReached func(err error)

	// OnSuccess is called when the operation succeeds after retries
	OnSuccess func(attempt int)
}

// DefaultRetryConfig returns a default retry config
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:     3,
		InitialInterval: time.Second,
		MaxInterval:     time.Second * 30,
		Multiplier:      2.0,
		Jitter:          true,
		MaxDuration:     time.Minute * 5,
		RetryableErrors: []error{
			types.NewError(types.ErrTimeout, "timeout", nil),
			types.NewError(types.ErrSystem, "system error", nil),
			types.NewError(types.ErrNetwork, "network error", nil),
			types.NewError(types.ErrUnknown, "unknown error", nil),
			types.NewError(types.ErrServiceUnavailable, "service unavailable", nil),
			types.NewError(types.ErrInvalidParam, "invalid parameter", nil),
		},
	}
}
