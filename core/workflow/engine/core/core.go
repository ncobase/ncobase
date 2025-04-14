package core

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/coordinator"
	"ncobase/core/workflow/engine/executor"
	"ncobase/core/workflow/engine/handler"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"runtime"
	"sync"
	"time"

	"github.com/ncobase/ncore/concurrency"
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/expression"
)

// Core represents the workflow engine core
type Core struct {
	// Services
	services *service.Service
	// Extension manager
	em ext.ManagerInterface

	// Metrics
	metrics *metrics.Collector
	// Coordinator
	coordinator *coordinator.Coordinator
	// Expression
	expression *expression.Expression

	// Configuration
	cfg *config.Config

	// State manager
	sm *StateManager
	// Concurrency manager
	cm *concurrency.Manager
	// Data flow manager
	dfm *DataFlowManager

	// Executors
	Executors *executor.Manager
	// Handlers
	Handlers *handler.Manager

	// Runtime state
	Status  types.EngineStatus
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running sync.Map

	// Logger
	logger logger.Logger
}

// NewCore creates a new workflow engine core
func NewCore(cfg *config.Config, svc *service.Service, em ext.ManagerInterface) (*Core, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Initialize state manager
	sm := NewStateManager()

	// Initialize concurrency manager
	cm, err := concurrency.NewManager(cfg.Engine.MaxConcurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to create concurrency manager: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	core := &Core{
		services: svc,
		em:       em,
		cfg:      cfg,
		ctx:      ctx,
		cancel:   cancel,
		Status:   types.EngineReady,
		sm:       sm,
		cm:       cm,
	}

	// Initialize data flow manager
	core.dfm = NewDataFlowManager()

	// Initialize components
	if err = core.initComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	return core, nil
}

// Initialize initializes core components
func (c *Core) Initialize() error {
	// Initialize executors
	if err := c.Executors.Start(); err != nil {
		return fmt.Errorf("initialize executors failed: %w", err)
	}

	// Initialize handlers
	if err := c.Handlers.Start(); err != nil {
		return fmt.Errorf("initialize handlers failed: %w", err)
	}

	// Initialize expressions

	c.expression = expression.NewExpression(c.cfg.Components.Expression)

	// Initialize metrics
	if err := c.metrics.Start(c.ctx); err != nil {
		return fmt.Errorf("initialize metrics failed: %w", err)
	}

	return nil
}

// Start starts the engine core
func (c *Core) Start() error {
	if c.Status != types.EngineReady {
		return types.ErrInvalidStatus
	}

	// Start components
	if err := c.Executors.Start(); err != nil {
		return fmt.Errorf("failed to start executors: %w", err)
	}

	if err := c.Handlers.Start(); err != nil {
		return fmt.Errorf("failed to start handlers: %w", err)
	}

	if err := c.metrics.Start(c.ctx); err != nil {
		return fmt.Errorf("failed to start metrics: %w", err)
	}

	c.Status = types.EngineRunning

	// Start background tasks
	c.wg.Add(2)
	go c.processMetrics()
	go c.monitorState()

	return nil
}

// Stop stops the engine core
func (c *Core) Stop() error {
	if c.Status != types.EngineRunning {
		return nil
	}

	c.Status = types.EngineStopped

	// Signal stop
	c.cancel()

	// Stop components
	if err := c.Executors.Stop(); err != nil {
		return fmt.Errorf("failed to stop executors: %w", err)
	}

	if err := c.Handlers.Stop(); err != nil {
		return fmt.Errorf("failed to stop handlers: %w", err)
	}

	c.metrics.Stop()

	// Wait for background tasks
	c.wg.Wait()

	return nil
}

// Execute executes a workflow process
func (c *Core) Execute(ctx context.Context, processID string) error {
	if c.Status != types.EngineRunning {
		return types.ErrNotRunning
	}

	// Get process
	process, err := c.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// Track running process
	c.running.Store(processID, process)
	defer c.running.Delete(processID)

	// Create execution context
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(c.cfg.Engine.DefaultTimeout))
	defer cancel()

	// Get start node
	startNode, err := c.getStartNode(execCtx, process.ID)
	if err != nil {
		return fmt.Errorf("failed to get start node: %w", err)
	}

	// Execute start node
	if err := c.ExecuteNode(execCtx, process, startNode); err != nil {
		if err := c.rollback(execCtx, process); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		return fmt.Errorf("execution failed: %w", err)
	}

	return nil
}

// evaluateExpression evaluates a condition expression
func (c *Core) evaluateExpression(ctx context.Context, expr string, variables map[string]any) (bool, error) {
	result, err := c.expression.Evaluate(ctx, expr, variables)
	if err != nil {
		return false, fmt.Errorf("expression evaluation failed: %w", err)
	}

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
		return false, fmt.Errorf("expression result must be boolean")
	}
}

// CompleteNode handles node completion
func (c *Core) CompleteNode(ctx context.Context, node *structs.ReadNode) error {
	// Get node handler
	_, err := c.Handlers.GetHandler(types.HandlerType(node.Type))
	if err != nil {
		return fmt.Errorf("get handler failed: %w", err)
	}

	// Update node status
	_, err = c.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
			// EndTime: func() *time.Time { t := time.Now(); return &t }(),
		},
	})
	if err != nil {
		return fmt.Errorf("update node status failed: %w", err)
	}

	// Get process
	process, err := c.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: node.ProcessID,
	})
	if err != nil {
		return fmt.Errorf("get process failed: %w", err)
	}

	// Check if process is completed
	if isEndNode := c.isEndNode(node); isEndNode {
		return c.CompleteProcess(ctx, process.ID)
	}

	// Execute next nodes
	nextNodes, err := c.getNextNodes(ctx, process, node)
	if err != nil {
		return fmt.Errorf("get next nodes failed: %w", err)
	}

	for _, next := range nextNodes {
		if err := c.ExecuteNode(ctx, process, next); err != nil {
			return fmt.Errorf("execute next node failed: %w", err)
		}
	}

	return nil
}

// CompleteProcess completes a process
func (c *Core) CompleteProcess(ctx context.Context, processID string) error {
	process, err := c.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// Update process status
	_, err = c.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status:  string(structs.StatusCompleted),
			EndTime: func() *int64 { t := time.Now().UnixMilli(); return &t }(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update process: %w", err)
	}

	// Publish event
	c.em.PublishEvent(structs.EventProcessCompleted, &structs.EventData{
		ProcessID:   process.ID,
		ProcessName: process.ProcessCode,
	})

	return nil
}

// isEndNode checks if a node is an end node
func (c *Core) isEndNode(node *structs.ReadNode) bool {
	return node.Type == string(structs.NodeEnd) || len(node.NextNodes) == 0
}

// executeNodeWithTimeout executes node with timeout control
func (c *Core) executeNodeWithTimeout(ctx context.Context, node *structs.ReadNode) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(c.cfg.Engine.DefaultTimeout)*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- c.ExecuteNode(timeoutCtx, nil, node)
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("node execution timeout")
	}
}

// handleNodeError handles node execution errors
func (c *Core) handleNodeError(ctx context.Context, node *structs.ReadNode, err error) error {
	// Update node status
	_, updateErr := c.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusError),
			Properties: map[string]any{
				"error":      err.Error(),
				"error_time": time.Now(),
			},
		},
	})
	if updateErr != nil {
		return fmt.Errorf("failed to update node status: %w", updateErr)
	}

	// Publish error event
	c.em.PublishEvent(structs.EventNodeError, &structs.EventData{
		ProcessID: node.ProcessID,
		NodeID:    node.NodeKey,
		NodeType:  structs.NodeType(node.Type),
		Details: map[string]any{
			"error": err.Error(),
		},
	})

	return err
}

// validateProcess validates process state
func (c *Core) validateProcess(process *structs.ReadProcess) error {
	if process.Status == string(structs.StatusCompleted) ||
		process.Status == string(structs.StatusTerminated) {
		return fmt.Errorf("process is already ended")
	}

	if process.IsSuspended {
		return fmt.Errorf("process is suspended")
	}

	return nil
}

// ExecuteNode executes a node
func (c *Core) ExecuteNode(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) error {
	if err := c.validateNode(node); err != nil {
		return err
	}

	// Get node handler
	h, err := c.Handlers.GetHandler(types.HandlerType(node.Type))
	if err != nil {
		return fmt.Errorf("get handler failed: %w", err)
	}

	// Pre-execute node
	if err := c.preExecuteNode(ctx, process, node); err != nil {
		return err
	}

	// Execute node
	start := time.Now().UnixMilli()
	err = h.Execute(ctx, node)
	duration := time.Duration(time.Now().UnixMilli() - start)

	c.metrics.RecordDuration("node_duration", duration)
	if err != nil {
		c.metrics.AddCounter("node_failure", 1)
		return c.handleNodeError(ctx, node, err)
	}
	c.metrics.AddCounter("node_success", 1)

	// Post-execute node
	if err := c.postExecuteNode(ctx, process, node); err != nil {
		return err
	}

	return nil
}

// preExecuteNode prepares a node for execution
func (c *Core) preExecuteNode(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) error {
	// update node status
	_, err := c.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusActive),
		},
	})
	if err != nil {
		return fmt.Errorf("update node status failed: %w", err)
	}

	// publish event
	c.em.PublishEvent(structs.EventNodeStarted, &structs.EventData{
		ProcessID: process.ID,
		NodeID:    node.NodeKey,
		NodeType:  structs.NodeType(node.Type),
	})

	return nil
}

// postExecuteNode post-executes a node
func (c *Core) postExecuteNode(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) error {
	// update node status
	_, err := c.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
			// EndTime: func() *time.Time { t := time.Now(); return &t }(),
		},
	})
	if err != nil {
		return fmt.Errorf("update node status failed: %w", err)
	}

	// publish event
	c.em.PublishEvent(structs.EventNodeCompleted, &structs.EventData{
		ProcessID: process.ID,
		NodeID:    node.NodeKey,
		NodeType:  structs.NodeType(node.Type),
	})

	return nil
}

// validateNode validates a node
func (c *Core) validateNode(node *structs.ReadNode) error {
	if node == nil {
		return types.NewError(types.ErrValidation, "node is nil", nil)
	}

	if node.Type == "" {
		return types.NewError(types.ErrValidation, "node type is required", nil)
	}

	if node.ProcessID == "" {
		return types.NewError(types.ErrValidation, "process ID is required", nil)
	}

	return nil
}

// monitorState monitors the state of the engine
func (c *Core) monitorState() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkExecutorHealth()
			c.checkResourceLimits()
			c.cleanupStaleExecutions()
		}
	}
}

// checkExecutorHealth checks the health of the Executors
func (c *Core) checkExecutorHealth() {
	if !c.Executors.IsHealthy() {
		c.logger.Warn(c.ctx, "unhealthy executors detected")
		if err := c.Executors.Reset(); err != nil {
			c.logger.Error(c.ctx, "reset executors failed", err)
		}
	}
}

// checkResourceLimits checks the resource limits
func (c *Core) checkResourceLimits() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	if memStats.Alloc > uint64(c.cfg.Engine.MaxMemory) {
		c.logger.Warn(c.ctx, "memory usage exceeds limit")
	}

	if runtime.NumGoroutine() > c.cfg.Engine.MaxGoroutines {
		c.logger.Warn(c.ctx, "goroutine count exceeds limit")
	}
}

// cleanupStaleExecutions cleans up stale executions
func (c *Core) cleanupStaleExecutions() {
	c.running.Range(func(key, value any) bool {
		process := value.(*structs.ReadProcess)
		if process.StartTime != nil && time.Since(time.UnixMilli(*process.StartTime)) > time.Duration(c.cfg.Engine.DefaultTimeout) {
			c.logger.Warn(c.ctx, "cleanup stale process", "id", process.ID)
			c.running.Delete(key)
		}
		return true
	})
}

func (c *Core) rollback(ctx context.Context, process *structs.ReadProcess) error {
	// Get completed nodes
	nodes, err := c.services.Node.List(ctx, &structs.ListNodeParams{
		ProcessID: process.ID,
		Status:    string(structs.StatusCompleted),
	})
	if err != nil {
		return err
	}

	// Rollback each node in reverse order
	for i := len(nodes.Items) - 1; i >= 0; i-- {
		node := nodes.Items[i]
		h, err := c.Handlers.GetHandler(types.HandlerType(node.Type))
		if err != nil {
			continue
		}

		if err := h.Rollback(ctx, node); err != nil {
			return fmt.Errorf("rollback node %s failed: %w", node.ID, err)
		}
	}

	// Update process status
	_, err = c.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status: string(structs.StatusError),
		},
	})

	return err
}

// ResumeProcesses resumes processes after restart
func (c *Core) ResumeProcesses(ctx context.Context) error {
	// Get all active processes
	processes, err := c.services.Process.List(ctx, &structs.ListProcessParams{
		Status: string(structs.StatusActive),
	})
	if err != nil {
		return fmt.Errorf("list active processes failed: %w", err)
	}

	// Resume each process
	for _, process := range processes.Items {
		if err := c.resumeProcess(ctx, process); err != nil {
			c.logger.Errorf(ctx, "resume process %s failed: %v", process.ID, err)
			continue
		}
	}

	return nil
}

func (c *Core) resumeProcess(ctx context.Context, process *structs.ReadProcess) error {
	// Get current node
	node, err := c.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: process.ID,
		NodeKey:   process.CurrentNode,
	})
	if err != nil {
		return fmt.Errorf("get current node failed: %w", err)
	}

	// Resume from current node
	return c.ExecuteNode(ctx, process, node)
}

// Validate validates workflow definition
func (c *Core) Validate(ctx context.Context, processID string) error {
	// Get process definition
	process, err := c.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("get process failed: %w", err)
	}

	// Get all nodes
	nodes, err := c.services.Node.List(ctx, &structs.ListNodeParams{
		ProcessID: process.ID,
	})
	if err != nil {
		return fmt.Errorf("list nodes failed: %w", err)
	}

	// Validate each node
	for _, node := range nodes.Items {
		if err := c.validateNode(node); err != nil {
			return fmt.Errorf("validate node %s failed: %w", node.ID, err)
		}
	}

	// Validate process structure
	if err := c.validateProcessStructure(process, nodes.Items); err != nil {
		return fmt.Errorf("validate process structure failed: %w", err)
	}

	return nil
}

//	func (c *Core) validateNode(ctx context.Context, node *structs.ReadNode) error {
//		// Get handler
//		handler, err := c.handlers.GetHandler(node.Type)
//		if err != nil {
//			return err
//		}
//
//		// Validate node configuration
//		return handler.Validate(node)
//	}
//
// validateProcessStructure validates process structure
func (c *Core) validateProcessStructure(_ *structs.ReadProcess, nodes []*structs.ReadNode) error {
	// Build node map
	nodeMap := make(map[string]*structs.ReadNode)
	for _, node := range nodes {
		nodeMap[node.NodeKey] = node
	}

	// Validate node connections
	for _, node := range nodes {
		// Check next nodes exist
		for _, nextKey := range node.NextNodes {
			if _, exists := nodeMap[nextKey]; !exists {
				return fmt.Errorf("next node %s not found", nextKey)
			}
		}

		// Check parallel nodes exist
		for _, parallelKey := range node.ParallelNodes {
			if _, exists := nodeMap[parallelKey]; !exists {
				return fmt.Errorf("parallel node %s not found", parallelKey)
			}
		}
	}

	// Check for cycles
	visited := make(map[string]bool)
	path := make(map[string]bool)

	var checkCycle func(nodeKey string) error
	checkCycle = func(nodeKey string) error {
		if path[nodeKey] {
			return fmt.Errorf("cycle detected at node %s", nodeKey)
		}
		if visited[nodeKey] {
			return nil
		}

		visited[nodeKey] = true
		path[nodeKey] = true

		node := nodeMap[nodeKey]
		for _, nextKey := range node.NextNodes {
			if err := checkCycle(nextKey); err != nil {
				return err
			}
		}

		path[nodeKey] = false
		return nil
	}

	// Start cycle check from start nodes
	for _, node := range nodes {
		if node.Type == string(structs.NodeStart) {
			if err := checkCycle(node.NodeKey); err != nil {
				return err
			}
		}
	}

	return nil
}

// Resource management

// Core engine metrics initialization
func (c *Core) initMetrics() error {
	if !c.cfg.Components.Metrics.Enabled {
		return nil
	}

	// Core counters
	c.metrics.RegisterCounter("core_executions_total")
	c.metrics.RegisterCounter("core_executions_success")
	c.metrics.RegisterCounter("core_executions_failed")
	c.metrics.RegisterCounter("core_executions_timeout")
	c.metrics.RegisterCounter("core_executions_cancelled")

	// Core gauges
	c.metrics.RegisterGauge("core_executions_active")
	c.metrics.RegisterGauge("core_memory_usage")
	c.metrics.RegisterGauge("core_cpu_usage")
	c.metrics.RegisterGauge("core_goroutines")

	// Core histograms
	c.metrics.RegisterHistogram("core_execution_time", 1000)
	c.metrics.RegisterHistogram("core_process_time", 1000)

	return nil
}

// GetResourceUsage returns resource usage
func (c *Core) GetResourceUsage() map[string]any {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]any{
		"memory_alloc": memStats.Alloc,
		"memory_total": memStats.TotalAlloc,
		"memory_sys":   memStats.Sys,
		"goroutines":   runtime.NumGoroutine(),
		"gc_cycles":    memStats.NumGC,
	}
}

// Monitor resource limits
func (c *Core) monitorResources() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.checkResourceLimits()
		}
	}
}

// initComponents initializes core components
func (c *Core) initComponents() (err error) {
	// Initialize executors
	c.Executors, err = executor.NewManager(c.cfg, c.services, c.em)
	if err != nil {
		return fmt.Errorf("failed to create executors: %w", err)
	}

	// Initialize handlers
	c.Handlers = handler.NewManager(c.services, c.em, c.cfg)

	// Initialize metrics
	m, err := metrics.NewCollector(c.cfg.Components.Metrics)
	if err != nil {
		return fmt.Errorf("failed to create metrics collector: %w", err)
	}
	c.metrics = m

	if err := c.initMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize expressions
	c.expression = expression.NewExpression(c.cfg.Components.Expression)

	return nil
}

// getStartNode returns start node
func (c *Core) getStartNode(ctx context.Context, processID string) (*structs.ReadNode, error) {
	nodes, err := c.services.Node.List(ctx, &structs.ListNodeParams{
		ProcessID: processID,
		Type:      string(structs.NodeStart),
	})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, types.ErrNotFound
	}

	return nodes.Items[0], nil
}

// getNextNodes returns next nodes
func (c *Core) getNextNodes(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) ([]*structs.ReadNode, error) {
	var nextNodes []*structs.ReadNode

	switch node.Type {
	case string(structs.NodeExclusive):
		// Evaluate conditions to get next node
		next, err := c.evaluateConditions(ctx, process, node)
		if err != nil {
			return nil, err
		}
		nextNodes = append(nextNodes, next)

	case string(structs.NodeParallel):
		// Get all parallel branches
		for _, nodeKey := range node.ParallelNodes {
			next, err := c.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: process.ID,
				NodeKey:   nodeKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, next)
		}

	default:
		// Get single next node
		for _, nodeKey := range node.NextNodes {
			next, err := c.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: process.ID,
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

// evaluateConditions evaluates conditions
func (c *Core) evaluateConditions(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) (*structs.ReadNode, error) {
	// Get conditions from node config
	conditions, ok := node.Properties["conditions"].([]any)
	if !ok {
		return nil, fmt.Errorf("no conditions configured")
	}

	// Evaluate each condition
	for _, cond := range conditions {
		condition, ok := cond.(map[string]any)
		if !ok {
			continue
		}

		expr, _ := condition["expression"].(string)
		match, err := c.evaluateExpression(ctx, expr, process.Variables)
		if err != nil {
			continue
		}

		if match {
			nextNode, _ := condition["next_node"].(string)
			return c.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: process.ID,
				NodeKey:   nextNode,
			})
		}
	}

	// Use default path
	if defaultPath, ok := node.Properties["default_path"].(string); ok {
		return c.services.Node.Get(ctx, &structs.FindNodeParams{
			ProcessID: process.ID,
			NodeKey:   defaultPath,
		})
	}

	return nil, fmt.Errorf("no matching condition and no default path")
}

// processMetrics processes metrics
func (c *Core) processMetrics() {
	defer c.wg.Done()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectMetrics()
		}
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

func (c *Core) collectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.metrics.SetGauge("core_memory_usage", float64(memStats.Alloc))
	c.metrics.SetGauge("core_cpu_usage", getCPUUsage())
	c.metrics.SetGauge("core_goroutines", float64(runtime.NumGoroutine()))
}

// checkState checks the state of the core
func (c *Core) checkState() {
	// Check component health
	if !c.Executors.IsHealthy() {
		c.logger.Warn(c.ctx, "unhealthy executors detected")
	}
	if !c.Handlers.IsHealthy() {
		c.logger.Warn(c.ctx, "unhealthy handlers detected")
	}

	// Clean up stale processes
	c.cleanupStaleProcesses()
}

// ErrorTracker Implementation for enhanced error tracking
type ErrorTracker struct {
	errors     map[string][]error
	maxHistory int
	mu         sync.RWMutex
}

// NewErrorTracker creates a new ErrorTracker
func NewErrorTracker(maxHistory int) *ErrorTracker {
	return &ErrorTracker{
		errors:     make(map[string][]error),
		maxHistory: maxHistory,
	}
}

// TrackError tracks an error
func (et *ErrorTracker) TrackError(category string, err error) {
	if err == nil {
		return
	}

	et.mu.Lock()
	defer et.mu.Unlock()

	history := et.errors[category]
	if len(history) >= et.maxHistory {
		// Remove oldest error
		history = history[1:]
	}
	history = append(history, err)
	et.errors[category] = history
}

// GetErrors returns the errors for a given category
func (et *ErrorTracker) GetErrors(category string) []error {
	et.mu.RLock()
	defer et.mu.RUnlock()

	if history, ok := et.errors[category]; ok {
		result := make([]error, len(history))
		copy(result, history)
		return result
	}
	return nil
}

// Clear clears the errors for a given category
func (et *ErrorTracker) Clear(category string) {
	et.mu.Lock()
	defer et.mu.Unlock()
	delete(et.errors, category)
}

// GetMetrics returns the core metrics
func (c *Core) GetMetrics() map[string]any {
	if !c.cfg.Components.Metrics.Enabled {
		return nil
	}

	m := make(map[string]any)

	// Core metrics
	m["executions"] = map[string]any{
		"total":   c.metrics.GetCounter("core_executions_total"),
		"active":  c.metrics.GetGauge("core_executions_active"),
		"success": c.metrics.GetCounter("core_executions_success"),
		"failed":  c.metrics.GetCounter("core_executions_failed"),
		"timeout": c.metrics.GetCounter("core_executions_timeout"),
		"latency": c.metrics.GetHistogram("core_execution_time"),
	}

	// Resource metrics
	m["resources"] = map[string]any{
		"memory":     c.metrics.GetGauge("core_memory_usage"),
		"cpu":        c.metrics.GetGauge("core_cpu_usage"),
		"goroutines": c.metrics.GetGauge("core_goroutines"),
	}

	return m
}

// countRunningProcesses counts the number of running processes
func (c *Core) countRunningProcesses() int {
	count := 0
	c.running.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// cleanupStaleProcesses cleans up stale processes
func (c *Core) cleanupStaleProcesses() {
	c.running.Range(func(key, value any) bool {
		process := value.(*structs.ReadProcess)
		if process.StartTime != nil && time.Since(time.UnixMilli(*process.StartTime)) > time.Duration(c.cfg.Engine.DefaultTimeout) {
			// Process timed out, clean up
			c.running.Delete(key)
		}
		return true
	})
}
