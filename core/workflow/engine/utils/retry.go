package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"time"
)

// RetryableFunc represents a function that can be retried
type RetryableFunc func(ctx context.Context) error

// RetryWithBackoff executes function with exponential backoff retry
func RetryWithBackoff(ctx context.Context, f RetryableFunc, cfg *config.RetryConfig) error {
	if cfg == nil {
		cfg = config.DefaultRetryConfig()
	}

	startTime := time.Now()
	attempt := 0

	for attempt < cfg.MaxAttempts {
		// Check max duration
		if cfg.MaxDuration > 0 && time.Since(startTime) > cfg.MaxDuration {
			err := types.NewError(types.ErrTimeout, "max retry duration exceeded", nil)
			if cfg.OnMaxAttemptsReached != nil {
				cfg.OnMaxAttemptsReached(err)
			}
			return err
		}

		// Execute function
		err := f(ctx)
		if err == nil {
			if cfg.OnSuccess != nil {
				cfg.OnSuccess(attempt)
			}
			return nil
		}

		// Check if error is retryable
		if !isRetryableError(err, cfg.RetryableErrors) {
			return err
		}

		attempt++
		if attempt >= cfg.MaxAttempts {
			if cfg.OnMaxAttemptsReached != nil {
				cfg.OnMaxAttemptsReached(err)
			}
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// Notify retry callback
		if cfg.OnRetry != nil {
			cfg.OnRetry(attempt, err)
		}

		// Calculate delay
		delay := calculateDelay(attempt, cfg)

		// Wait for next attempt or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return nil
}

// isRetryableError checks if the error should be retried
func isRetryableError(err error, retryableErrors []error) bool {
	if err == nil {
		return false
	}

	// Check against retryable errors list
	for _, retryableErr := range retryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	return false
}

// calculateDelay calculates the delay for next retry
func calculateDelay(attempt int, cfg *config.RetryConfig) time.Duration {
	// Calculate base delay with multiplier
	delay := float64(cfg.InitialInterval) * math.Pow(cfg.Multiplier, float64(attempt-1))

	// Cap at max interval
	if delay > float64(cfg.MaxInterval) {
		delay = float64(cfg.MaxInterval)
	}

	// Add jitter if enabled
	if cfg.Jitter {
		delay = delay * (0.5 + rand.Float64())
	}

	return time.Duration(delay)
}

// // AsyncRetry represents an async retry operation
// type AsyncRetry struct {
// 	ID        string
// 	StartTime time.Time
// 	Attempts  int
// 	LastError error
// 	Done      chan struct{}
// 	Result    chan error
// }
//
// // RetryAsync executes function asynchronously with retry
// func RetryAsync(ctx context.Context, f RetryableFunc, cfg *config.RetryConfig) *AsyncRetry {
// 	retry := &AsyncRetry{
// 		ID:        uuid.New().String(),
// 		StartTime: time.Now(),
// 		Done:      make(chan struct{}),
// 		Result:    make(chan error, 1),
// 	}
//
// 	go func() {
// 		defer close(retry.Done)
// 		defer close(retry.Result)
//
// 		err := RetryWithBackoff(ctx, f, cfg)
// 		retry.Result <- err
// 	}()
//
// 	return retry
// }
