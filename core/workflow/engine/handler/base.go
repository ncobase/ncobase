package handler

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	ext "github.com/ncobase/ncore/ext/types"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/queue"
	"github.com/ncobase/ncore/pkg/worker"
)

// TaskInfo tracks task execution state
type TaskInfo struct {
	ID         string
	StartTime  time.Time
	Status     string
	Error      error
	RetryCount int
	Variables  map[string]any
}

// HistoryRecord tracks execution history
type HistoryRecord struct {
	Time      time.Time
	EventType string
	Details   map[string]any
}

// BaseHandler provides common handler functionality
type BaseHandler struct {
	// Basic info
	handlerType types.HandlerType
	name        string
	priority    int

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	logger   logger.Logger

	// Components
	metrics *metrics.Collector

	// Task processing
	taskQueue *queue.TaskQueue
	workPool  *worker.Pool

	// Configuration
	config *config.BaseHandlerConfig

	// Runtime state
	status    types.HandlerStatus
	lastError error
	mu        sync.RWMutex

	// Task tracking
	tasks   sync.Map // taskID -> *TaskInfo
	history []*HistoryRecord

	// Execution stats
	execCount    atomic.Int64
	successCount atomic.Int64
	failureCount atomic.Int64

	// Context for lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(handlerType types.HandlerType, name string, svc *service.Service, em ext.ManagerInterface, cfg *config.BaseHandlerConfig) *BaseHandler {
	if cfg == nil {
		cfg = config.DefaultBaseHandlerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	h := &BaseHandler{
		handlerType: handlerType,
		name:        name,
		services:    svc,
		em:          em,
		config:      cfg,
		status:      types.HandlerReady,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize task queue
	queueCfg := queue.DefaultConfig()
	queueCfg.Workers = h.config.MaxWorkers
	queueCfg.QueueSize = int(h.config.QueueSize)
	queueCfg.TaskTimeout = h.config.Timeout

	h.taskQueue = queue.NewTaskQueue(context.Background(), queue.WithConfig(queueCfg))

	// Initialize worker pool
	workerCfg := &worker.Config{
		MaxWorkers:  h.config.MaxWorkers,
		QueueSize:   int(h.config.QueueSize),
		TaskTimeout: h.config.Timeout,
	}
	h.workPool = worker.NewPool(workerCfg, h)

	return h
}

// Type returns handler type
func (h *BaseHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *BaseHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *BaseHandler) Priority() int { return h.priority }

// Start handler
func (h *BaseHandler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status != types.HandlerReady {
		return types.NewError(types.ErrInvalidStatus, "handler must be ready to start", nil)
	}

	// Start components
	h.taskQueue.Start()
	h.workPool.Start()

	// Initialize metrics if enabled
	if h.config.EnableMetrics {
		// h.metrics = metrics.NewCollector(nil)
		h.initMetrics()
	}

	h.status = types.HandlerRunning
	return nil
}

// Stop handler
func (h *BaseHandler) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status != types.HandlerRunning {
		return nil
	}

	// Stop components with timeout
	h.cancel()

	_ = h.taskQueue.Stop(h.ctx)
	h.workPool.Stop(h.ctx)

	h.status = types.HandlerStopped
	return nil
}

// Reset resets handler state
func (h *BaseHandler) Reset() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == types.HandlerRunning {
		return types.NewError(types.ErrInvalidStatus, "cannot reset while running", nil)
	}

	// Reset state
	h.status = types.HandlerReady
	h.lastError = nil
	h.tasks = sync.Map{}
	h.history = nil

	// Reset stats
	h.execCount.Store(0)
	h.successCount.Store(0)
	h.failureCount.Store(0)

	// Reset components
	if h.metrics != nil {
		h.metrics.Reset()
	}

	if h.taskQueue != nil {
		ctx := context.Background()
		_ = h.taskQueue.Stop(ctx)
		h.taskQueue = queue.NewTaskQueue(ctx)
	}

	if h.workPool != nil {
		ctx := context.Background()
		h.workPool.Stop(ctx)
		workerCfg := &worker.Config{
			MaxWorkers:  h.config.MaxWorkers,
			QueueSize:   int(h.config.QueueSize),
			TaskTimeout: h.config.Timeout,
		}
		h.workPool = worker.NewPool(workerCfg, h)
	}

	return nil
}

// Status returns handler status
func (h *BaseHandler) Status() types.HandlerStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// IsHealthy returns health status
func (h *BaseHandler) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status == types.HandlerRunning
}

// GetCapabilities returns handler capabilities
func (h *BaseHandler) GetCapabilities() *types.HandlerCapabilities {
	return &types.HandlerCapabilities{
		SupportsRollback: h.config.EnableRollback,
		SupportsRetry:    h.config.MaxRetries > 0,
		SupportsAsync:    h.config.EnableAsync,
		SupportsBatch:    false,
		MaxConcurrency:   int(h.config.MaxConcurrent),
		MaxBatchSize:     int(h.config.MaxBatchSize),
	}
}

// Execute executes a node
func (h *BaseHandler) Execute(ctx context.Context, node *structs.ReadNode) error {
	if err := h.validateExecution(node); err != nil {
		return err
	}

	startTime := time.Now()
	h.execCount.Add(1)

	// Submit task
	if err := h.submitTask(ctx, node); err != nil {
		h.failureCount.Add(1)
		h.recordError(err)
		return err
	}

	h.successCount.Add(1)
	h.recordExecution(time.Since(startTime), nil)
	return nil
}

// Complete completes a node
func (h *BaseHandler) Complete(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	if err := h.validateStatus(); err != nil {
		return err
	}

	// Record completion
	h.recordHistory("node.complete", map[string]any{
		"node_id":  node.ID,
		"action":   req.Action,
		"operator": req.Operator,
		"comment":  req.Comment,
	})

	return h.completeInternal(ctx, node, req)
}

// Validate validates a node
func (h *BaseHandler) Validate(node *structs.ReadNode) error {
	if node == nil {
		return types.NewError(types.ErrInvalidParam, "node is nil", nil)
	}

	if node.Type != h.handlerType.String() {
		return types.NewError(types.ErrInvalidParam, "invalid node type", nil)
	}

	// Validate config if present
	if c, ok := node.Properties["config"].(map[string]any); ok {
		if err := h.validateConfig(c); err != nil {
			return err
		}
	}

	return h.validateInternal(node)
}

// validateConfig validates handler configuration
func (h *BaseHandler) validateConfig(config map[string]any) error {
	// Validate workers
	if workers, ok := config["workers"].(int); ok {
		if workers <= 0 {
			return types.NewError(types.ErrValidation, "workers must be positive", nil)
		}
	}

	// Validate queue size
	if queueSize, ok := config["queue_size"].(int); ok {
		if queueSize <= 0 {
			return types.NewError(types.ErrValidation, "queue size must be positive", nil)
		}
	}

	// Validate timeout
	if timeout, ok := config["timeout"].(time.Duration); ok {
		if timeout < 0 {
			return types.NewError(types.ErrValidation, "timeout cannot be negative", nil)
		}
	}

	return nil
}

// initMetrics initializes metrics collectors
func (h *BaseHandler) initMetrics() {
	h.metrics.RegisterCounter("exec_total")
	h.metrics.RegisterCounter("exec_success")
	h.metrics.RegisterCounter("exec_failure")
	h.metrics.RegisterCounter("exec_timeout")
	h.metrics.RegisterCounter("exec_cancelled")

	h.metrics.RegisterGauge("active_tasks")
	h.metrics.RegisterGauge("queue_size")

	h.metrics.RegisterHistogram("exec_duration", 1000)
	h.metrics.RegisterHistogram("task_duration", 1000)
}

// GetMetrics returns handler metrics
func (h *BaseHandler) GetMetrics() map[string]any {
	return map[string]any{
		"executions":     h.execCount.Load(),
		"successful":     h.successCount.Load(),
		"failed":         h.failureCount.Load(),
		"active_tasks":   h.countActiveTasks(),
		"queue_metrics":  h.taskQueue.GetMetrics(),
		"worker_metrics": h.workPool.GetMetrics(),
	}
}

// Process queued tasks
func (h *BaseHandler) Process(task any) error {
	if qTask, ok := task.(*queue.QueuedTask); ok {
		if node, ok := qTask.Data.(*structs.ReadNode); ok {
			return h.executeInternal(qTask.Context, node)
		}
	}
	return types.NewError(types.ErrInvalidParam, "invalid task type", nil)
}

// Internal methods that should be implemented by concrete handlers
func (h *BaseHandler) executeInternal(_ context.Context, _ *structs.ReadNode) error {
	return types.NewError(types.ErrNotImplemented, "executeInternal not implemented", nil)
}

// completeInternal completes a node
func (h *BaseHandler) completeInternal(_ context.Context, _ *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	return types.NewError(types.ErrNotImplemented, "completeInternal not implemented", nil)
}

// validateInternal performs internal validation
func (h *BaseHandler) validateInternal(_ *structs.ReadNode) error {
	return nil // Default validation passes
}

// submitTask submits a task
func (h *BaseHandler) submitTask(ctx context.Context, node *structs.ReadNode) error {
	task := &queue.QueuedTask{
		ID:       node.ID,
		Type:     h.handlerType.String(),
		Context:  ctx,
		Data:     node,
		Priority: h.priority,
	}

	return h.taskQueue.Push(task)
}

// validateExecution validates execution
func (h *BaseHandler) validateExecution(node *structs.ReadNode) error {
	if err := h.validateStatus(); err != nil {
		return err
	}
	return h.Validate(node)
}

// validateStatus checks handler status
func (h *BaseHandler) validateStatus() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch h.status {
	case types.HandlerRunning:
		return nil
	case types.HandlerStopped:
		return types.NewError(types.ErrInvalidStatus, "handler is stopped", nil)
	case types.HandlerError:
		return types.NewError(types.ErrInvalidStatus, "handler is in error state", nil)
	case types.HandlerPaused:
		return types.NewError(types.ErrInvalidStatus, "handler is paused", nil)
	default:
		return types.NewError(types.ErrInvalidStatus, fmt.Sprintf("invalid handler status: %s", h.status), nil)
	}
}

// recordExecution records execution
func (h *BaseHandler) recordExecution(duration time.Duration, err error) {
	h.execCount.Add(1)
	if err != nil {
		h.failureCount.Add(1)
	} else {
		h.successCount.Add(1)
	}

	if h.metrics != nil {
		h.metrics.RecordDuration("execution_time", duration)
	}
}

// countActiveTasks counts active tasks
func (h *BaseHandler) countActiveTasks() int {
	count := 0
	h.tasks.Range(func(_, value any) bool {
		taskInfo := value.(*TaskInfo)
		if taskInfo.Status == string(types.ExecutionActive) {
			count++
		}
		return true
	})
	return count
}

// GetHistory returns execution history
func (h *BaseHandler) GetHistory() []*HistoryRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	history := make([]*HistoryRecord, len(h.history))
	copy(history, h.history)
	return history
}

// recordHistory adds an event to history
func (h *BaseHandler) recordHistory(eventType string, details map[string]any) {
	record := &HistoryRecord{
		Time:      time.Now(),
		EventType: eventType,
		Details:   details,
	}
	h.history = append(h.history, record)

	// Also publish event
	h.em.PublishEvent(eventType, details)
}

// Configure configures handler
func (h *BaseHandler) Configure(cfg *config.BaseHandlerConfig) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.status == types.HandlerRunning {
		return types.NewError(types.ErrInvalidStatus, "cannot configure while running", nil)
	}

	h.config = cfg
	return nil
}

func (h *BaseHandler) GetConfig() *config.BaseHandlerConfig {
	return h.config
}

// Rollback rollbacks node execution
func (h *BaseHandler) Rollback(ctx context.Context, node *structs.ReadNode) error {
	if err := h.validateStatus(); err != nil {
		return err
	}
	return h.rollbackInternal(ctx, node)
}

func (h *BaseHandler) rollbackInternal(_ context.Context, node *structs.ReadNode) error {
	return types.NewError(types.ErrNotImplemented, "rollbackInternal not implemented", nil)
}

// Cancel cancels node execution
func (h *BaseHandler) Cancel(_ context.Context, nodeID string) error {
	taskInfo, ok := h.tasks.Load(nodeID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info := taskInfo.(*TaskInfo)
	info.Status = string(types.ExecutionCancelled)

	if success := h.taskQueue.Cancel(nodeID); !success {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	h.tasks.Delete(nodeID)

	return nil
}

// Task management methods
func (h *BaseHandler) trackTask(taskID string) {
	h.tasks.Store(taskID, &TaskInfo{
		ID:        taskID,
		StartTime: time.Now(),
		Status:    string(types.ExecutionPending),
	})
}

func (h *BaseHandler) updateTaskStatus(taskID string, status string, err error) {
	if task, ok := h.tasks.Load(taskID); ok {
		info := task.(*TaskInfo)
		info.Status = status
		info.Error = err
		h.tasks.Store(taskID, info)
	}
}

func (h *BaseHandler) cleanupTask(taskID string) {
	h.tasks.Delete(taskID)
}

// Error handling methods
func (h *BaseHandler) recordError(err error) {
	if err == nil {
		return
	}

	h.mu.Lock()
	h.lastError = err
	h.mu.Unlock()

	// Publish error event
	h.em.PublishEvent("handler.error", map[string]any{
		"handler_type": h.handlerType,
		"handler_name": h.name,
		"error":        err.Error(),
		"time":         time.Now(),
	})
}

// GetLastError returns the last error
func (h *BaseHandler) GetLastError() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastError
}

// GetResourceUsage returns current resource usage
func (h *BaseHandler) GetResourceUsage() *types.ResourceUsage {
	return &types.ResourceUsage{
		ActiveJobs:     h.workPool.GetMetrics()["active_workers"],
		GoroutineCount: runtime.NumGoroutine(),
	}
}

// GetState returns handler state info
func (h *BaseHandler) GetState() *types.HandlerState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return &types.HandlerState{
		"name":        h.name,
		"type":        h.handlerType,
		"status":      h.status,
		"active_jobs": h.workPool.GetMetrics()["active_workers"],
		"resources":   h.GetResourceUsage(),
		"last_error":  h.lastError,
		"metrics":     h.GetMetrics(),
		"extra":       make(map[string]any),
	}
}
