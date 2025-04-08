package executor

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"ncore/pkg/logger"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sony/gobreaker"
)

// ServiceExecutor handles service execution
type ServiceExecutor struct {
	*BaseExecutor

	// Dependencies
	services *service.Service
	em       *extension.Manager
	logger   logger.Logger

	// Components
	metrics *metrics.Collector
	breaker *gobreaker.CircuitBreaker
	retries *RetryExecutor

	// Service tracking
	servicesMap sync.Map // serviceID -> *ServiceInfo
	activeCount atomic.Int64

	// Registry
	registry *ServiceRegistry

	// Cache
	cache *sync.Map

	// Config
	config *config.ServiceExecutorConfig

	// Runtime state
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// ServiceInfo represents service execution info
type ServiceInfo struct {
	ID         string
	Name       string
	Type       string
	Status     string
	Input      map[string]any
	CacheKey   string
	CacheTTL   time.Duration
	StartTime  time.Time
	EndTime    *time.Time
	Timeout    time.Duration
	RetryCount int
	Error      error
	mu         sync.RWMutex
}

// ServiceProvider interface
type ServiceProvider interface {
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)
	Validate(input map[string]any) error
}

// NewServiceExecutor creates a new service executor
func NewServiceExecutor(svc *service.Service, em *extension.Manager, cfg *config.Config) (*ServiceExecutor, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &ServiceExecutor{
		BaseExecutor: NewBaseExecutor(types.ServiceExecutor, "Service Executor", cfg),
		services:     svc,
		em:           em,
		config:       cfg.Executors.Service,
		registry:     NewServiceRegistry(),
		cache:        &sync.Map{},
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize circuit breaker
	executor.breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "service-executor",
		MaxRequests: uint32(executor.config.CBMaxRequests),
		Interval:    executor.config.CBInterval,
		Timeout:     executor.config.CBTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	// Initialize metrics if enabled
	if executor.config.EnableMetrics {
		collector, err := metrics.NewCollector(cfg.Components.Metrics)
		if err != nil {
			return nil, fmt.Errorf("create metrics collector failed: %w", err)
		}
		executor.metrics = collector
	}

	// Initialize retry executor
	retryExecutor, err := NewRetryExecutor(svc, em, cfg)
	if err != nil {
		return nil, fmt.Errorf("create retry executor failed: %w", err)
	}
	executor.retries = retryExecutor

	return executor, nil
}

// Start starts the service executor
func (e *ServiceExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status() != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "executor not in pending state", nil)
	}

	// Start retry executor
	if err := e.retries.Start(); err != nil {
		return fmt.Errorf("start retry executor failed: %w", err)
	}

	// Initialize metrics if enabled
	if e.config.EnableMetrics {
		if err := e.initMetrics(); err != nil {
			return fmt.Errorf("initialize metrics failed: %w", err)
		}
	}

	return nil
}

// Stop stops the service executor
func (e *ServiceExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.Status() != types.ExecutionActive {
		return nil
	}

	// Cancel context
	e.cancel()

	// Stop retry executor
	if err := e.retries.Stop(); err != nil {
		return fmt.Errorf("stop retry executor failed: %w", err)
	}

	return nil
}

// Execute service execution
func (e *ServiceExecutor) Execute(ctx context.Context, req *types.Request) (*types.Response, error) {
	// Validate request
	if err := e.validateRequest(req); err != nil {
		return nil, err
	}

	// Get service name from request
	serviceName, ok := req.Context["service"].(string)
	if !ok {
		return nil, types.NewError(types.ErrInvalidParam, "missing service name", nil)
	}

	// Get service provider
	provider, exists := e.registry.GetProvider(serviceName)
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "service provider not found", nil)
	}

	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Create service info
	info := &ServiceInfo{
		ID:        req.ID,
		Name:      serviceName,
		Status:    string(types.ExecutionActive),
		Input:     req.Variables,
		StartTime: time.Now(),
		Timeout:   e.config.Timeout,
	}

	// Track service execution
	e.servicesMap.Store(req.ID, info)
	e.activeCount.Add(1)
	defer func() {
		e.servicesMap.Delete(req.ID)
		e.activeCount.Add(-1)
	}()

	// Check cache if enabled
	if e.config.EnableCache {
		if response := e.checkCache(req.ID); response != nil {
			return response, nil
		}
	}

	// Check circuit breaker state
	if !e.Ready() {
		return nil, types.NewError(types.ErrSystem, "circuit breaker is open", nil)
	}

	// Execute with circuit breaker
	result, err := e.breaker.Execute(func() (any, error) {
		// Execute with retry
		var output map[string]any
		err := e.retries.ExecuteWithRetry(execCtx, req.ID, func(retryCtx context.Context) error {
			select {
			case <-retryCtx.Done():
				return types.NewError(types.ErrTimeout, "service execution timeout", nil)
			default:
				var execErr error
				output, execErr = provider.Execute(retryCtx, req.Variables)
				return execErr
			}
		})
		if err != nil {
			return nil, err
		}
		return output, nil
	})

	if err != nil {
		// Record metrics
		if e.metrics != nil {
			e.metrics.AddCounter("service_error", 1)
			if errors.Is(err, context.DeadlineExceeded) {
				e.metrics.AddCounter("service_timeout", 1)
			}
		}
		return nil, fmt.Errorf("service execution failed: %w", err)
	}

	// Update service info
	info.mu.Lock()
	info.Status = string(types.ExecutionCompleted)
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Cache result if enabled
	if e.config.EnableCache {
		e.cacheResponse(req.ID, result.(map[string]any), e.config.CacheTTL)
	}

	return &types.Response{
		ID:        req.ID,
		Status:    types.ExecutionCompleted,
		Data:      result,
		StartTime: info.StartTime,
		EndTime:   info.EndTime,
		Duration:  time.Since(info.StartTime),
	}, nil
}

// ExecuteServiceAsync executes service asynchronously
func (e *ServiceExecutor) ExecuteServiceAsync(_ context.Context, req *types.Request) error {
	if !e.config.EnableAsync {
		return types.NewError(types.ErrInvalidParam, "async execution not enabled", nil)
	}

	// Create error channel to handle immediate errors
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		// Create new context for async execution
		asyncCtx, cancel := context.WithTimeout(context.Background(), e.config.Timeout)
		defer cancel()

		_, err := e.Execute(asyncCtx, req)
		if err != nil {
			errCh <- err
			e.logger.Error(asyncCtx, "async service execution failed", err)

			// Publish failure event
			e.em.PublishEvent("service.failed", map[string]any{
				"service_id": req.ID,
				"error":      err.Error(),
				"time":       time.Now(),
			})
		}
	}()

	// Wait for immediate errors
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("async service execution failed: %w", err)
		}
	case <-time.After(time.Second): // Wait briefly for immediate failures
	}

	return nil
}

// CancelService cancels a service execution
func (e *ServiceExecutor) CancelService(ctx context.Context, serviceID string) error {
	info, err := e.GetServiceInfo(ctx, serviceID)
	if err != nil {
		return err
	}

	info.mu.Lock()
	if info.Status != string(types.ExecutionActive) {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "service not active", nil)
	}

	// Update status
	info.Status = string(types.ExecutionCancelled)
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Clean up
	e.servicesMap.Delete(serviceID)
	e.activeCount.Add(-1)

	// Publish event
	e.em.PublishEvent("service.cancelled", map[string]any{
		"service_id": serviceID,
		"time":       now,
	})

	return nil
}

// RollbackService rolls back a service execution
func (e *ServiceExecutor) RollbackService(ctx context.Context, serviceID string) error {
	info, err := e.GetServiceInfo(ctx, serviceID)
	if err != nil {
		return err
	}

	info.mu.Lock()
	if info.Status != string(types.ExecutionCompleted) {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "service not completed", nil)
	}

	// Get provider
	provider, err := e.GetProvider(info.Name)
	if err != nil {
		return err
	}

	// Execute rollback if supported
	if rollbacker, ok := provider.(interface {
		Rollback(ctx context.Context, input map[string]any) error
	}); ok {
		if err := rollbacker.Rollback(ctx, info.Input); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
	}

	// Update status
	info.Status = string(types.ExecutionRollbacked)
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Clean up cache
	if info.CacheKey != "" {
		e.cache.Delete(info.CacheKey)
	}

	return nil
}

// Cancel cancels service execution
func (e *ServiceExecutor) Cancel(ctx context.Context, id string) error {
	return e.CancelService(ctx, id)
}

// Rollback rolls back service execution
func (e *ServiceExecutor) Rollback(ctx context.Context, id string) error {
	return e.RollbackService(ctx, id)
}

// RegisterProvider registers a service provider
func (e *ServiceExecutor) RegisterProvider(name string, provider ServiceProvider) error {
	if provider == nil {
		return types.NewError(types.ErrInvalidParam, "provider is nil", nil)
	}

	e.registry.RegisterProvider(name, provider)
	return nil
}

// GetProvider gets a service provider
func (e *ServiceExecutor) GetProvider(name string) (ServiceProvider, error) {
	provider, exists := e.registry.GetProvider(name)
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "provider not found", nil)
	}
	return provider, nil
}

// GetServiceInfo gets service execution info
func (e *ServiceExecutor) GetServiceInfo(_ context.Context, serviceID string) (*ServiceInfo, error) {
	info, ok := e.servicesMap.Load(serviceID)
	if !ok {
		return nil, types.NewError(types.ErrNotFound, "service not found", nil)
	}
	return info.(*ServiceInfo), nil
}

// GetActiveServices gets all active services
func (e *ServiceExecutor) GetActiveServices(_ context.Context) ([]*ServiceInfo, error) {
	var active []*ServiceInfo
	e.servicesMap.Range(func(_, value any) bool {
		info := value.(*ServiceInfo)
		info.mu.RLock()
		if info.Status == string(types.ExecutionActive) {
			active = append(active, info)
		}
		info.mu.RUnlock()
		return true
	})
	return active, nil
}

// Status returns executor status
func (e *ServiceExecutor) Status() types.ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetCapabilities returns executor capabilities
func (e *ServiceExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    e.config.EnableAsync,
		SupportsRetry:    true,
		SupportsRollback: true,
		MaxConcurrency:   e.config.CBMaxRequests,
		MaxBatchSize:     0, // Service executor doesn't support batch
		AllowedActions: []string{
			"execute",
			"cancel",
			"rollback",
		},
	}
}

// IsHealthy returns executor health status
func (e *ServiceExecutor) IsHealthy() bool {
	return e.Status() == types.ExecutionActive && e.Ready()
}

// Ready checks if circuit breaker is ready
func (e *ServiceExecutor) Ready() bool {
	counts := e.breaker.Counts()
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	return counts.Requests >= 3 && failureRatio < 0.6
}

// Service registry

type ServiceRegistry struct {
	providers map[string]ServiceProvider
	mu        sync.RWMutex
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		providers: make(map[string]ServiceProvider),
	}
}

func (r *ServiceRegistry) RegisterProvider(name string, provider ServiceProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

func (r *ServiceRegistry) GetProvider(name string) (ServiceProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[name]
	return provider, exists
}

// Cache related types and methods

type CacheEntry struct {
	Data      map[string]any
	ExpiresAt time.Time
}

func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

func (e *ServiceExecutor) checkCache(key string) *types.Response {
	if value, ok := e.cache.Load(key); ok {
		entry := value.(*CacheEntry)
		if !entry.IsExpired() {
			if e.metrics != nil {
				e.metrics.AddCounter("cache_hit", 1)
			}
			return &types.Response{
				ID:     key,
				Status: types.ExecutionCompleted,
				Data:   entry.Data,
			}
		}
		e.cache.Delete(key)
	}

	if e.metrics != nil {
		e.metrics.AddCounter("cache_miss", 1)
	}
	return nil
}

// cacheResponse caches a response
func (e *ServiceExecutor) cacheResponse(key string, data map[string]any, ttl time.Duration) {
	entry := &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
	e.cache.Store(key, entry)
}

// handleTimeout handles timeout strategy for a service
func (e *ServiceExecutor) handleTimeout(task *structs.ReadTask) error {
	node, err := e.services.Node.Get(e.ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
	if err != nil {
		return err
	}

	strategy := e.getTimeoutStrategy(node)
	switch strategy {
	case structs.TimeoutAutoPass:
		return e.handleTimeoutAutoPass(task)
	case structs.TimeoutAutoFail:
		return e.handleTimeoutAutoFail(task)
	case structs.TimeoutAlert:
		return e.handleTimeoutAlert(task)
	default:
		return fmt.Errorf("unknown timeout strategy: %s", strategy)
	}
}

// getTimeoutStrategy gets timeout strategy for a node
func (e *ServiceExecutor) getTimeoutStrategy(node *structs.ReadNode) structs.TimeoutStrategy {
	if strategy, ok := node.TimeoutConfig["strategy"].(string); ok {
		return structs.TimeoutStrategy(strategy)
	}

	// Fallback to template
	template, err := e.services.Template.Get(e.ctx, &structs.FindTemplateParams{
		ID: node.TemplateID,
	})
	if err == nil && template.TimeoutConfig != nil {
		if strategy, ok := template.TimeoutConfig["strategy"].(string); ok {
			return structs.TimeoutStrategy(strategy)
		}
	}

	return structs.TimeoutNone
}

// handleTimeoutAutoPass automatically passes a timed-out task
func (e *ServiceExecutor) handleTimeoutAutoPass(task *structs.ReadTask) error {
	req := &structs.CompleteTaskRequest{
		TaskID:   task.ID,
		Action:   structs.ActionApprove,
		Operator: "system",
		Comment:  "Auto approved due to timeout",
	}

	_, err := e.services.Task.Complete(e.ctx, req)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	e.em.PublishEvent(structs.EventTaskCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		Action:    structs.ActionApprove,
	})

	return nil
}

// handleTimeoutAutoFail automatically fails a timed-out task
func (e *ServiceExecutor) handleTimeoutAutoFail(task *structs.ReadTask) error {
	req := &structs.CompleteTaskRequest{
		TaskID:   task.ID,
		Action:   structs.ActionReject,
		Operator: "system",
		Comment:  "Auto rejected due to timeout",
	}

	_, err := e.services.Task.Complete(e.ctx, req)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	e.em.PublishEvent(structs.EventTaskCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		Action:    structs.ActionReject,
	})

	return nil
}

// handleTimeoutAlert updates the task status on timeout
func (e *ServiceExecutor) handleTimeoutAlert(task *structs.ReadTask) error {
	_, err := e.services.Task.Update(e.ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			IsTimeout: true,
		},
	})
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	e.em.PublishEvent(structs.EventTaskTimeout, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
	})

	return nil
}

// completeTimerNode completes a timer node
func (e *ServiceExecutor) completeTimerNode(task *structs.ReadTask) error {
	node, err := e.services.Node.Get(e.ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
	if err != nil {
		return fmt.Errorf("get node: %w", err)
	}

	// Update node status
	_, err = e.services.Node.Update(e.ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
		},
	})
	if err != nil {
		return fmt.Errorf("update node: %w", err)
	}

	e.em.PublishEvent(structs.EventNodeCompleted, &structs.EventData{
		ProcessID: task.ProcessID,
		NodeID:    task.NodeKey,
	})

	return nil
}

// getTimerTriggerTime calculates the next trigger time for a timer node
func (e *ServiceExecutor) getTimerTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	timerType, _ := config["type"].(string)
	switch timerType {
	case "delay":
		return e.getDelayTriggerTime(node, config)
	case "cron":
		return e.getCronTriggerTime(node, config)
	case "date":
		return e.getDateTriggerTime(node, config)
	default:
		return time.Time{}
	}
}

// getDelayTriggerTime calculates the next trigger time for a delay timer
func (e *ServiceExecutor) getDelayTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	duration, ok := config["duration"].(string)
	if !ok {
		return time.Time{}
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return time.Time{}
	}

	startTime := time.Unix(*node.CreatedAt, 0)
	return startTime.Add(d)
}

// getCronTriggerTime calculates the next trigger time for a cron timer
func (e *ServiceExecutor) getCronTriggerTime(node *structs.ReadNode, config map[string]any) time.Time {
	expr, ok := config["cron"].(string)
	if !ok {
		return time.Time{}
	}

	schedule, err := cron.ParseStandard(expr)
	if err != nil {
		return time.Time{}
	}

	var lastTime time.Time
	if node.UpdatedAt != nil {
		lastTime = time.Unix(*node.UpdatedAt, 0)
	} else {
		lastTime = time.Unix(*node.CreatedAt, 0)
	}

	return schedule.Next(lastTime)
}

// getDateTriggerTime calculates the next trigger time for a date timer
func (e *ServiceExecutor) getDateTriggerTime(_ *structs.ReadNode, config map[string]any) time.Time {
	dateStr, ok := config["date"].(string)
	if !ok {
		return time.Time{}
	}

	triggerTime, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}
	}

	return triggerTime
}

func (e *ServiceExecutor) validateRequest(req *types.Request) error {
	if req == nil {
		return types.NewError(types.ErrInvalidParam, "request is nil", nil)
	}

	if req.Type != types.ServiceExecutor {
		return types.NewError(types.ErrInvalidParam, "invalid request type", nil)
	}

	if req.Context == nil {
		return types.NewError(types.ErrInvalidParam, "missing request context", nil)
	}

	if _, ok := req.Context["service"].(string); !ok {
		return types.NewError(types.ErrInvalidParam, "missing service name", nil)
	}

	return nil
}

// initMetrics initializes metrics
func (e *ServiceExecutor) initMetrics() error {
	// Service execution metrics
	e.metrics.RegisterCounter("service_total")
	e.metrics.RegisterCounter("service_success")
	e.metrics.RegisterCounter("service_error")
	e.metrics.RegisterCounter("service_timeout")
	e.metrics.RegisterCounter("service_cancelled")

	// Circuit breaker metrics
	e.metrics.RegisterCounter("circuit_open")
	e.metrics.RegisterCounter("circuit_close")
	e.metrics.RegisterGauge("circuit_failure_rate")

	// Cache metrics
	e.metrics.RegisterCounter("cache_hit")
	e.metrics.RegisterCounter("cache_miss")
	e.metrics.RegisterGauge("cache_size")

	// Resource metrics
	e.metrics.RegisterGauge("active_services")
	e.metrics.RegisterGauge("queued_services")

	// Performance metrics
	e.metrics.RegisterHistogram("service_execution_time", 1000)
	e.metrics.RegisterHistogram("service_queue_time", 1000)

	return nil
}

// GetMetrics returns executor metrics
func (e *ServiceExecutor) GetMetrics() map[string]any {
	if !e.config.EnableMetrics {
		return nil
	}

	m := make(map[string]any)

	// Service execution metrics
	m["service_total"] = e.metrics.GetCounter("service_total")
	m["service_success"] = e.metrics.GetCounter("service_success")
	m["service_error"] = e.metrics.GetCounter("service_error")
	m["service_timeout"] = e.metrics.GetCounter("service_timeout")
	m["service_cancelled"] = e.metrics.GetCounter("service_cancelled")

	// Circuit breaker metrics
	m["circuit_open"] = e.metrics.GetCounter("circuit_open")
	m["circuit_close"] = e.metrics.GetCounter("circuit_close")
	m["circuit_failure_rate"] = e.metrics.GetGauge("circuit_failure_rate")

	// Cache metrics
	m["cache_hit"] = e.metrics.GetCounter("cache_hit")
	m["cache_miss"] = e.metrics.GetCounter("cache_miss")
	m["cache_size"] = e.metrics.GetGauge("cache_size")

	// Resource metrics
	m["active_services"] = e.metrics.GetGauge("active_services")
	m["queued_services"] = e.metrics.GetGauge("queued_services")

	// Performance metrics
	m["service_execution_time"] = e.metrics.GetHistogram("service_execution_time")
	m["service_queue_time"] = e.metrics.GetHistogram("service_queue_time")

	return m
}

// ResourceUsage returns current resource usage
func (e *ServiceExecutor) ResourceUsage() *types.ResourceUsage {
	return &types.ResourceUsage{
		ActiveJobs: e.activeCount.Load(),
	}
}

// Status management methods

func (e *ServiceExecutor) setStatus(status types.ExecutionStatus) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.status = status
}

func (e *ServiceExecutor) recordServiceMetrics(info *ServiceInfo, err error) {
	if !e.config.EnableMetrics {
		return
	}

	e.metrics.AddCounter("service_total", 1)

	if err != nil {
		e.metrics.AddCounter("service_error", 1)
		if errors.Is(err, context.DeadlineExceeded) {
			e.metrics.AddCounter("service_timeout", 1)
		}
	} else {
		e.metrics.AddCounter("service_success", 1)
	}

	if info.EndTime != nil {
		duration := info.EndTime.Sub(info.StartTime)
		e.metrics.RecordValue("service_execution_time", duration.Seconds())
	}

	e.metrics.SetGauge("active_services", float64(e.activeCount.Load()))
}

// Event publishing helper
func (e *ServiceExecutor) publishEvent(eventType string, serviceID string, details map[string]any) {
	if details == nil {
		details = make(map[string]any)
	}
	details["service_id"] = serviceID
	details["time"] = time.Now()

	e.em.PublishEvent(eventType, details)
}
