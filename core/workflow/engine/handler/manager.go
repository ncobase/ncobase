package handler

import (
	"context"
	"fmt"
	"ncobase/common/extension"
	"ncobase/common/logger"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"sync"
	"time"
)

// Manager manages workflow handlers
type Manager struct {
	// Dependencies
	services *service.Service
	em       *extension.Manager
	logger   logger.Logger

	// Runtime components
	metrics *metrics.Collector

	// Handler registry
	handlers map[types.HandlerType]types.Handler
	mu       sync.RWMutex

	// Configuration
	cfg *config.Config

	// Resource tracking
	activeHandlers sync.Map // handlerID -> *HandlerInfo
	handlerCounts  sync.Map // handlerType -> int

	// Runtime state
	status types.HandlerStatus
	ctx    context.Context
	cancel context.CancelFunc
}

// HandlerInfo tracks handler runtime information
type HandlerInfo struct {
	ID          string
	Type        types.HandlerType
	Status      types.HandlerStatus
	StartTime   time.Time
	LastActive  time.Time
	ExecCount   int64
	ErrorCount  int64
	ResourceUse *types.ResourceUsage
	mu          sync.RWMutex
}

// NewManager creates a new handler manager
func NewManager(svc *service.Service, em *extension.Manager, cfg *config.Config) *Manager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		services: svc,
		em:       em,
		cfg:      cfg,
		handlers: make(map[types.HandlerType]types.Handler),
		ctx:      ctx,
		cancel:   cancel,
		status:   types.HandlerReady,
	}

	// Initialize metrics if enabled
	if cfg.Components.Metrics.Enabled {
		m.initMetrics(cfg.Components.Metrics)
	}

	// Register built-in handlers
	m.registerBuiltinHandlers()

	return m
}

// Start starts all handlers
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != types.HandlerReady {
		return types.NewError(types.ErrInvalidStatus, "manager not in ready state", nil)
	}

	// Start each handler in order
	startOrder := []types.HandlerType{
		types.ServiceHandler,
		types.ApprovalHandler,
		types.NotificationHandler,
		types.ScriptHandler,
		types.TimerHandler,
	}

	for _, hType := range startOrder {
		handler := m.handlers[hType]
		if err := handler.Start(); err != nil {
			// Rollback started handlers
			m.stopHandlers()
			return fmt.Errorf("failed to start %s handler: %w", hType, err)
		}
	}

	m.status = types.HandlerRunning
	return nil
}

// Stop stops all handlers
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.status != types.HandlerRunning {
		return nil
	}

	// Cancel context
	m.cancel()

	// Stop handlers
	if err := m.stopHandlers(); err != nil {
		return fmt.Errorf("stop handlers failed: %w", err)
	}

	m.status = types.HandlerStopped
	return nil
}

// GetHandler gets a handler by type
func (m *Manager) GetHandler(handlerType types.HandlerType) (types.Handler, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handler, exists := m.handlers[handlerType]
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "handler not found", nil)
	}

	return handler, nil
}

// RegisterHandler registers a new handler
func (m *Manager) RegisterHandler(handler types.Handler) error {
	if handler == nil {
		return types.NewError(types.ErrInvalidParam, "handler is nil", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	handlerType := handler.Type()
	if _, exists := m.handlers[handlerType]; exists {
		return types.NewError(types.ErrConflict, "handler already registered", nil)
	}

	m.handlers[handlerType] = handler
	return nil
}

// UnregisterHandler unregisters a handler
func (m *Manager) UnregisterHandler(handlerType types.HandlerType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	handler, exists := m.handlers[handlerType]
	if !exists {
		return types.NewError(types.ErrNotFound, "handler not found", nil)
	}

	// Stop handler if running
	if handler.Status() == types.HandlerRunning {
		if err := handler.Stop(); err != nil {
			return fmt.Errorf("stop handler failed: %w", err)
		}
	}

	delete(m.handlers, handlerType)
	return nil
}

// ListHandlers lists all registered handlers
func (m *Manager) ListHandlers() []types.HandlerType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var handlers []types.HandlerType
	for hType := range m.handlers {
		handlers = append(handlers, hType)
	}
	return handlers
}

// IsHealthy checks if all handlers are healthy
func (m *Manager) IsHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.status != types.HandlerRunning {
		return false
	}

	for _, handler := range m.handlers {
		if !handler.IsHealthy() {
			return false
		}
	}
	return true
}

// GetMetrics returns handler metrics
func (m *Manager) GetMetrics() map[string]map[string]any {
	metrics := make(map[string]map[string]any)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for hType, handler := range m.handlers {
		metrics[string(hType)] = handler.GetMetrics()
	}

	return metrics
}

// Internal methods

// stopHandlers stops all handlers in reverse order
func (m *Manager) stopHandlers() error {
	stopOrder := []types.HandlerType{
		types.TimerHandler,
		types.ScriptHandler,
		types.NotificationHandler,
		types.ApprovalHandler,
		types.ServiceHandler,
	}

	var errs []error
	for _, hType := range stopOrder {
		if handler, exists := m.handlers[hType]; exists {
			if err := handler.Stop(); err != nil {
				errs = append(errs, fmt.Errorf("stop %s handler failed: %w", hType, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("stop handlers failed with %d errors: %v", len(errs), errs)
	}
	return nil
}

// initMetrics initializes metrics collection
func (m *Manager) initMetrics(cfg *metrics.Config) error {
	collector, err := metrics.NewCollector(cfg)
	if err != nil {
		return fmt.Errorf("create metrics collector failed: %w", err)
	}
	m.metrics = collector

	// Register metrics
	m.metrics.RegisterCounter("handler_total")
	m.metrics.RegisterCounter("handler_error")
	m.metrics.RegisterCounter("handler_timeout")
	m.metrics.RegisterGauge("handlers_active")
	m.metrics.RegisterHistogram("handler_execution_time", 1000)

	return nil
}

// registerBuiltinHandlers registers built-in handlers
func (m *Manager) registerBuiltinHandlers() {
	// Register approval handler
	approvalHandler, err := NewApprovalHandler(m.services, m.em, m.cfg)
	if err == nil {
		m.handlers[types.ApprovalHandler] = approvalHandler
	}

	// Register service handler
	// serviceHandler, err := NewServiceHandler(m.services, m.em, m.cfg)
	// if err == nil {
	// 	m.handlers[types.ServiceHandler] = serviceHandler
	// }

	// Register notification handler
	m.handlers[types.NotificationHandler] = NewNotificationHandler(m.services, m.em, m.cfg)

	// Register script handler
	m.handlers[types.ScriptHandler] = NewScriptHandler(m.services, m.em, m.cfg)

	// Register timer handler
	m.handlers[types.TimerHandler] = NewTimerHandler(m.services, m.em, m.cfg)
}

// trackHandler tracks handler execution
func (m *Manager) trackHandler(handlerType types.HandlerType, info *HandlerInfo) {
	// Update active count
	count := int64(0)
	if v, ok := m.handlerCounts.Load(handlerType); ok {
		count = v.(int64)
	}
	m.handlerCounts.Store(handlerType, count+1)

	// Store handler info
	m.activeHandlers.Store(info.ID, info)

	// Update metrics
	if m.metrics != nil {
		m.metrics.AddCounter("handler_total", 1)
		m.metrics.SetGauge("handlers_active", float64(count+1))
	}
}

// untrackHandler removes handler tracking
func (m *Manager) untrackHandler(handlerType types.HandlerType, handlerID string) {
	// Update active count
	count := int64(0)
	if v, ok := m.handlerCounts.Load(handlerType); ok {
		count = v.(int64)
	}
	if count > 0 {
		m.handlerCounts.Store(handlerType, count-1)
	}

	// Remove handler info
	m.activeHandlers.Delete(handlerID)

	// Update metrics
	if m.metrics != nil {
		m.metrics.SetGauge("handlers_active", float64(count-1))
	}
}

// getHandlerInfo gets handler runtime info
func (m *Manager) getHandlerInfo(handlerID string) (*HandlerInfo, bool) {
	value, ok := m.activeHandlers.Load(handlerID)
	if !ok {
		return nil, false
	}
	return value.(*HandlerInfo), true
}

// checkResourceLimits checks handler resource limits
func (m *Manager) checkResourceLimits(handlerType types.HandlerType) error {
	count := int64(0)
	if v, ok := m.handlerCounts.Load(handlerType); ok {
		count = v.(int64)
	}

	// Check against configured limits
	if m.cfg != nil && count >= int64(m.cfg.Engine.MaxHandlers) {
		return types.NewError(types.ErrResourceExhausted, "handler limit exceeded", nil)
	}

	return nil
}

// resetHandler resets a handler to initial state
func (m *Manager) resetHandler(handlerType types.HandlerType) error {
	handler, exists := m.handlers[handlerType]
	if !exists {
		return types.NewError(types.ErrNotFound, "handler not found", nil)
	}

	// Stop handler if running
	if handler.Status() == types.HandlerRunning {
		if err := handler.Stop(); err != nil {
			return fmt.Errorf("stop handler failed: %w", err)
		}
	}

	// Clear tracking
	m.handlerCounts.Store(handlerType, int64(0))
	m.activeHandlers.Range(func(key, value any) bool {
		info := value.(*HandlerInfo)
		if info.Type == handlerType {
			m.activeHandlers.Delete(key)
		}
		return true
	})

	// Reset metrics
	if m.metrics != nil {
		// TODO: Reset handler specific metrics
		// prefix := string(handlerType) + "_"
		m.metrics.Reset()
	}

	return nil
}

// Validate validates handler configuration
func (m *Manager) Validate(handlerType types.HandlerType, config any) error {
	handler, exists := m.handlers[handlerType]
	if !exists {
		return types.NewError(types.ErrNotFound, "handler not found", nil)
	}

	// Get handler capabilities
	caps := handler.GetCapabilities()

	// Validate against capabilities
	if err := m.validateAgainstCapabilities(config, caps); err != nil {
		return err
	}

	return nil
}

// validateAgainstCapabilities validates config against handler capabilities
func (m *Manager) validateAgainstCapabilities(config any, caps *types.HandlerCapabilities) error {
	// Add validation logic based on capabilities
	return nil
}
