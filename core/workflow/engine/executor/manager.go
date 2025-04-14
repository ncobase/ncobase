package executor

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"sync"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Manager manages all executors
type Manager struct {
	executors map[types.ExecutorType]types.Executor
	mu        sync.RWMutex

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	logger   logger.Logger
	cfg      *config.Config

	// Runtime state
	status types.ExecutionStatus
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager creates a new executor manager
func NewManager(cfg *config.Config, svc *service.Service, em ext.ManagerInterface) (*Manager, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		executors: make(map[types.ExecutorType]types.Executor),
		services:  svc,
		em:        em,
		cfg:       cfg,
		ctx:       ctx,
		cancel:    cancel,
		status:    types.ExecutionPending,
	}

	// Initialize executors
	if err := m.initExecutors(); err != nil {
		cancel()
		return nil, fmt.Errorf("init executors failed: %w", err)
	}

	return m, nil
}

// Start starts all executors
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "manager not in pending state", nil)
	}

	// Start executors in order
	order := []types.ExecutorType{
		types.RetryExecutor,   // Start retry first
		types.ServiceExecutor, // Then service
		types.NodeExecutor,    // Then node
		types.TaskExecutor,    // Then task
		types.ProcessExecutor, // Process last
	}

	for _, execType := range order {
		executor := m.executors[execType]
		if err := executor.Start(); err != nil {
			err := m.stopExecutors()
			if err != nil {
				return err
			} // Stop started ones
			return fmt.Errorf("start %s executor failed: %w", execType, err)
		}
	}

	m.status = types.ExecutionActive
	return nil
}

// Stop stops all executors
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != types.ExecutionActive {
		return nil
	}

	m.cancel()
	if err := m.stopExecutors(); err != nil {
		return fmt.Errorf("stop executors failed: %w", err)
	}

	m.status = types.ExecutionStopped
	return nil
}

// GetExecutor gets an executor by type
func (m *Manager) GetExecutor(execType types.ExecutorType) (types.Executor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	executor, exists := m.executors[execType]
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "executor not found", nil)
	}

	return executor, nil
}

// GetStatus gets manager status
func (m *Manager) GetStatus() types.ExecutionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// IsHealthy checks if all executors are healthy
func (m *Manager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.status != types.ExecutionActive {
		return false
	}

	for _, executor := range m.executors {
		if !executor.IsHealthy() {
			return false
		}
	}

	return true
}

// GetMetrics gets metrics from all executors
func (m *Manager) GetMetrics() map[string]map[string]any {
	metrics := make(map[string]map[string]any)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for execType, executor := range m.executors {
		metrics[string(execType)] = executor.GetMetrics()
	}

	return metrics
}

// Internal methods

// initExecutors initializes all executors
func (m *Manager) initExecutors() error {
	// Create retry executor
	retryExecutor, err := NewRetryExecutor(m.services, m.em, m.cfg)
	if err != nil {
		return fmt.Errorf("create retry executor failed: %w", err)
	}
	m.executors[types.RetryExecutor] = retryExecutor

	// Create service executor
	serviceExecutor, err := NewServiceExecutor(m.services, m.em, m.cfg)
	if err != nil {
		return fmt.Errorf("create service executor failed: %w", err)
	}
	m.executors[types.ServiceExecutor] = serviceExecutor

	// Create node executor
	nodeExecutor, err := NewNodeExecutor(m.services, m.em, m.cfg)
	if err != nil {
		return fmt.Errorf("create node executor failed: %w", err)
	}
	m.executors[types.NodeExecutor] = nodeExecutor

	// Create task executor
	taskExecutor, err := NewTaskExecutor(m.services, m.em, m.cfg)
	if err != nil {
		return fmt.Errorf("create task executor failed: %w", err)
	}
	m.executors[types.TaskExecutor] = taskExecutor

	// Create process executor
	processExecutor, err := NewProcessExecutor(m.services, m.em, m.cfg)
	if err != nil {
		return fmt.Errorf("create process executor failed: %w", err)
	}

	// Set process executor's dependent executors
	processExecutor.SetDependentExecutors(nodeExecutor, taskExecutor)

	m.executors[types.ProcessExecutor] = processExecutor

	return nil
}

// stopExecutors stops all executors in reverse order
func (m *Manager) stopExecutors() error {
	order := []types.ExecutorType{
		types.ProcessExecutor,
		types.TaskExecutor,
		types.NodeExecutor,
		types.ServiceExecutor,
		types.RetryExecutor,
	}

	var errs []error
	for _, execType := range order {
		if executor, exists := m.executors[execType]; exists {
			if err := executor.Stop(); err != nil {
				errs = append(errs, fmt.Errorf("stop %s executor failed: %w", execType, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("stop executors failed with %d errors: %v", len(errs), errs)
	}
	return nil
}

// Reset resets all executors
func (m *Manager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status == types.ExecutionActive {
		return types.NewError(types.ErrInvalidParam, "cannot reset while active", nil)
	}

	for execType, executor := range m.executors {
		if err := executor.Reset(); err != nil {
			return fmt.Errorf("reset %s executor failed: %w", execType, err)
		}
	}

	m.status = types.ExecutionPending
	return nil
}

// ResourceUsage returns resource usage of all executors
func (m *Manager) ResourceUsage() map[types.ExecutorType]*types.ResourceUsage {
	usage := make(map[types.ExecutorType]*types.ResourceUsage)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for execType, executor := range m.executors {
		metrics := executor.GetMetrics()
		if metrics == nil {
			continue
		}

		// Extract resource metrics
		usage[execType] = &types.ResourceUsage{
			MemoryUsage:    metrics["memory_usage"].(int64),
			CPUUsage:       metrics["cpu_usage"].(float64),
			GoroutineCount: metrics["goroutine_count"].(int),
			ActiveJobs:     metrics["active_jobs"].(int64),
		}
	}

	return usage
}
