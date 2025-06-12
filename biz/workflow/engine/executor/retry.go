package executor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"ncobase/workflow/engine/config"
	"ncobase/workflow/engine/metrics"
	"ncobase/workflow/engine/types"
	"ncobase/workflow/service"
	"sync"
	"sync/atomic"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// RetryExecutor handles retry operations
type RetryExecutor struct {
	*BaseExecutor

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	logger   logger.Logger

	// Runtime components
	metrics *metrics.Collector

	// Retry state tracking
	retryState    sync.Map // key -> *RetryState
	activeRetries atomic.Int32

	// Configuration
	policy *config.RetryConfig
	config *config.RetryExecutorConfig

	// Retry statistics
	stats *RetryStats

	// Runtime state
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// RetryInfo represents retry attempt information
type RetryInfo struct {
	Key         string
	Attempt     int
	LastAttempt time.Time
	NextAttempt time.Time
	Error       error
	BackoffTime time.Duration
}

// RetryState tracks retry state
type RetryState struct {
	LastAttempt     time.Time
	NextAttempt     time.Time
	AttemptCount    int
	LastError       error
	BackoffDuration time.Duration
	MaxAttempts     int
	mu              sync.RWMutex
}

// RetryStats tracks retry statistics
type RetryStats struct {
	RetryAttempts       int64
	RetrySuccesses      int64
	RetryFailures       int64
	LastRetryTime       time.Time
	AvgRetryLatency     time.Duration
	MaxRetryLatency     time.Duration
	ConsecutiveFailures int64
	SuccessRate         float64
	mu                  sync.RWMutex
}

// NewRetryExecutor creates a new retry executor
func NewRetryExecutor(svc *service.Service, em ext.ManagerInterface, cfg *config.Config) (*RetryExecutor, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &RetryExecutor{
		BaseExecutor: NewBaseExecutor(types.RetryExecutor, "Retry Executor", cfg),
		services:     svc,
		em:           em,
		config:       cfg.Executors.Retry,
		policy:       cfg.Executors.Retry.Policy,
		ctx:          ctx,
		cancel:       cancel,
		stats:        &RetryStats{},
	}

	// Initialize metrics if enabled
	if executor.config.MetricsEnabled {
		collector, err := metrics.NewCollector(cfg.Components.Metrics)
		if err != nil {
			return nil, fmt.Errorf("create metrics collector failed: %w", err)
		}
		executor.metrics = collector
	}

	return executor, nil
}

// Start starts the retry executor
func (e *RetryExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status() != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "executor not in pending state", nil)
	}

	// Initialize metrics if enabled
	if e.config.MetricsEnabled {
		if err := e.initMetrics(); err != nil {
			return fmt.Errorf("initialize metrics failed: %w", err)
		}
	}

	return nil
}

// Stop stops the retry executor
func (e *RetryExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status() != types.ExecutionActive {
		return nil
	}

	// Cancel context
	e.cancel()

	return nil
}

// ExecuteWithRetry executes with retry logic
func (e *RetryExecutor) ExecuteWithRetry(ctx context.Context, key string, fn func(context.Context) error) error {
	// Create retry context
	retryCtx := e.createRetryContext(ctx, key)
	defer retryCtx.cleanup()

	startTime := time.Now()

	for retryCtx.attempt < retryCtx.maxAttempts {
		// Check if max duration exceeded
		if e.policy.MaxDuration > 0 && time.Since(startTime) > e.policy.MaxDuration {
			return types.NewError(types.ErrTimeout, "max retry duration exceeded", nil)
		}

		// Execute function
		err := fn(retryCtx.ctx)
		if err == nil {
			e.recordSuccess(retryCtx)
			if e.policy.OnSuccess != nil {
				e.policy.OnSuccess(retryCtx.attempt)
			}
			return nil
		}

		// Handle retry
		if cont := e.handleRetry(retryCtx, err); !cont {
			return err
		}
	}

	// Handle max attempts reached
	if e.policy.OnMaxAttemptsReached != nil {
		e.policy.OnMaxAttemptsReached(retryCtx.lastError)
	}

	return fmt.Errorf("max retry attempts (%d) reached: %w", retryCtx.maxAttempts, retryCtx.lastError)
}

// GetRetryInfo gets retry information
func (e *RetryExecutor) GetRetryInfo(key string) (*RetryInfo, error) {
	state, ok := e.retryState.Load(key)
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "retry state not found", nil)
	}

	retryState := state.(*RetryState)
	retryState.mu.RLock()
	defer retryState.mu.RUnlock()

	return &RetryInfo{
		Key:         key,
		Attempt:     retryState.AttemptCount,
		LastAttempt: retryState.LastAttempt,
		NextAttempt: retryState.NextAttempt,
		Error:       retryState.LastError,
		BackoffTime: retryState.BackoffDuration,
	}, nil
}

// SetRetryPolicy sets retry policy
func (e *RetryExecutor) SetRetryPolicy(policy *config.RetryConfig) error {
	if policy == nil {
		return types.NewError(types.ErrInvalidParam, "policy is nil", nil)
	}

	e.mu.Lock()
	e.policy = policy
	e.mu.Unlock()

	return nil
}

// GetRetryPolicy gets retry policy
func (e *RetryExecutor) GetRetryPolicy() *config.RetryConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.policy
}

// GetRetryStats gets retry statistics
func (e *RetryExecutor) GetRetryStats() *RetryStats {
	e.stats.mu.RLock()
	defer e.stats.mu.RUnlock()

	// Return copy of stats
	return &RetryStats{
		RetryAttempts:       e.stats.RetryAttempts,
		RetrySuccesses:      e.stats.RetrySuccesses,
		RetryFailures:       e.stats.RetryFailures,
		LastRetryTime:       e.stats.LastRetryTime,
		AvgRetryLatency:     e.stats.AvgRetryLatency,
		MaxRetryLatency:     e.stats.MaxRetryLatency,
		ConsecutiveFailures: e.stats.ConsecutiveFailures,
		SuccessRate:         e.stats.SuccessRate,
	}
}

// Helper types and methods

type retryContext struct {
	ctx         context.Context
	cancel      context.CancelFunc
	key         string
	attempt     int
	maxAttempts int
	startTime   time.Time
	lastError   error
	metadata    map[string]any
	logger      logger.Logger
}

func (e *RetryExecutor) createRetryContext(ctx context.Context, key string) *retryContext {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	return &retryContext{
		ctx:         ctxWithCancel,
		cancel:      cancel,
		key:         key,
		attempt:     0,
		maxAttempts: e.policy.MaxAttempts,
		startTime:   time.Now(),
		metadata:    make(map[string]any),
		logger:      e.logger,
	}
}

func (rc *retryContext) cleanup() {
	rc.cancel()
}

func (e *RetryExecutor) handleRetry(rc *retryContext, err error) bool {
	rc.attempt++
	rc.lastError = err

	// Check if error is retryable
	if !e.isRetryableError(err) {
		return false
	}

	// Calculate delay
	delay := e.calculateDelay(rc.attempt)

	// Execute retry callback
	if e.policy.OnRetry != nil {
		e.policy.OnRetry(rc.attempt, err)
	}

	// Record retry stats
	e.recordRetry(rc)

	// Wait before next attempt
	select {
	case <-rc.ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

// isRetryableError checks if an error is retryable
func (e *RetryExecutor) isRetryableError(err error) bool {
	// Check against configured retryable errors
	for _, retryableErr := range e.policy.RetryableErrors {
		if errors.Is(err, retryableErr) {
			return true
		}
	}

	// Check if error is explicitly non-retryable
	var nonRetryableErr *types.WorkflowError
	if errors.As(err, &nonRetryableErr) {
		return false
	}

	// Default to retryable
	return true
}

// calculateDelay calculates the retry delay
func (e *RetryExecutor) calculateDelay(attempt int) time.Duration {
	// Calculate exponential backoff
	delay := float64(e.policy.InitialInterval) * math.Pow(e.policy.Multiplier, float64(attempt-1))

	// Apply max interval
	if delay > float64(e.policy.MaxInterval) {
		delay = float64(e.policy.MaxInterval)
	}

	// Add jitter if enabled
	if e.policy.Jitter {
		delay = delay * (0.5 + rand.Float64())
	}

	return time.Duration(delay)
}

// Stats tracking

// initMetrics initializes metrics
func (e *RetryExecutor) initMetrics() error {
	// Retry attempt metrics
	e.metrics.RegisterCounter("retry_attempts")
	e.metrics.RegisterCounter("retry_success")
	e.metrics.RegisterCounter("retry_failure")
	e.metrics.RegisterCounter("retry_max_attempts")

	// Latency metrics
	e.metrics.RegisterHistogram("retry_latency", 1000)
	e.metrics.RegisterGauge("retry_backoff")

	// State metrics
	e.metrics.RegisterGauge("active_retries")
	e.metrics.RegisterGauge("success_rate")

	return nil
}

func (e *RetryExecutor) updateStats(updateFunc func(*RetryStats)) {
	e.stats.mu.Lock()
	defer e.stats.mu.Unlock()
	updateFunc(e.stats)
}

func (e *RetryExecutor) recordRetry(rc *retryContext) {
	e.updateStats(func(stats *RetryStats) {
		stats.RetryAttempts++
		stats.ConsecutiveFailures++
		stats.LastRetryTime = time.Now()

		if e.metrics != nil {
			e.metrics.RecordValue("retry_latency", time.Since(rc.startTime).Seconds())
		}
	})
}

func (e *RetryExecutor) recordSuccess(rc *retryContext) {
	duration := time.Since(rc.startTime)

	e.updateStats(func(stats *RetryStats) {
		stats.RetrySuccesses++
		stats.ConsecutiveFailures = 0
		if duration > stats.MaxRetryLatency {
			stats.MaxRetryLatency = duration
		}
		stats.AvgRetryLatency = (stats.AvgRetryLatency + duration) / 2
		stats.SuccessRate = float64(stats.RetrySuccesses) / float64(stats.RetrySuccesses+stats.RetryFailures)

		if e.metrics != nil {
			e.metrics.AddCounter("retry_success", 1)
		}
	})
}

// BaseExecutorInterface implementation

// Execute executes retry
func (e *RetryExecutor) Execute(ctx context.Context, req *types.Request) (*types.Response, error) {
	return nil, types.NewError(types.ErrNotImplemented, "direct execution not supported", nil)
}

// Cancel cancels retry execution
func (e *RetryExecutor) Cancel(ctx context.Context, id string) error {
	state, ok := e.retryState.Load(id)
	if !ok {
		return types.NewError(types.ErrNotFound, "retry state not found", nil)
	}

	retryState := state.(*RetryState)
	retryState.mu.Lock()
	defer retryState.mu.Unlock()

	e.retryState.Delete(id)
	e.activeRetries.Add(-1)

	return nil
}

// Rollback rolls back a retry execution
func (e *RetryExecutor) Rollback(ctx context.Context, id string) error {
	return types.NewError(types.ErrNotImplemented, "rollback not supported", nil)
}

// GetMetrics returns the retry executor metrics
func (e *RetryExecutor) GetMetrics() map[string]any {
	if !e.config.MetricsEnabled {
		return nil
	}

	return map[string]any{
		"retry_attempts":     e.metrics.GetCounter("retry_attempts"),
		"retry_success":      e.metrics.GetCounter("retry_success"),
		"retry_failure":      e.metrics.GetCounter("retry_failure"),
		"retry_max_attempts": e.metrics.GetCounter("retry_max_attempts"),
		"retry_latency":      e.metrics.GetHistogram("retry_latency"),
		"active_retries":     e.metrics.GetGauge("active_retries"),
		"success_rate":       e.metrics.GetGauge("success_rate"),
		"retry_backoff":      e.metrics.GetGauge("retry_backoff"),
	}
}

// GetCapabilities returns the retry executor capabilities
func (e *RetryExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    false,
		SupportsRetry:    true,
		SupportsRollback: false,
		MaxConcurrency:   e.config.Policy.MaxAttempts,
		MaxBatchSize:     0, // Retry executor doesn't support batch
		AllowedActions: []string{
			"retry",
			"cancel",
		},
	}
}

// Status returns the retry executor status
func (e *RetryExecutor) Status() types.ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// IsHealthy returns true if the retry executor is healthy
func (e *RetryExecutor) IsHealthy() bool {
	return e.Status() == types.ExecutionActive
}

// Internal retry state management

// clearRetryState cleans up retry state
func (e *RetryExecutor) clearRetryState(key string) {
	e.retryState.Delete(key)
	e.activeRetries.Add(-1)
}

// initRetryState initializes retry state for a key
func (e *RetryExecutor) initRetryState(key string) *RetryState {
	state := &RetryState{
		MaxAttempts: e.policy.MaxAttempts,
	}
	e.retryState.Store(key, state)
	e.activeRetries.Add(1)
	return state
}

// updateRetryState updates retry state after an attempt
func (e *RetryExecutor) updateRetryState(state *RetryState, err error, backoff time.Duration) {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.AttemptCount++
	state.LastAttempt = time.Now()
	state.LastError = err
	state.BackoffDuration = backoff
	state.NextAttempt = state.LastAttempt.Add(backoff)
}

// Extended retry functionality

// RetryWithContext executes a function with retry and context
func (e *RetryExecutor) RetryWithContext(ctx context.Context, opts *RetryOptions, fn func(context.Context) error) error {
	if opts == nil {
		opts = DefaultRetryOptions()
	}

	var lastErr error
	for attempt := 0; attempt <= opts.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fn(ctx); err == nil {
				return nil
			} else {
				lastErr = err
				if !e.shouldRetry(err, opts) {
					return err
				}
				if attempt < opts.MaxAttempts {
					time.Sleep(e.getBackoffDuration(attempt, opts))
				}
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// RetryOptions represents retry configuration options
type RetryOptions struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffFactor   float64
	RetryableErrors []error
	OnRetry         func(attempt int, err error)
}

// DefaultRetryOptions returns default retry options
func DefaultRetryOptions() *RetryOptions {
	return &RetryOptions{
		MaxAttempts:    3,
		InitialBackoff: time.Second,
		MaxBackoff:     time.Minute,
		BackoffFactor:  2.0,
	}
}

func (e *RetryExecutor) shouldRetry(err error, opts *RetryOptions) bool {
	// Check if error is in retryable list
	if len(opts.RetryableErrors) > 0 {
		for _, retryableErr := range opts.RetryableErrors {
			if errors.Is(err, retryableErr) {
				return true
			}
		}
		return false
	}

	// Default retry behavior
	return e.isRetryableError(err)
}

func (e *RetryExecutor) getBackoffDuration(attempt int, opts *RetryOptions) time.Duration {
	backoff := float64(opts.InitialBackoff) * math.Pow(opts.BackoffFactor, float64(attempt))
	if backoff > float64(opts.MaxBackoff) {
		backoff = float64(opts.MaxBackoff)
	}
	if e.policy.Jitter {
		backoff = backoff * (0.5 + rand.Float64())
	}
	return time.Duration(backoff)
}

// Advanced retry features

// RetryGroup represents a organization of retryable operations
type RetryGroup struct {
	executor *RetryExecutor
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	errors   []error
	mu       sync.Mutex
}

// NewRetryGroup creates a new retry group
func (e *RetryExecutor) NewRetryGroup(ctx context.Context) *RetryGroup {
	groupCtx, cancel := context.WithCancel(ctx)
	return &RetryGroup{
		executor: e,
		ctx:      groupCtx,
		cancel:   cancel,
	}
}

// Go executes a function in the retry group
func (g *RetryGroup) Go(key string, fn func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := g.executor.ExecuteWithRetry(g.ctx, key, fn); err != nil {
			g.mu.Lock()
			g.errors = append(g.errors, err)
			g.mu.Unlock()
		}
	}()
}

// Wait waits for all retries to complete
func (g *RetryGroup) Wait() []error {
	g.wg.Wait()
	g.cancel()
	return g.errors
}

// RetryWithFallback executes with fallback function
func (e *RetryExecutor) RetryWithFallback(ctx context.Context, key string, fn func(context.Context) error, fallback func(error) error) error {
	err := e.ExecuteWithRetry(ctx, key, fn)
	if err != nil {
		return fallback(err)
	}
	return nil
}

// RetryWithTimeout executes with timeout
func (e *RetryExecutor) RetryWithTimeout(ctx context.Context, key string, timeout time.Duration, fn func(context.Context) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return e.ExecuteWithRetry(timeoutCtx, key, fn)
}

// RetryWithCircuitBreaker executes with circuit breaker
type RetryWithCircuitBreaker struct {
	executor     *RetryExecutor
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailure  time.Time
	mu           sync.RWMutex
}

func (e *RetryExecutor) NewRetryWithCircuitBreaker(maxFailures int, resetTimeout time.Duration) *RetryWithCircuitBreaker {
	return &RetryWithCircuitBreaker{
		executor:     e,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

func (cb *RetryWithCircuitBreaker) Execute(ctx context.Context, key string, fn func(context.Context) error) error {
	cb.mu.RLock()
	if cb.failures >= cb.maxFailures {
		if time.Since(cb.lastFailure) < cb.resetTimeout {
			cb.mu.RUnlock()
			return fmt.Errorf("circuit breaker is open")
		}
		// Reset circuit breaker
		cb.mu.RUnlock()
		cb.mu.Lock()
		cb.failures = 0
		cb.mu.Unlock()
	} else {
		cb.mu.RUnlock()
	}

	err := cb.executor.ExecuteWithRetry(ctx, key, fn)
	if err != nil {
		cb.mu.Lock()
		cb.failures++
		cb.lastFailure = time.Now()
		cb.mu.Unlock()
	}
	return err
}

// RetryableFunc represents a retryable function
type RetryableFunc func(ctx context.Context) error

// WithMaxAttempts sets maximum retry attempts
func (f RetryableFunc) WithMaxAttempts(attempts int) RetryableFunc {
	return func(ctx context.Context) error {
		opts := DefaultRetryOptions()
		opts.MaxAttempts = attempts
		return f(ctx)
	}
}

// WithBackoff sets backoff configuration
func (f RetryableFunc) WithBackoff(initial, max time.Duration, factor float64) RetryableFunc {
	return func(ctx context.Context) error {
		opts := DefaultRetryOptions()
		opts.InitialBackoff = initial
		opts.MaxBackoff = max
		opts.BackoffFactor = factor
		return f(ctx)
	}
}

// WithRetryableErrors sets retryable errors
func (f RetryableFunc) WithRetryableErrors(errors []error) RetryableFunc {
	return func(ctx context.Context) error {
		opts := DefaultRetryOptions()
		opts.RetryableErrors = errors
		return f(ctx)
	}
}

// WithOnRetry sets retry callback
func (f RetryableFunc) WithOnRetry(onRetry func(attempt int, err error)) RetryableFunc {
	return func(ctx context.Context) error {
		opts := DefaultRetryOptions()
		opts.OnRetry = onRetry
		return f(ctx)
	}
}
