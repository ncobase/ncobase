package core

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	wec "ncobase/core/workflow/engine/context"
	"ncobase/core/workflow/engine/executor"
	wet "ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"runtime"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/convert"
)

// Engine implements the workflow engine Interface
type Engine struct {
	// Core engine components
	core        *Core
	execManager *executor.Manager

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	config   *config.Config

	// Engine state
	startTime time.Time

	logger logger.Logger
}

// NewEngine creates a new workflow engine
func NewEngine(cfg *config.Config, svc *service.Service, em ext.ManagerInterface) (*Engine, error) {
	// Initialize core engine
	c, err := NewCore(cfg, svc, em)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine core: %w", err)
	}

	// Create executor manager
	exm, err := executor.NewManager(cfg, svc, em)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor manager: %w", err)
	}

	engine := &Engine{
		core:        c,
		execManager: exm,
		services:    svc,
		em:          em,
		config:      cfg,
	}

	return engine, nil
}

// Start starts the workflow engine
func (e *Engine) Start(_ context.Context) error {
	if err := e.core.Start(); err != nil {
		return fmt.Errorf("failed to start engine core: %w", err)
	}

	if err := e.execManager.Start(); err != nil {
		return fmt.Errorf("failed to start executor manager: %w", err)
	}

	e.startTime = time.Now()
	return nil
}

// Stop stops the workflow engine
func (e *Engine) Stop(_ context.Context) error {
	if err := e.execManager.Stop(); err != nil {
		return fmt.Errorf("failed to stop executor manager: %w", err)
	}

	if err := e.core.Stop(); err != nil {
		return fmt.Errorf("failed to stop engine core: %w", err)
	}

	return nil
}

// Status returns the engine status
func (e *Engine) Status() wet.EngineStatus {
	return e.core.Status
}

// StartProcess starts a new workflow process
func (e *Engine) StartProcess(ctx context.Context, req *structs.StartProcessRequest) (*structs.StartProcessResponse, error) {
	// Create workflow context
	wfCtx := wec.NewContext(ctx)
	wfCtx.WithOperator(req.Initiator)

	// Get process template
	template, err := e.services.Template.Get(ctx, &structs.FindTemplateParams{
		Code: req.TemplateID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create process instance
	process, err := e.services.Process.Create(ctx, &structs.ProcessBody{
		TemplateID:  template.ID,
		BusinessKey: req.BusinessKey,
		ModuleCode:  req.ModuleCode,
		FormCode:    req.FormCode,
		Initiator:   req.Initiator,
		Variables:   req.Variables,
		Status:      string(structs.StatusActive),
		Priority:    req.Priority,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create process: %w", err)
	}

	wfCtx.WithProcess(process)

	// Start process execution
	if err := e.core.Execute(wfCtx.Context(), process.ID); err != nil {
		return nil, fmt.Errorf("failed to execute process: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventProcessStarted, &structs.EventData{
		ProcessID:    process.ID,
		ProcessName:  process.ProcessCode,
		TemplateID:   template.ID,
		TemplateName: template.Name,
		ModuleCode:   req.ModuleCode,
		FormCode:     req.FormCode,
		Operator:     req.Initiator,
		Variables:    req.Variables,
	})

	return &structs.StartProcessResponse{
		ProcessID: process.ID,
		Status:    structs.Status(process.Status),
		StartTime: convert.ToPointer(time.Now().UnixMilli()),
		Variables: process.Variables,
	}, nil
}

// CompleteProcess completes a process
func (e *Engine) CompleteProcess(ctx context.Context, processID string) error {
	if err := e.core.CompleteProcess(ctx, processID); err != nil {
		return fmt.Errorf("engine complete process failed: %w", err)
	}
	return nil
}

// TerminateProcess terminates a process
func (e *Engine) TerminateProcess(ctx context.Context, req *structs.TerminateProcessRequest) error {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: req.ProcessID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// TODO: Stop all active nodes
	// for _, nodeKey := range process.ActiveNodes {
	// 	if err := e.execManager.CancelNode(ctx, nodeKey); err != nil {
	// 		e.logger.Errorf(ctx, "failed to cancel node %s: %v", nodeKey, err)
	// 	}
	// }

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Status:  string(structs.StatusTerminated),
			EndTime: func() *int64 { t := time.Now().UnixMilli(); return &t }(),
		},
	})

	// Publish event
	e.em.PublishEvent(structs.EventProcessTerminated, &structs.EventData{
		ProcessID: process.ID,
		Operator:  req.Operator,
		Comment:   req.Comment,
		Details: map[string]any{
			"reason": req.Reason,
		},
	})

	return err
}

// SuspendProcess suspends a running process
func (e *Engine) SuspendProcess(ctx context.Context, processID string, reason string) error {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			IsSuspended:   true,
			SuspendReason: reason,
			Status:        string(structs.StatusSuspended),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to suspend process: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventProcessSuspended, &structs.EventData{
		ProcessID: process.ID,
		Comment:   reason,
	})

	return nil
}

// ResumeProcess resumes a suspended process
func (e *Engine) ResumeProcess(ctx context.Context, processID string) error {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	if process.Status != string(structs.StatusSuspended) {
		return fmt.Errorf("process is not suspended")
	}

	// Update process status
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			IsSuspended:   false,
			SuspendReason: "",
			Status:        string(structs.StatusActive),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to resume process: %w", err)
	}

	// Resume execution from current node
	if err := e.core.Execute(ctx, process.ID); err != nil {
		return fmt.Errorf("failed to resume execution: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventProcessResumed, &structs.EventData{
		ProcessID: process.ID,
	})

	return nil
}

// GetProcessStatus gets detailed process status
func (e *Engine) GetProcessStatus(ctx context.Context, processID string) (*wet.ProcessStatus, error) {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	// Get active nodes
	activeNodes, err := e.GetActiveNodes(ctx, processID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active nodes: %w", err)
	}

	// Convert []*structs.ReadNode to []any
	activeNodesSlice := make([]any, len(activeNodes))
	for i, node := range activeNodes {
		activeNodesSlice[i] = node
	}

	return &wet.ProcessStatus{
		ProcessID:   process.ID,
		Status:      wet.ExecutionStatus(process.Status),
		CurrentNode: process.CurrentNode,
		ActiveNodes: activeNodesSlice,
		Variables:   process.Variables,
		StartTime:   process.CreatedAt,
		EndTime:     process.UpdatedAt,
	}, nil
}

// GetActiveNodes gets currently active nodes
func (e *Engine) GetActiveNodes(ctx context.Context, processID string) ([]*structs.ReadNode, error) {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	var activeNodes []*structs.ReadNode
	for _, nodeKey := range process.ActiveNodes {
		node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
			ProcessID: processID,
			NodeKey:   nodeKey,
		})
		if err != nil {
			continue
		}
		activeNodes = append(activeNodes, node)
	}

	return activeNodes, nil
}

// GetNextNodes gets possible next nodes from current position
func (e *Engine) GetNextNodes(ctx context.Context, processID string, nodeKey string) ([]*structs.ReadNode, error) {
	// Get current node
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	var nextNodes []*structs.ReadNode

	switch node.Type {
	case string(structs.NodeExclusive):
		// Get possible paths from gateway
		gwc := node.Properties["gatewayConfig"].(map[string]any)
		paths, _ := gwc["paths"].([]any)

		for _, p := range paths {
			path, _ := p.(map[string]any)
			nextNodeKey, _ := path["target"].(string)

			nextNode, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: processID,
				NodeKey:   nextNodeKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, nextNode)
		}

	case string(structs.NodeParallel):
		// Get all parallel branches
		for _, nodeKey := range node.ParallelNodes {
			next, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: processID,
				NodeKey:   nodeKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, next)
		}

	default:
		// Get direct next nodes
		for _, nextKey := range node.NextNodes {
			next, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
				ProcessID: processID,
				NodeKey:   nextKey,
			})
			if err != nil {
				continue
			}
			nextNodes = append(nextNodes, next)
		}
	}

	return nextNodes, nil
}

// Task related methods

func (e *Engine) CompleteTask(ctx context.Context, req *structs.CompleteTaskRequest) (*structs.CompleteTaskResponse, error) {
	// Get task
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check task status
	if task.Status != string(structs.StatusPending) {
		return nil, fmt.Errorf("task is not in pending status")
	}

	// Complete task
	endTime := time.Now().UnixMilli()
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:    string(structs.StatusCompleted),
			Action:    string(req.Action),
			Comment:   req.Comment,
			Variables: task.Variables,
			EndTime:   &endTime,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	// Complete associated node
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := e.core.CompleteNode(ctx, node); err != nil {
		return nil, fmt.Errorf("failed to complete node: %w", err)
	}

	// Get next nodes
	nextNodes, err := e.GetNextNodes(ctx, task.ProcessID, task.NodeKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get next nodes: %w", err)
	}

	nextNodeKeys := make([]string, len(nextNodes))
	for i, node := range nextNodes {
		nextNodeKeys[i] = node.NodeKey
	}

	return &structs.CompleteTaskResponse{
		TaskID:    task.ID,
		ProcessID: task.ProcessID,
		Action:    req.Action,
		EndTime:   &endTime,
		NextNodes: nextNodeKeys,
	}, nil
}

// DelegateTask delegates a task
func (e *Engine) DelegateTask(ctx context.Context, req *structs.DelegateTaskRequest) error {
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check task status
	if task.Status != string(structs.StatusPending) {
		return fmt.Errorf("task cannot be delegated in current status")
	}

	// Update task
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Assignees: []string{req.Delegator},
			Comment:   req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventTaskDelegated, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  structs.NodeType(task.NodeType),
		Operator:  req.Delegator,
		Action:    structs.ActionDelegate,
		Details: map[string]any{
			"delegate_to": req.Delegate,
			"reason":      req.Reason,
		},
	})

	return nil
}

// TransferTask transfers a task
func (e *Engine) TransferTask(ctx context.Context, req *structs.TransferTaskRequest) error {
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check task status
	if task.Status != string(structs.StatusPending) {
		return fmt.Errorf("task cannot be transferred in current status")
	}

	// Update task
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Assignees: []string{req.Transferor},
			Comment:   req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventTaskTransferred, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  structs.NodeType(task.NodeType),
		Operator:  req.Transferor,
		Action:    structs.ActionTransfer,
		Details: map[string]any{
			"transfer_to": req.Transferee,
			"reason":      req.Reason,
		},
	})

	return nil
}

// WithdrawTask withdraws a task
func (e *Engine) WithdrawTask(ctx context.Context, req *structs.WithdrawTaskRequest) error {
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status != string(structs.StatusPending) {
		return fmt.Errorf("task cannot be withdrawn in current status")
	}

	// Update task status
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:  string(structs.StatusWithdrawn),
			Comment: req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Reset node status
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: task.ProcessID,
		NodeKey:   task.NodeKey,
	})
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if err := e.RollbackNode(ctx, node.ProcessID, node.NodeKey); err != nil {
		return fmt.Errorf("failed to rollback node: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventTaskWithdrawn, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  structs.NodeType(task.NodeType),
		Operator:  req.Operator,
		Comment:   req.Comment,
	})

	return nil
}

// UrgeTask urges a task
func (e *Engine) UrgeTask(ctx context.Context, req *structs.UrgeTaskRequest) error {
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ProcessID: req.TaskID,
	})
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update task
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			IsUrged:   true,
			UrgeCount: task.UrgeCount + 1,
			Comment:   req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// Publish event
	e.em.PublishEvent(structs.EventTaskUrged, &structs.EventData{
		ProcessID: task.ProcessID,
		TaskID:    task.ID,
		NodeID:    task.NodeKey,
		NodeType:  structs.NodeType(task.NodeType),
		Operator:  req.Operator,
		Action:    structs.ActionUrge,
		Comment:   req.Comment,
		Variables: req.Variables,
	})

	return nil
}

// Variable operations

func (e *Engine) SetVariable(ctx context.Context, processID string, key string, value any) error {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	variables := process.Variables
	if variables == nil {
		variables = make(map[string]any)
	}
	variables[key] = value

	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			Variables: variables,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to update process variables: %w", err)
	}

	// Publish event
	e.em.PublishEvent("workflow.variables.updated", &structs.EventData{
		ProcessID: process.ID,
		Variables: variables,
		Details: map[string]any{
			"key":   key,
			"value": value,
		},
	})

	return nil
}

func (e *Engine) GetVariable(ctx context.Context, processID string, key string) (any, error) {
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	if process.Variables == nil {
		return nil, nil
	}

	value, exists := process.Variables[key]
	if !exists {
		return nil, fmt.Errorf("variable %s not found", key)
	}

	return value, nil
}

// Node operations

// RollbackNode rolls back a node
func (e *Engine) RollbackNode(ctx context.Context, processID string, nodeKey string) error {
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	// Get node handler
	handler, err := e.core.Handlers.GetHandler(wet.HandlerType(node.Type))
	if err != nil {
		return fmt.Errorf("get handler failed: %w", err)
	}

	// execute rollback
	if err := handler.Rollback(ctx, node); err != nil {
		return fmt.Errorf("rollback node failed: %w", err)
	}

	// Update node status
	_, err = e.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusError),
			Properties: map[string]any{
				"rollback_time": time.Now(),
			},
		},
	})

	return err
}

// JumpToNode jumps to a node
func (e *Engine) JumpToNode(ctx context.Context, processID string, targetNodeKey string, operator string, reason string) error {
	// Get target node
	targetNode, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   targetNodeKey,
	})
	if err != nil {
		return fmt.Errorf("failed to get target node: %w", err)
	}

	// Get process instance
	process, err := e.services.Process.Get(ctx, &structs.FindProcessParams{
		ProcessKey: processID,
	})
	if err != nil {
		return fmt.Errorf("failed to get process: %w", err)
	}

	// TODO: Cancel active nodes
	// for _, nodeKey := range process.ActiveNodes {
	// 	if err := e.core.Executors.CancelNode(ctx, nodeKey); err != nil {
	// 		e.logger.Warn(ctx, "failed to cancel active node", "nodeKey", nodeKey)
	// 	}
	// }

	// Update process
	_, err = e.services.Process.Update(ctx, &structs.UpdateProcessBody{
		ID: process.ID,
		ProcessBody: structs.ProcessBody{
			CurrentNode: targetNodeKey,
			ActiveNodes: []string{targetNodeKey},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update process: %w", err)
	}

	// Execute target node
	if err := e.core.ExecuteNode(ctx, process, targetNode); err != nil {
		return fmt.Errorf("failed to execute target node: %w", err)
	}

	// Publish event
	e.em.PublishEvent("workflow.node.jumped", &structs.EventData{
		ProcessID: processID,
		NodeID:    targetNodeKey,
		NodeType:  structs.NodeType(targetNode.Type),
		Operator:  operator,
		Comment:   reason,
	})

	return nil
}

// GetNodeInfo returns node info
func (e *Engine) GetNodeInfo(ctx context.Context, processID string, nodeKey string) (*structs.ReadNode, error) {
	node, err := e.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	return node, nil
}

// GetMetrics returns engine metrics
func (e *Engine) getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// GetMetrics returns engine metrics
func (e *Engine) GetMetrics() map[string]any {
	metrics := make(map[string]any)

	// engine metrics
	metrics["engine"] = map[string]any{
		"status":     e.Status(),
		"uptime":     time.Since(e.startTime).Seconds(),
		"start_time": e.startTime.Unix(),
	}

	// component metrics
	metrics["core"] = e.core.Metrics()
	metrics["executors"] = e.execManager.GetMetrics()

	// Add Resource metrics
	metrics["resource"] = map[string]any{
		"memory_usage":    e.getMemoryUsage(),
		"goroutine_count": runtime.NumGoroutine(),
		"num_cpu":         runtime.NumCPU(),
	}

	return metrics
}
