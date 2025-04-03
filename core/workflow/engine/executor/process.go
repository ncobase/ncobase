package executor

import (
	"context"
	"fmt"
	"ncobase/common/expression"
	"ncobase/common/extension"
	"ncobase/common/logger"
	"ncobase/common/uuid"
	"ncobase/core/workflow/engine/config"
	"sync"
	"time"

	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
)

// ProcessExecutor handles process execution
type ProcessExecutor struct {
	*BaseExecutor

	// Services & components
	services   *service.Service
	em         *extension.Manager
	expression *expression.Expression

	// Dependent executors
	nodeExecutor *NodeExecutor
	taskExecutor *TaskExecutor

	// Process state tracking
	processes sync.Map // processID -> *ProcessInfo

	// Configuration
	config *config.ProcessExecutorConfig

	logger logger.Logger
}

// ProcessInfo tracks process execution state
type ProcessInfo struct {
	ID          string
	Status      types.ExecutionStatus
	CurrentNode string
	StartTime   time.Time
	EndTime     *time.Time
	Variables   map[string]any
	Error       error
	RetryCount  int
	mu          sync.RWMutex
}

// NewProcessExecutor creates a new process executor
func NewProcessExecutor(svc *service.Service, em *extension.Manager, cfg *config.Config) (*ProcessExecutor, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	executor := &ProcessExecutor{
		BaseExecutor: NewBaseExecutor(types.ProcessExecutor, "Process Executor", cfg),
		expression:   expression.NewExpression(cfg.Components.Expression),
		services:     svc,
		em:           em,
		config:       cfg.Executors.Process,
	}

	return executor, nil
}

// Start starts the process executor
func (e *ProcessExecutor) Start() error {
	if err := e.BaseExecutor.Start(); err != nil {
		return err
	}

	// Initialize dependent executors if needed
	if e.nodeExecutor == nil || e.taskExecutor == nil {
		return types.NewError(types.ErrSystem, "dependent executors not initialized", nil)
	}

	return nil
}

// SetDependentExecutors sets dependent executors
func (e *ProcessExecutor) SetDependentExecutors(node *NodeExecutor, task *TaskExecutor) {
	e.nodeExecutor = node
	e.taskExecutor = task
}

// ConfigureProcess configures process specific settings
func (e *ProcessExecutor) ConfigureProcess(cfg *config.ProcessExecutorConfig) {
	e.config = cfg
}

// StartProcess starts a new process instance
func (e *ProcessExecutor) StartProcess(ctx context.Context, req *types.StartProcessRequest) (*types.StartProcessResponse, error) {
	// Validate request
	if req == nil || req.TemplateID == "" {
		return nil, types.NewError(types.ErrInvalidParam, "invalid request", nil)
	}

	// Get process template
	template, err := e.services.Template.Get(ctx, &structs.FindTemplateParams{
		Code: req.TemplateID,
	})
	if err != nil {
		return nil, fmt.Errorf("get template failed: %w", err)
	}

	// Create process instance
	process, err := e.services.Process.Create(ctx, &structs.ProcessBody{
		TemplateID:  template.ID,
		ProcessCode: uuid.NewString(), // Generate unique code
		BusinessKey: req.BusinessKey,
		Variables:   req.Variables,
		Status:      string(types.ExecutionActive),
	})
	if err != nil {
		return nil, fmt.Errorf("create process failed: %w", err)
	}

	// Create execution request
	execReq := &types.Request{
		ID:   process.ID,
		Type: types.ProcessExecutor,
		Context: map[string]any{
			"process": process,
		},
		Variables: req.Variables,
	}

	// Execute process
	resp, err := e.Execute(ctx, execReq)
	if err != nil {
		return nil, fmt.Errorf("execute process failed: %w", err)
	}

	st := resp.StartTime.UnixMilli()
	return &types.StartProcessResponse{
		ProcessID: process.ID,
		Status:    resp.Status,
		StartTime: &st,
		Variables: process.Variables,
	}, nil
}

// SuspendProcess suspends a running process
func (e *ProcessExecutor) SuspendProcess(ctx context.Context, processID string, reason string) error {
	// Get process
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("get process failed: %w", err)
	}

	// Validate process state
	if process.Status != string(types.ExecutionActive) {
		return types.NewError(types.ErrInvalidParam, "process not active", nil)
	}

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status:        string(types.ExecutionSuspended),
			IsSuspended:   true,
			SuspendReason: reason,
		},
	})
	if err != nil {
		return fmt.Errorf("update process failed: %w", err)
	}

	// Cancel active nodes
	for _, nodeID := range process.ActiveNodes {
		if err := e.nodeExecutor.Cancel(ctx, nodeID); err != nil {
			e.logger.Error(ctx, "cancel node failed", err)
		}
	}

	// Publish event
	e.em.PublishEvent(structs.EventProcessSuspended, &structs.EventData{
		ProcessID: processID,
		Comment:   reason,
	})

	return nil
}

// ResumeProcess resumes a suspended process
func (e *ProcessExecutor) ResumeProcess(ctx context.Context, processID string) error {
	// Get process
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("get process failed: %w", err)
	}

	// Validate process state
	if process.Status != string(types.ExecutionSuspended) {
		return types.NewError(types.ErrInvalidParam, "process not suspended", nil)
	}

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status:        string(types.ExecutionActive),
			IsSuspended:   false,
			SuspendReason: "",
		},
	})
	if err != nil {
		return fmt.Errorf("update process failed: %w", err)
	}

	// Resume from current node
	if process.CurrentNode != "" {
		node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
			ProcessID: process.ID,
			NodeKey:   process.CurrentNode,
		})
		if err != nil {
			return fmt.Errorf("get node failed: %w", err)
		}

		if err := e.nodeExecutor.ExecuteNode(ctx, node); err != nil {
			return fmt.Errorf("execute node failed: %w", err)
		}
	}

	// Publish event
	e.em.PublishEvent(structs.EventProcessResumed, &structs.EventData{
		ProcessID: processID,
	})

	return nil
}

// GetProcessVariables gets process variables
func (e *ProcessExecutor) GetProcessVariables(processID string) (map[string]any, error) {
	process, err := e.services.Process.Get(context.Background(), &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return nil, fmt.Errorf("get process failed: %w", err)
	}

	return process.Variables, nil
}

// doExecute process execution logic
func (e *ProcessExecutor) doExecute(ctx context.Context, req *types.Request) (*types.Response, error) {
	process, ok := req.Context["process"].(*structs.ReadProcess)
	if !ok {
		return nil, types.NewError(types.ErrInvalidParam, "invalid process in request", nil)
	}

	// Create process info
	info := &ProcessInfo{
		ID:        process.ID,
		Status:    types.ExecutionActive,
		StartTime: time.Now(),
		Variables: process.Variables,
	}
	e.processes.Store(process.ID, info)
	defer e.processes.Delete(process.ID)

	// Run pre-execute hook
	if e.config.Hooks.BeforeExecute != nil {
		if err := e.config.Hooks.BeforeExecute(ctx, process); err != nil {
			return nil, e.handleProcessError(ctx, process, err)
		}
	}

	// Find start node
	startNode, err := e.findStartNode(ctx, process)
	if err != nil {
		return nil, e.handleProcessError(ctx, process, err)
	}

	// Execute start node
	nodeReq := &types.Request{
		ID:   process.ID,
		Type: types.NodeExecutor,
		Context: map[string]any{
			"node": startNode,
		},
		Variables: process.Variables,
	}

	nodeResp, err := e.nodeExecutor.Execute(ctx, nodeReq)
	if err != nil {
		return nil, e.handleProcessError(ctx, process, err)
	}

	// Update process info
	info.mu.Lock()
	info.Status = types.ExecutionCompleted
	info.EndTime = nodeResp.EndTime
	info.mu.Unlock()

	// Run post-execute hook
	if e.config.Hooks.AfterExecute != nil {
		e.config.Hooks.AfterExecute(ctx, process, nil)
	}

	return &types.Response{
		ID:        req.ID,
		Status:    types.ExecutionCompleted,
		Data:      process,
		StartTime: info.StartTime,
		EndTime:   info.EndTime,
		Duration:  time.Since(info.StartTime),
	}, nil
}

// Cancel cancels process execution
func (e *ProcessExecutor) Cancel(ctx context.Context, processID string) error {
	// Get process info
	info, ok := e.processes.Load(processID)
	if !ok {
		return types.NewError(types.ErrNotFound, "process not found", nil)
	}
	pInfo := info.(*ProcessInfo)

	pInfo.mu.Lock()
	if pInfo.Status != types.ExecutionActive {
		pInfo.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "process not active", nil)
	}

	// Update status
	pInfo.Status = types.ExecutionCancelled
	now := time.Now()
	pInfo.EndTime = &now
	pInfo.mu.Unlock()

	// Cancel active nodes
	if pInfo.CurrentNode != "" {
		if err := e.nodeExecutor.Cancel(ctx, pInfo.CurrentNode); err != nil {
			return fmt.Errorf("cancel node failed: %w", err)
		}
	}

	// Update process in database
	_, err := e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: processID,
		ProcessBody: structs.ProcessBody{
			Status: string(types.ExecutionCancelled),
		},
	})

	if err != nil {
		return fmt.Errorf("update process failed: %w", err)
	}

	// Publish event
	e.em.PublishEvent(string(types.EventProcessTerminated), &types.Event{
		Type:      types.EventProcessTerminated,
		ProcessID: processID,
		Details: map[string]any{
			"reason": "cancelled",
		},
		Timestamp: time.Now(),
	})

	return nil
}

// Rollback rolls back process execution
func (e *ProcessExecutor) Rollback(ctx context.Context, processID string) error {
	// Get process
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("get process failed: %w", err)
	}

	// Get completed nodes
	nodes, err := e.services.Node.List(ctx, &structs.ListNodeParams{
		ProcessID: processID,
		Status:    string(types.ExecutionCompleted),
	})
	if err != nil {
		return fmt.Errorf("list nodes failed: %w", err)
	}

	// Rollback nodes in reverse order
	for i := len(nodes.Items) - 1; i >= 0; i-- {
		node := nodes.Items[i]
		if err := e.nodeExecutor.Rollback(ctx, node.ID); err != nil {
			return fmt.Errorf("rollback node failed: %w", err)
		}
	}

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status: string(types.ExecutionRollbacked),
		},
	})

	return err
}

// findStartNode finds start node
func (e *ProcessExecutor) findStartNode(ctx context.Context, process *structs.ReadProcess) (*structs.ReadNode, error) {
	nodes, err := e.services.Node.List(ctx, &structs.ListNodeParams{
		ProcessID: process.ID,
		Type:      string(types.NodeStart),
	})
	if err != nil {
		return nil, fmt.Errorf("list nodes failed: %w", err)
	}

	if len(nodes.Items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "start node not found", nil)
	}

	return nodes.Items[0], nil
}

// getNextNodes gets next nodes
func (c *ProcessExecutor) getNextNodes(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) ([]*structs.ReadNode, error) {
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
func (c *ProcessExecutor) evaluateConditions(ctx context.Context, process *structs.ReadProcess, node *structs.ReadNode) (*structs.ReadNode, error) {
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

// evaluateExpression evaluates a condition expression
func (c *ProcessExecutor) evaluateExpression(ctx context.Context, expr string, variables map[string]any) (bool, error) {
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

// handleProcessError handles process error
func (e *ProcessExecutor) handleProcessError(ctx context.Context, process *structs.ReadProcess, err error) error {
	// Run error hook
	if e.config.Hooks.OnError != nil {
		e.config.Hooks.OnError(ctx, process, err)
	}

	// Rollback if configured
	if e.config.RollbackOnError {
		if rbErr := e.Rollback(ctx, process.ID); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
		}
	}

	// Update process status
	_, updateErr := e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status: string(types.ExecutionError),
		},
	})
	if updateErr != nil {
		return fmt.Errorf("update process failed: %v (original error: %v)", updateErr, err)
	}

	// Publish error event
	e.em.PublishEvent(string(types.EventNodeFailed), &types.Event{
		Type:      types.EventNodeFailed,
		ProcessID: process.ID,
		Details: map[string]any{
			"error": err.Error(),
		},
		Timestamp: time.Now(),
	})

	return err
}

// Additional helper methods

// GetProcessInfo returns process execution info
func (e *ProcessExecutor) GetProcessInfo(processID string) (*ProcessInfo, bool) {
	info, ok := e.processes.Load(processID)
	if !ok {
		return nil, false
	}
	return info.(*ProcessInfo), true
}

// GetActiveProcesses returns all active processes
func (e *ProcessExecutor) GetActiveProcesses() []*ProcessInfo {
	var active []*ProcessInfo
	e.processes.Range(func(key, value any) bool {
		info := value.(*ProcessInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive {
			active = append(active, info)
		}
		info.mu.RUnlock()
		return true
	})
	return active
}

// validateProcess validates process state
func (e *ProcessExecutor) validateProcess(process *structs.ReadProcess) error {
	if process.Status == string(types.ExecutionCompleted) ||
		process.Status == string(types.ExecutionTerminated) ||
		process.Status == string(types.ExecutionRollbacked) {
		return types.NewError(types.ErrInvalidParam, "process already ended", nil)
	}

	if process.Status == string(types.ExecutionSuspended) {
		return types.NewError(types.ErrInvalidParam, "process is suspended", nil)
	}

	return nil
}

// handleSubprocess handles subprocess execution
func (e *ProcessExecutor) handleSubprocess(ctx context.Context, parentID string, subprocess *structs.ReadProcess) error {
	if !e.config.AutoStartSubProcess {
		return nil
	}

	// Create execution request
	req := &types.Request{
		ID:   subprocess.ID,
		Type: types.ProcessExecutor,
		Context: map[string]any{
			"process": subprocess,
		},
		Variables: subprocess.Variables,
	}

	// Execute subprocess
	subCtx := ctx
	if e.config.SubProcessTimeout > 0 {
		var cancel context.CancelFunc
		subCtx, cancel = context.WithTimeout(ctx, e.config.SubProcessTimeout)
		defer cancel()
	}

	_, err := e.Execute(subCtx, req)
	if err != nil {
		return fmt.Errorf("subprocess execution failed: %w", err)
	}

	// Wait for completion if configured
	if e.config.WaitSubProcess {
		if err := e.waitForSubprocess(subCtx, subprocess.ID); err != nil {
			return fmt.Errorf("wait for subprocess failed: %w", err)
		}
	}

	return nil
}

// waitForSubprocess waits for subprocess completion
func (e *ProcessExecutor) waitForSubprocess(ctx context.Context, processID string) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
				ProcessKey: processID,
			})
			if err != nil {
				return err
			}

			if process.Status != string(types.ExecutionActive) {
				return nil
			}
		}
	}
}

// initMetrics initialize process executor metrics
func (e *ProcessExecutor) initMetrics() error {
	// Process execution metrics
	e.metrics.RegisterCounter("process_total")
	e.metrics.RegisterCounter("process_success")
	e.metrics.RegisterCounter("process_error")
	e.metrics.RegisterCounter("process_timeout")
	e.metrics.RegisterCounter("process_cancelled")

	// Process state metrics
	e.metrics.RegisterGauge("processes_active")
	e.metrics.RegisterGauge("processes_suspended")
	e.metrics.RegisterGauge("processes_completed")

	// Subprocess metrics
	e.metrics.RegisterCounter("subprocess_started")
	e.metrics.RegisterCounter("subprocess_completed")
	e.metrics.RegisterCounter("subprocess_failed")

	// Performance metrics
	e.metrics.RegisterHistogram("process_execution_time", 1000)
	e.metrics.RegisterHistogram("node_completion_time", 1000)

	return nil
}

// GetMetrics returns executor metrics
func (e *ProcessExecutor) GetMetrics() map[string]any {
	if !e.config.EnableMetrics {
		return nil
	}

	return map[string]any{
		"process_total":          e.metrics.GetCounter("process_total"),
		"process_success":        e.metrics.GetCounter("process_success"),
		"process_error":          e.metrics.GetCounter("process_error"),
		"process_timeout":        e.metrics.GetCounter("process_timeout"),
		"process_cancelled":      e.metrics.GetCounter("process_cancelled"),
		"processes_active":       e.metrics.GetGauge("processes_active"),
		"processes_suspended":    e.metrics.GetGauge("processes_suspended"),
		"processes_completed":    e.metrics.GetGauge("processes_completed"),
		"subprocess_started":     e.metrics.GetCounter("subprocess_started"),
		"subprocess_completed":   e.metrics.GetCounter("subprocess_completed"),
		"subprocess_failed":      e.metrics.GetCounter("subprocess_failed"),
		"process_execution_time": e.metrics.GetHistogram("process_execution_time"),
		"node_completion_time":   e.metrics.GetHistogram("node_completion_time"),
	}
}

// GetCapabilities returns executor capabilities
func (e *ProcessExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    true,
		SupportsRetry:    true,
		SupportsRollback: true,
		MaxConcurrency:   int(e.config.MaxConcurrent),
		MaxBatchSize:     int(e.config.MaxBatchSize),
		AllowedActions: []string{
			"start",
			"complete",
			"suspend",
			"resume",
			"terminate",
			"cancel",
			"rollback",
		},
	}
}

// Status return this exector status
func (e *ProcessExecutor) Status() types.ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}
