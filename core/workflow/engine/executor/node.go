package executor

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"sync"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/validation/expression"

	"ncobase/core/workflow/engine/handler"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/logging/logger"

	"github.com/jinzhu/copier"
)

// NodeExecutor handles node execution
type NodeExecutor struct {
	*BaseExecutor

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	handlers *handler.Manager
	logger   logger.Logger

	// Runtime components
	expression *expression.Expression
	metrics    *metrics.Collector
	retries    *RetryExecutor

	// Node tracking
	nodes       sync.Map // nodeID -> *NodeInfo
	activeNodes sync.Map
	joinNodes   sync.Map // nodeID -> *JoinContext

	// Configuration
	config *config.NodeExecutorConfig

	// Internal state
	status types.ExecutionStatus
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NodeInfo tracks node execution state
type NodeInfo struct {
	ID         string
	Type       types.NodeType
	Status     types.ExecutionStatus
	StartTime  time.Time
	EndTime    *time.Time
	RetryCount int
	Error      error
	Variables  map[string]any
	mu         sync.RWMutex
}

// JoinContext tracks join node status
type JoinContext struct {
	NodeID    string
	Required  int
	Completed int
	Error     error
	mu        sync.RWMutex
}

// NewNodeExecutor creates a new node executor
func NewNodeExecutor(svc *service.Service, em ext.ManagerInterface, cfg *config.Config) (*NodeExecutor, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &NodeExecutor{
		BaseExecutor: NewBaseExecutor(types.NodeExecutor, "Node Executor", cfg),
		services:     svc,
		em:           em,
		config:       cfg.Executors.Node,
		ctx:          ctx,
		cancel:       cancel,
		status:       types.ExecutionPending,
	}

	// Initialize metrics if enabled
	if executor.config.EnableMetrics {
		collector, err := metrics.NewCollector(cfg.Components.Metrics)
		if err != nil {
			return nil, fmt.Errorf("create metrics collector failed: %w", err)
		}
		executor.metrics = collector
	}

	// Initialize expression
	executor.expression = expression.NewExpression(cfg.Components.Expression)

	// Initialize retry executor
	retryExecutor, err := NewRetryExecutor(svc, em, cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize retry executor failed: %w", err)
	}
	executor.retries = retryExecutor

	return executor, nil
}

// Start starts the node executor
func (e *NodeExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "executor not in pending state", nil)
	}

	// Start retry executor
	if err := e.retries.Start(); err != nil {
		return fmt.Errorf("start retry executor failed: %w", err)
	}

	// Initialize metrics
	if e.config.EnableMetrics {
		if err := e.initMetrics(); err != nil {
			return fmt.Errorf("initialize metrics failed: %w", err)
		}
	}

	e.status = types.ExecutionActive
	return nil
}

// Stop stops the node executor
func (e *NodeExecutor) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionActive {
		return nil
	}

	// Cancel context
	e.cancel()

	// Stop retry executor
	if err := e.retries.Stop(); err != nil {
		return fmt.Errorf("stop retry executor failed: %w", err)
	}

	// Wait for all goroutines
	e.wg.Wait()

	e.status = types.ExecutionStopped
	return nil
}

// ExecuteNode executes a node
func (e *NodeExecutor) ExecuteNode(ctx context.Context, node *structs.ReadNode) error {
	// Run pre-execute hook
	if e.config.Hooks.BeforeExecute != nil {
		if err := e.config.Hooks.BeforeExecute(ctx, node); err != nil {
			return e.handleNodeError(ctx, node, err)
		}
	}

	// Create node info
	info := &NodeInfo{
		ID:        node.ID,
		Type:      types.NodeType(node.Type),
		Status:    types.ExecutionActive,
		StartTime: time.Now(),
		Variables: node.Variables,
	}
	e.nodes.Store(node.ID, info)
	e.activeNodes.Store(node.ID, info)
	defer func() {
		e.nodes.Delete(node.ID)
		e.activeNodes.Delete(node.ID)
	}()

	// Get node handler
	h, err := e.handlers.GetHandler(types.HandlerType(node.Type))
	if err != nil {
		return e.handleNodeError(ctx, node, err)
	}

	// Execute with retry
	err = e.retries.ExecuteWithRetry(ctx, node.ID, func(retryCtx context.Context) error {
		// Execute node
		n := &structs.ReadNode{}
		err := copier.CopyWithOption(n, node, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			return fmt.Errorf("copy node failed: %w", err)
		}
		err = h.Execute(retryCtx, n)
		return err
	})

	if err != nil {
		return e.handleNodeError(ctx, node, err)
	}

	// Update node info
	info.mu.Lock()
	info.Status = types.ExecutionCompleted
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Run post-execute hook
	if e.config.Hooks.AfterExecute != nil {
		e.config.Hooks.AfterExecute(ctx, node, nil)
	}

	// Auto complete if configured
	if e.config.AutoComplete {
		return e.CompleteNode(ctx, node.ID)
	}

	return nil
}

// CompleteNode completes a node
func (e *NodeExecutor) CompleteNode(ctx context.Context, nodeID string) error {
	// Get node info
	info, ok := e.nodes.Load(nodeID)
	if !ok {
		return types.NewError(types.ErrNotFound, "node not found", nil)
	}
	nodeInfo := info.(*NodeInfo)

	nodeInfo.mu.Lock()
	if nodeInfo.Status != types.ExecutionActive && nodeInfo.Status != types.ExecutionCompleted {
		nodeInfo.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "invalid node status", nil)
	}
	nodeInfo.mu.Unlock()

	// Get node
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		NodeKey: nodeID,
	})
	if err != nil {
		return fmt.Errorf("get node failed: %w", err)
	}

	// Run pre-complete hook
	if e.config.Hooks.BeforeComplete != nil {
		if err := e.config.Hooks.BeforeComplete(ctx, node); err != nil {
			return err
		}
	}

	// Update node status
	_, err = e.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: nodeID,
		NodeBody: structs.NodeBody{
			Status: string(types.ExecutionCompleted),
		},
	})
	if err != nil {
		return fmt.Errorf("update node failed: %w", err)
	}

	// Run post-complete hook
	if e.config.Hooks.AfterComplete != nil {
		e.config.Hooks.AfterComplete(ctx, node, nil)
	}

	// Execute next nodes
	nextNodes, err := e.GetNextNodes(ctx, node)
	if err != nil {
		return fmt.Errorf("get next nodes failed: %w", err)
	}

	for _, next := range nextNodes {
		if err := e.ExecuteNode(ctx, next); err != nil {
			return fmt.Errorf("execute next node failed: %w", err)
		}
	}

	return nil
}

// CancelNode cancels a node
func (e *NodeExecutor) CancelNode(ctx context.Context, nodeID string) error {
	// Get node info
	info, ok := e.nodes.Load(nodeID)
	if !ok {
		return types.NewError(types.ErrNotFound, "node not found", nil)
	}
	nodeInfo := info.(*NodeInfo)

	nodeInfo.mu.Lock()
	if nodeInfo.Status != types.ExecutionActive {
		nodeInfo.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "node not active", nil)
	}

	// Update status
	nodeInfo.Status = types.ExecutionCancelled
	now := time.Now()
	nodeInfo.EndTime = &now
	nodeInfo.mu.Unlock()

	// Get node
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		NodeKey: nodeID,
	})
	if err != nil {
		return fmt.Errorf("get node failed: %w", err)
	}

	// Get handler
	h, err := e.handlers.GetHandler(types.HandlerType(node.Type))
	if err != nil {
		return fmt.Errorf("get handler failed: %w", err)
	}

	// Cancel handler execution
	if err := h.Cancel(ctx, nodeID); err != nil {
		return fmt.Errorf("cancel handler failed: %w", err)
	}

	// Update node status
	_, err = e.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: nodeID,
		NodeBody: structs.NodeBody{
			Status: string(types.ExecutionCancelled),
		},
	})

	return err
}

// RollbackNode rolls back a node
func (e *NodeExecutor) RollbackNode(ctx context.Context, nodeID string) error {
	// Run pre-rollback hook
	if e.config.Hooks.BeforeRollback != nil {
		if err := e.config.Hooks.BeforeRollback(ctx, nil); err != nil {
			return err
		}
	}

	// Get node
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		NodeKey: nodeID,
	})
	if err != nil {
		return fmt.Errorf("get node failed: %w", err)
	}

	// Get handler
	h, err := e.handlers.GetHandler(types.HandlerType(node.Type))
	if err != nil {
		return fmt.Errorf("get handler failed: %w", err)
	}

	// Execute rollback
	err = h.Rollback(ctx, node)

	// Run post-rollback hook
	if e.config.Hooks.AfterRollback != nil {
		e.config.Hooks.AfterRollback(ctx, node, err)
	}

	if err != nil {
		return fmt.Errorf("rollback handler failed: %w", err)
	}

	// Update node status
	_, err = e.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: nodeID,
		NodeBody: structs.NodeBody{
			Status: string(types.ExecutionRollbacked),
		},
	})

	return err
}

// GetNodeInfo gets node execution info
func (e *NodeExecutor) GetNodeInfo(nodeID string) (*NodeInfo, bool) {
	info, ok := e.nodes.Load(nodeID)
	if !ok {
		return nil, false
	}
	return info.(*NodeInfo), true
}

// GetActiveNodes gets all active nodes
func (e *NodeExecutor) GetActiveNodes() []*NodeInfo {
	var active []*NodeInfo
	e.activeNodes.Range(func(_, value any) bool {
		info := value.(*NodeInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive {
			active = append(active, info)
		}
		info.mu.RUnlock()
		return true
	})
	return active
}

// GetNodeStatus gets node status
func (e *NodeExecutor) GetNodeStatus(nodeID string) types.ExecutionStatus {
	info, ok := e.nodes.Load(nodeID)
	if !ok {
		return types.ExecutionPending
	}
	nodeInfo := info.(*NodeInfo)
	nodeInfo.mu.RLock()
	defer nodeInfo.mu.RUnlock()

	return nodeInfo.Status
}

// HandleJoinNode handles join node execution
func (e *NodeExecutor) HandleJoinNode(ctx context.Context, node *structs.ReadNode) error {
	if !e.config.TrackJoinStatus {
		return nil
	}

	// Get or create join context
	joinCtx := &JoinContext{
		NodeID: node.ID,
	}

	// Get required join count from node config
	if required, ok := node.Properties["required"].(int); ok {
		joinCtx.Required = required
	} else {
		// Default to all incoming paths
		incoming, _ := node.Properties["incoming"].([]any)
		joinCtx.Required = len(incoming)
	}

	value, loaded := e.joinNodes.LoadOrStore(node.ID, joinCtx)
	if loaded {
		joinCtx = value.(*JoinContext)
	}

	// Update join status
	joinCtx.mu.Lock()
	joinCtx.Completed++
	complete := joinCtx.Completed >= joinCtx.Required
	joinCtx.mu.Unlock()

	if !complete {
		return nil // Wait for other branches
	}

	// Get next nodes
	nextNodes, err := e.GetNextNodes(ctx, node)
	if err != nil {
		return fmt.Errorf("get next nodes failed: %w", err)
	}

	// Execute next nodes
	for _, next := range nextNodes {
		if err := e.ExecuteNode(ctx, next); err != nil {
			return fmt.Errorf("execute next node failed: %w", err)
		}
	}

	// Clean up join context
	e.joinNodes.Delete(node.ID)

	return nil
}

// GetNextNodes gets next nodes to execute
func (e *NodeExecutor) GetNextNodes(ctx context.Context, node *structs.ReadNode) ([]*structs.ReadNode, error) {
	var nextNodes []*structs.ReadNode

	switch node.Type {
	case string(types.NodeExclusive):
		// Get next node based on conditions
		next, err := e.getExclusiveNextNode(ctx, node)
		if err != nil {
			return nil, err
		}
		if next != nil {
			nextNodes = append(nextNodes, next)
		}

	case string(types.NodeParallel):
		// Execute all parallel branches
		for _, nodeKey := range node.ParallelNodes {
			next, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: node.ProcessID,
				NodeKey:   nodeKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, next)
		}

	default:
		// Get direct next nodes
		for _, nodeKey := range node.NextNodes {
			next, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: node.ProcessID,
				NodeKey:   nodeKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, next)
		}
	}

	return nextNodes, nil
}

// Internal helper methods

func (e *NodeExecutor) handleNodeError(ctx context.Context, node *structs.ReadNode, err error) error {
	// Run error hook
	if e.config.Hooks.OnError != nil {
		e.config.Hooks.OnError(ctx, node, err)
	}

	// Update metrics
	if e.metrics != nil {
		e.metrics.AddCounter("node_error", 1)
	}

	// Update node status
	_, updateErr := e.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(types.ExecutionError),
			Properties: map[string]any{
				"error":      err.Error(),
				"error_time": time.Now(),
			},
		},
	})
	if updateErr != nil {
		return fmt.Errorf("update node failed: %v (original error: %v)", updateErr, err)
	}

	// Publish error event
	e.em.PublishEvent(string(types.EventNodeFailed), &types.Event{
		Type:      types.EventNodeFailed,
		ProcessID: node.ProcessID,
		NodeID:    node.ID,
		Details: map[string]any{
			"error": err.Error(),
		},
		Timestamp: time.Now(),
	})

	return err
}

// getExclusiveNextNode gets the next node for an exclusive gateway
func (e *NodeExecutor) getExclusiveNextNode(ctx context.Context, node *structs.ReadNode) (*structs.ReadNode, error) {
	conditions, ok := node.Properties["conditions"].([]any)
	if !ok {
		return nil, types.NewError(types.ErrValidationFailed, "missing conditions", nil)
	}

	// Evaluate conditions in priority order
	for _, c := range conditions {
		condition, ok := c.(map[string]any)
		if !ok {
			continue
		}

		expr, ok := condition["expression"].(string)
		if !ok {
			continue
		}

		nextKey, ok := condition["next_node"].(string)
		if !ok {
			continue
		}

		// Evaluate condition
		result, err := e.evaluateCondition(ctx, expr, node.Variables)
		if err != nil {
			e.logger.Warn(ctx, "evaluate condition failed", "error", err)
			continue
		}

		if result {
			return e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: node.ProcessID,
				NodeKey:   nextKey,
			})
		}
	}

	// Use default path if no condition matches
	if defaultPath, ok := node.Properties["default_path"].(string); ok {
		return e.services.Node.Get(ctx, &structs.FindNodeParams{
			ProcessID: node.ProcessID,
			NodeKey:   defaultPath,
		})
	}

	return nil, nil
}

// evaluateCondition evaluates a condition expression
func (e *NodeExecutor) evaluateCondition(ctx context.Context, expr string, vars map[string]any) (bool, error) {
	// Evaluate condition expression
	result, err := e.expression.Evaluate(ctx, expr, vars)
	if err != nil {
		return false, fmt.Errorf("evaluate expression failed: %w", err)
	}

	// Convert result to boolean
	switch v := result.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case string:
		return v != "", nil
	default:
		return false, fmt.Errorf("invalid condition result type: %T", result)
	}
}

// handleJoinNode handles a join node
func (e *NodeExecutor) handleJoinNode(ctx context.Context, node *structs.ReadNode) error {
	if !e.config.TrackJoinStatus {
		return nil
	}

	// Get or create join context
	joinCtx := &JoinContext{
		NodeID: node.ID,
	}

	// Get required join count from node config
	if required, ok := node.Properties["required"].(int); ok {
		joinCtx.Required = required
	} else {
		// Default to all incoming paths
		incoming, _ := node.Properties["incoming"].([]any)
		joinCtx.Required = len(incoming)
	}

	value, loaded := e.joinNodes.LoadOrStore(node.ID, joinCtx)
	if loaded {
		joinCtx = value.(*JoinContext)
	}

	// Update join status
	joinCtx.mu.Lock()
	joinCtx.Completed++
	complete := joinCtx.Completed >= joinCtx.Required
	joinCtx.mu.Unlock()

	if !complete {
		return nil // Wait for other branches
	}

	// Get next nodes
	nextNodes, err := e.GetNextNodes(ctx, node)
	if err != nil {
		return fmt.Errorf("get next nodes failed: %w", err)
	}

	// Execute next nodes
	for _, next := range nextNodes {
		if err := e.ExecuteNode(ctx, next); err != nil {
			return fmt.Errorf("execute next node failed: %w", err)
		}
	}

	// Clean up join context
	e.joinNodes.Delete(node.ID)

	return nil
}

// initMetrics initializes metrics
func (e *NodeExecutor) initMetrics() error {
	// Register node execution metrics
	e.metrics.RegisterCounter("node_total")
	e.metrics.RegisterCounter("node_success")
	e.metrics.RegisterCounter("node_error")
	e.metrics.RegisterCounter("node_timeout")
	e.metrics.RegisterCounter("node_retry")

	// Register node state metrics
	e.metrics.RegisterGauge("nodes_active")
	e.metrics.RegisterGauge("nodes_pending")
	e.metrics.RegisterGauge("nodes_completed")

	// Register performance metrics
	e.metrics.RegisterHistogram("node_execution_time", 1000)
	e.metrics.RegisterHistogram("node_queue_time", 1000)

	return nil
}

// GetMetrics returns executor metrics
func (e *NodeExecutor) GetMetrics() map[string]any {
	if !e.config.EnableMetrics {
		return nil
	}

	m := make(map[string]any)

	// Node execution metrics
	m["node_total"] = e.metrics.GetCounter("node_total")
	m["node_success"] = e.metrics.GetCounter("node_success")
	m["node_error"] = e.metrics.GetCounter("node_error")
	m["node_timeout"] = e.metrics.GetCounter("node_timeout")
	m["node_retry"] = e.metrics.GetCounter("node_retry")

	// Node state metrics
	m["nodes_active"] = e.metrics.GetGauge("nodes_active")
	m["nodes_pending"] = e.metrics.GetGauge("nodes_pending")
	m["nodes_completed"] = e.metrics.GetGauge("nodes_completed")

	// Performance metrics
	m["node_execution_time"] = e.metrics.GetHistogram("node_execution_time")
	m["node_queue_time"] = e.metrics.GetHistogram("node_queue_time")

	return m
}

// GetCapabilities returns executor capabilities
func (e *NodeExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    true,
		SupportsRetry:    true,
		SupportsRollback: true,
		MaxConcurrency:   e.config.ConcurrentNodes,
		MaxBatchSize:     e.config.BufferSize,
		AllowedActions: []string{
			"execute",
			"complete",
			"cancel",
			"rollback",
		},
	}
}

// IsHealthy checks if executor is healthy
func (e *NodeExecutor) IsHealthy() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status == types.ExecutionActive
}

// Status return this executor status
func (e *NodeExecutor) Status() types.ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}
