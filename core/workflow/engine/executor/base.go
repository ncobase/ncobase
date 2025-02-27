package executor

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
)

// BaseExecutor provides base implementation for executors
type BaseExecutor struct {
	// Basic info
	execType types.ExecutorType
	name     string
	priority int

	// Runtime state
	status     types.ExecutionStatus
	activeJobs atomic.Int64
	mu         sync.RWMutex

	// Components
	metrics *metrics.Collector

	// Configurations
	config          *config.BaseExecutorConfig
	componentConfig *config.ComponentConfig
}

// NewBaseExecutor creates a new base executor
func NewBaseExecutor(execType types.ExecutorType, name string, cfg *config.Config) *BaseExecutor {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	return &BaseExecutor{
		execType:        execType,
		name:            name,
		status:          types.ExecutionPending,
		config:          cfg.Executors.Base,
		componentConfig: cfg.Components,
	}
}

// Basic info methods

func (e *BaseExecutor) Type() types.ExecutorType { return e.execType }
func (e *BaseExecutor) Name() string             { return e.name }
func (e *BaseExecutor) Priority() int            { return e.priority }

// Lifecycle methods

func (e *BaseExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "executor must be pending to start", nil)
	}

	// Initialize metrics if enabled
	if e.config.EnableMetrics {
		if err := e.initMetrics(); err != nil {
			return fmt.Errorf("init metrics failed: %w", err)
		}
	}

	e.status = types.ExecutionActive
	return nil
}

func (e *BaseExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionActive {
		return nil
	}

	e.status = types.ExecutionStopped
	return nil
}

// Pause pauses the executor
func (e *BaseExecutor) Pause() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionActive {
		return types.NewError(types.ErrInvalidParam, "executor must be active to pause", nil)
	}

	e.status = types.ExecutionSuspended
	return nil
}

// Resume resumes the executor
func (e *BaseExecutor) Resume() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionSuspended {
		return types.NewError(types.ErrInvalidParam, "executor must be suspended to resume", nil)
	}

	e.status = types.ExecutionActive
	return nil
}

// Reset resets the executor state
func (e *BaseExecutor) Reset() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == types.ExecutionActive {
		return types.NewError(types.ErrInvalidParam, "cannot reset active executor", nil)
	}

	// Reset metrics
	if e.config.EnableMetrics && e.metrics != nil {
		e.metrics.Reset()
	}

	// Reset state
	e.activeJobs.Store(0)
	e.status = types.ExecutionPending

	return nil
}

// validateConfig validates executor configuration
func (e *BaseExecutor) validateConfig(config *config.BaseExecutorConfig) error {
	if config == nil {
		return types.NewError(types.ErrInvalidParam, "config is nil", nil)
	}

	if config.MaxRetries < 0 {
		return types.NewError(types.ErrInvalidParam, "max retries must be non-negative", nil)
	}

	if config.MaxConcurrent <= 0 {
		return types.NewError(types.ErrInvalidParam, "max concurrent must be positive", nil)
	}

	if config.RetryInterval <= 0 {
		return types.NewError(types.ErrInvalidParam, "retry interval must be positive", nil)
	}

	if config.Workers <= 0 {
		return types.NewError(types.ErrInvalidParam, "workers must be positive", nil)
	}

	return nil
}

// Core operations

// Execute executes a request
func (e *BaseExecutor) Execute(ctx context.Context, req *types.Request) (*types.Response, error) {
	// Validate request
	if err := e.validateRequest(req); err != nil {
		return nil, err
	}

	// Check capacity
	if !e.checkCapacity() {
		return nil, types.NewError(types.ErrExecutionFailed, "executor capacity exceeded", nil)
	}

	// Track execution
	e.activeJobs.Add(1)
	defer e.activeJobs.Add(-1)

	startTime := time.Now()

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Execute with retry
	var resp *types.Response
	var err error

	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		select {
		case <-execCtx.Done():
			return nil, types.NewError(types.ErrTimeout, "execution timeout", execCtx.Err())
		default:
			resp, err = e.doExecute(execCtx, req)
			if err == nil {
				break
			}
			// Retry on retryable errors
			if !e.isRetryableError(err) {
				return nil, err
			}
			if attempt < e.config.MaxRetries {
				time.Sleep(e.config.RetryInterval)
			}
		}
	}

	// Record metrics
	if e.config.EnableMetrics {
		e.recordExecution(time.Since(startTime), err)
	}

	return resp, err
}

// Cancel cancels a request
func (e *BaseExecutor) Cancel(_ context.Context, _ string) error {
	// To be implemented by specific executors
	return types.NewError(types.ErrNotFound, "cancel not implemented", nil)
}

// Rollback rollbacks a request
func (e *BaseExecutor) Rollback(_ context.Context, _ string) error {
	// To be implemented by specific executors
	return types.NewError(types.ErrNotFound, "rollback not implemented", nil)
}

// Configure configures the executor

// SetMaxConcurrent sets the maximum number of concurrent requests
func (e *BaseExecutor) SetMaxConcurrent(max int32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if max > 0 {
		e.config.MaxConcurrent = max
	}
}

// SetTimeout sets the timeout
func (e *BaseExecutor) SetTimeout(timeout time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if timeout > 0 {
		e.config.Timeout = timeout
	}
}

// ResourceUsage returns resource usage
func (e *BaseExecutor) ResourceUsage() *types.ResourceUsage {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &types.ResourceUsage{
		MemoryUsage:    int64(memStats.Alloc),
		CPUUsage:       getCPUUsage(),
		GoroutineCount: runtime.NumGoroutine(),
		ActiveJobs:     e.activeJobs.Load(),
	}
}

// getCPUUsage returns the CPU usage percentage
func getCPUUsage() float64 {
	var cpuStats runtime.MemStats
	runtime.ReadMemStats(&cpuStats)

	// Calculate CPU usage based on GC stats and NumGC
	gcTime := float64(cpuStats.PauseTotalNs) / 1e9 // Convert to seconds
	timeSinceLastGC := time.Since(time.Unix(0, int64(cpuStats.LastGC))).Seconds()

	if timeSinceLastGC == 0 {
		return 0
	}

	// Rough CPU usage estimation
	return (gcTime / timeSinceLastGC) * 100
}

// Status & capabilities

func (e *BaseExecutor) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status == types.ExecutionActive
}

func (e *BaseExecutor) GetMetrics() map[string]any {
	if !e.config.EnableMetrics {
		return nil
	}

	return map[string]any{
		"active_jobs":   e.activeJobs.Load(),
		"success_count": e.metrics.GetCounter("exec_success"),
		"error_count":   e.metrics.GetCounter("exec_error"),
		"exec_time":     e.metrics.GetHistogram("exec_time"),
	}
}

func (e *BaseExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    e.config.EnableAsync,
		SupportsRetry:    e.config.MaxRetries > 0,
		SupportsRollback: false,
		MaxConcurrency:   int(e.config.MaxConcurrent),
		MaxBatchSize:     int(e.config.QueueSize),
	}
}

// validateRequest validates request
func (e *BaseExecutor) validateRequest(req *types.Request) error {
	if req == nil {
		return types.NewError(types.ErrInvalidParam, "request is nil", nil)
	}

	if req.Type != e.execType {
		return types.NewError(types.ErrInvalidParam, "invalid request type", nil)
	}

	return nil
}

func (e *BaseExecutor) checkCapacity() bool {
	return e.activeJobs.Load() < int64(e.config.MaxConcurrent)
}

func (e *BaseExecutor) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check error type
	var wErr *types.WorkflowError
	if !errors.As(err, &wErr) {
		return false
	}

	// Define retryable error codes
	retryableCodes := map[types.ErrorCode]bool{
		types.ErrTimeout: true,
		types.ErrSystem:  true,
	}

	return retryableCodes[wErr.Code]
}

func (e *BaseExecutor) doExecute(ctx context.Context, req *types.Request) (*types.Response, error) {
	// To be implemented by specific executors
	return nil, types.NewError(types.ErrNotFound, "execute not implemented", nil)
}

func (e *BaseExecutor) initMetrics() (err error) {
	e.metrics, err = metrics.NewCollector(e.componentConfig.Metrics)
	if err != nil {
		return fmt.Errorf("failed to create metrics collector: %w", err)
	}

	// Register basic metrics
	e.metrics.RegisterCounter("exec_total")
	e.metrics.RegisterCounter("exec_success")
	e.metrics.RegisterCounter("exec_error")
	e.metrics.RegisterHistogram("exec_time", 1000)

	return nil
}

func (e *BaseExecutor) recordExecution(duration time.Duration, err error) {
	e.metrics.AddCounter("exec_total", 1)
	if err != nil {
		e.metrics.AddCounter("exec_error", 1)
	} else {
		e.metrics.AddCounter("exec_success", 1)
	}
	e.metrics.RecordValue("exec_time", duration.Seconds())
}
