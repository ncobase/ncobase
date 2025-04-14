package executor

import (
	"context"
	"fmt"
	"math"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/scheduler"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"sort"
	"sync"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/expression"
)

// TaskExecutor handles task execution
type TaskExecutor struct {
	*BaseExecutor

	// Dependencies
	services *service.Service
	em       ext.ManagerInterface
	logger   logger.Logger

	// Runtime components
	expression *expression.Expression
	metrics    *metrics.Collector
	retries    *RetryExecutor
	scheduler  *scheduler.Scheduler

	// Task tracking
	tasks       sync.Map // taskID -> *TaskInfo
	activeTasks sync.Map
	userTasks   sync.Map // userID -> []taskID

	// Task assignment
	assignmentRules []AssignmentRule
	assignmentCache *sync.Map

	// Configuration
	config *config.TaskExecutorConfig

	// Internal state
	status types.ExecutionStatus
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TaskInfo tracks task execution state
type TaskInfo struct {
	ID          string
	Type        string
	Status      types.ExecutionStatus
	Assignee    string
	Delegated   bool
	Transferred bool
	StartTime   time.Time
	EndTime     *time.Time
	DueTime     *time.Time
	Priority    int
	RetryCount  int
	Error       error
	Variables   map[string]any
	mu          sync.RWMutex
}

// AssignmentRule represents task assignment rule
type AssignmentRule struct {
	Name       string         `json:"name"`
	Priority   int            `json:"priority"`
	Expression string         `json:"expression"`
	Assignees  []string       `json:"assignees"`
	Mode       string         `json:"mode"`
	Percentage int            `json:"percentage"`
	Conditions map[string]any `json:"conditions"`
	Enabled    bool           `json:"enabled"`
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(svc *service.Service, em ext.ManagerInterface, cfg *config.Config) (*TaskExecutor, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &TaskExecutor{
		BaseExecutor: NewBaseExecutor(types.TaskExecutor, "Task Executor", cfg),
		services:     svc,
		em:           em,
		config:       cfg.Executors.Task,
		ctx:          ctx,
		cancel:       cancel,
		status:       types.ExecutionPending,
	}

	// Initialize components
	var err error

	// Initialize metrics collector
	if executor.config.EnableMetrics {
		executor.metrics, err = metrics.NewCollector(cfg.Components.Metrics)
		if err != nil {
			return nil, fmt.Errorf("create metrics collector failed: %w", err)
		}
	}

	// Initialize expression
	executor.expression = expression.NewExpression(cfg.Components.Expression)

	// Initialize retry executor
	retryExecutor, err := NewRetryExecutor(svc, em, cfg)
	if err != nil {
		return nil, fmt.Errorf("create retry executor failed: %w", err)
	}
	executor.retries = retryExecutor

	// Initialize scheduler
	executor.scheduler = scheduler.NewScheduler(svc, em, cfg.Components.Scheduler)

	return executor, nil
}

// Lifecycle methods

// Start starts the task executor
func (e *TaskExecutor) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != types.ExecutionPending {
		return types.NewError(types.ErrInvalidParam, "executor not in pending state", nil)
	}

	// Start retry executor
	if err := e.retries.Start(); err != nil {
		return fmt.Errorf("start retry executor failed: %w", err)
	}

	if err := e.scheduler.Start(); err != nil {
		return fmt.Errorf("start scheduler failed: %w", err)
	}

	// Start background workers
	e.wg.Add(3)
	go e.processTimeouts()
	go e.processReminders()
	go e.processAutoAssignment()

	e.status = types.ExecutionActive
	return nil
}

// Stop stops the task executor
func (e *TaskExecutor) Stop() error {
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
	// Stop scheduler
	e.scheduler.Stop()

	// Wait for workers
	e.wg.Wait()

	e.status = types.ExecutionStopped
	return nil
}

// Core task operations

// CreateTask creates a new task
func (e *TaskExecutor) CreateTask(ctx context.Context, task *structs.TaskBody) (*structs.ReadTask, error) {
	// Run pre-create hook
	if e.config.Hooks.BeforeCreate != nil {
		if err := e.config.Hooks.BeforeCreate(ctx, task); err != nil {
			return nil, err
		}
	}

	// Create task
	createdTask, err := e.services.Task.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	// Create task info
	info := &TaskInfo{
		ID:        createdTask.ID,
		Type:      createdTask.NodeType,
		Status:    types.ExecutionActive,
		StartTime: time.Now(),
		Variables: createdTask.Variables,
	}

	if createdTask.DueTime != nil {
		dueTime := time.Unix(*createdTask.DueTime, 0)
		info.DueTime = &dueTime
	}

	// Track task
	e.tasks.Store(createdTask.ID, info)
	e.activeTasks.Store(createdTask.ID, info)

	// Auto assign if enabled
	if e.config.EnableAutoAssign {
		if err := e.AutoAssignTask(ctx, createdTask); err != nil {
			e.logger.Warn(ctx, "auto assign task failed", "error", err)
		}
	}

	// Run post-create hook
	if e.config.Hooks.AfterCreate != nil {
		if err := e.config.Hooks.AfterCreate(ctx, createdTask); err != nil {
			e.logger.Error(ctx, "post-create hook failed", err)
		}
	}

	// Schedule timeout check
	if info.DueTime != nil {
		err := e.scheduler.ScheduleTimeout(createdTask.ProcessID, createdTask.ID, *info.DueTime)
		if err != nil {
			e.logger.Error(ctx, "schedule timeout task failed", err)
		}
	}

	return createdTask, nil
}

// CompleteTask completes a task
func (e *TaskExecutor) CompleteTask(ctx context.Context, req *structs.CompleteTaskRequest) (*structs.CompleteTaskResponse, error) {
	// Run pre-complete hook
	if e.config.Hooks.BeforeComplete != nil {
		if err := e.config.Hooks.BeforeComplete(ctx, req); err != nil {
			return nil, err
		}
	}

	// Complete task
	resp, err := e.services.Task.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("complete task failed: %w", err)
	}

	// Update task info
	if info, ok := e.tasks.Load(req.TaskID); ok {
		taskInfo := info.(*TaskInfo)
		taskInfo.mu.Lock()
		taskInfo.Status = types.ExecutionCompleted
		now := time.Now()
		taskInfo.EndTime = &now
		taskInfo.mu.Unlock()
	}

	// Clean up task tracking
	e.activeTasks.Delete(req.TaskID)

	// Run post-complete hook
	if e.config.Hooks.AfterComplete != nil {
		if err := e.config.Hooks.AfterComplete(ctx, resp); err != nil {
			e.logger.Error(ctx, "post-complete hook failed", err)
		}
	}

	// Publish event
	e.em.PublishEvent(string(types.EventTaskCompleted), &types.Event{
		Type:   types.EventTaskCompleted,
		TaskID: req.TaskID,
		Details: map[string]any{
			"action":  req.Action,
			"comment": req.Comment,
		},
		Timestamp: time.Now(),
	})

	return resp, nil
}

// CancelTask cancels a task
func (e *TaskExecutor) CancelTask(ctx context.Context, taskID string) error {
	info, ok := e.GetTaskInfo(taskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}

	// Update status
	info.Status = types.ExecutionCancelled
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: taskID,
		TaskBody: structs.TaskBody{
			Status: string(types.ExecutionCancelled),
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Clean up task tracking
	e.activeTasks.Delete(taskID)

	// Publish event
	e.em.PublishEvent(string(types.EventTaskCancelled), &types.Event{
		Type:      types.EventTaskCancelled,
		TaskID:    taskID,
		Timestamp: time.Now(),
	})

	return nil
}

// ClaimTask claims a task
func (e *TaskExecutor) ClaimTask(ctx context.Context, taskID string, assignee string) error {
	info, ok := e.GetTaskInfo(taskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not claimable", nil)
	}

	if info.Assignee != "" {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task already assigned", nil)
	}

	// Update assignee
	info.Assignee = assignee
	info.mu.Unlock()

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: taskID,
		TaskBody: structs.TaskBody{
			Assignees: []string{assignee},
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Track user task
	e.trackUserTask(assignee, taskID)

	// Publish event
	e.em.PublishEvent("task.claimed", map[string]any{
		"task_id":  taskID,
		"assignee": assignee,
		"time":     time.Now(),
	})

	return nil
}

// Task delegation and transfer operations

// DelegateTask delegates a task
func (e *TaskExecutor) DelegateTask(ctx context.Context, req *structs.DelegateTaskRequest) error {
	if !e.config.AllowDelegate {
		return types.NewError(types.ErrInvalidParam, "task delegation not allowed", nil)
	}

	info, ok := e.GetTaskInfo(req.TaskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}

	if info.Delegated {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task already delegated", nil)
	}

	// Update task info
	info.Delegated = true
	info.Assignee = req.Delegate
	info.mu.Unlock()

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: req.TaskID,
		TaskBody: structs.TaskBody{
			Assignees:     []string{req.Delegate},
			IsDelegated:   true,
			DelegatedFrom: map[string]any{"user": req.Delegator},
			Comment:       req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Update user task tracking
	e.removeUserTask(req.Delegator, req.TaskID)
	e.trackUserTask(req.Delegate, req.TaskID)

	// Publish event
	e.em.PublishEvent(string(types.EventTaskDelegated), &types.Event{
		Type:   types.EventTaskDelegated,
		TaskID: req.TaskID,
		Details: map[string]any{
			"delegator": req.Delegator,
			"delegate":  req.Delegate,
			"reason":    req.Reason,
			"comment":   req.Comment,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// TransferTask transfers a task
func (e *TaskExecutor) TransferTask(ctx context.Context, req *structs.TransferTaskRequest) error {
	if !e.config.AllowTransfer {
		return types.NewError(types.ErrInvalidParam, "task transfer not allowed", nil)
	}

	info, ok := e.GetTaskInfo(req.TaskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}

	if info.Transferred {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task already transferred", nil)
	}

	// Update task info
	info.Transferred = true
	info.Assignee = req.Transferee
	info.mu.Unlock()

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: req.TaskID,
		TaskBody: structs.TaskBody{
			Assignees:     []string{req.Transferee},
			IsTransferred: true,
			Comment:       req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Update user task tracking
	e.removeUserTask(req.Transferor, req.TaskID)
	e.trackUserTask(req.Transferee, req.TaskID)

	// Publish event
	e.em.PublishEvent(string(types.EventTaskTransferred), &types.Event{
		Type:   types.EventTaskTransferred,
		TaskID: req.TaskID,
		Details: map[string]any{
			"transferor": req.Transferor,
			"transferee": req.Transferee,
			"reason":     req.Reason,
			"comment":    req.Comment,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// WithdrawTask withdraws a task
func (e *TaskExecutor) WithdrawTask(ctx context.Context, req *structs.WithdrawTaskRequest) error {
	if !e.config.AllowWithdraw {
		return types.NewError(types.ErrInvalidParam, "task withdrawal not allowed", nil)
	}

	info, ok := e.GetTaskInfo(req.TaskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}

	// Update task info
	info.Status = types.ExecutionWithdrawn
	now := time.Now()
	info.EndTime = &now
	info.mu.Unlock()

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: req.TaskID,
		TaskBody: structs.TaskBody{
			Status:  string(types.ExecutionWithdrawn),
			Comment: req.Comment,
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Clean up task tracking
	e.activeTasks.Delete(req.TaskID)

	// Publish event
	e.em.PublishEvent(string(types.EventTaskWithdrawn), &types.Event{
		Type:   types.EventTaskWithdrawn,
		TaskID: req.TaskID,
		Details: map[string]any{
			"operator": req.Operator,
			"reason":   req.Reason,
			"comment":  req.Comment,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// Task reminder operations

// UrgeTask urges a task
func (e *TaskExecutor) UrgeTask(ctx context.Context, req *structs.UrgeTaskRequest) error {
	if !e.config.AllowUrge {
		return types.NewError(types.ErrInvalidParam, "task urging not allowed", nil)
	}

	info, ok := e.GetTaskInfo(req.TaskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.RLock()
	if info.Status != types.ExecutionActive {
		info.mu.RUnlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}
	info.mu.RUnlock()

	// Update task in database
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ID: req.TaskID,
	})
	if err != nil {
		return err
	}

	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: req.TaskID,
		TaskBody: structs.TaskBody{
			IsUrged:   true,
			UrgeCount: task.UrgeCount + 1,
			Comment:   req.Comment,
			Variables: req.Variables,
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Publish event
	e.em.PublishEvent(string(types.EventTaskUrged), &types.Event{
		Type:   types.EventTaskUrged,
		TaskID: req.TaskID,
		Details: map[string]any{
			"operator":   req.Operator,
			"comment":    req.Comment,
			"variables":  req.Variables,
			"urge_count": task.UrgeCount + 1,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// RemindTask sends a reminder for a task
func (e *TaskExecutor) RemindTask(ctx context.Context, taskID string, message string) error {
	info, ok := e.GetTaskInfo(taskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.RLock()
	if info.Status != types.ExecutionActive {
		info.mu.RUnlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}
	info.mu.RUnlock()

	// Create reminder event
	e.em.PublishEvent("task.reminder", map[string]any{
		"task_id": taskID,
		"message": message,
		"time":    time.Now(),
	})

	// Update task reminder count
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ID: taskID,
	})
	if err != nil {
		return err
	}

	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: taskID,
		TaskBody: structs.TaskBody{
			IsUrged:   true,
			UrgeCount: task.UrgeCount + 1,
			Comment:   message,
		},
	})

	return err
}

// Task assignment operations

// AssignTask assigns a task
func (e *TaskExecutor) AssignTask(ctx context.Context, taskID string, assignees []string) error {
	info, ok := e.GetTaskInfo(taskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	if info.Status != types.ExecutionActive {
		info.mu.Unlock()
		return types.NewError(types.ErrInvalidParam, "task not active", nil)
	}

	// Get task
	task, err := e.services.Task.Get(ctx, &structs.FindTaskParams{
		ID: taskID,
	})
	if err != nil {
		info.mu.Unlock()
		return err
	}

	// Validate assignment
	if err := e.ValidateTaskAssignment(task, assignees); err != nil {
		info.mu.Unlock()
		return err
	}

	// Run pre-assign hook
	if e.config.Hooks.BeforeAssign != nil {
		if err := e.config.Hooks.BeforeAssign(ctx, task, assignees); err != nil {
			info.mu.Unlock()
			return err
		}
	}

	// Update task info
	oldAssignee := info.Assignee
	info.Assignee = assignees[0] // Primary assignee
	info.mu.Unlock()

	// Update task in database
	_, err = e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: taskID,
		TaskBody: structs.TaskBody{
			Assignees: assignees,
		},
	})
	if err != nil {
		return fmt.Errorf("update task failed: %w", err)
	}

	// Update user task tracking
	if oldAssignee != "" {
		e.removeUserTask(oldAssignee, taskID)
	}
	for _, assignee := range assignees {
		e.trackUserTask(assignee, taskID)
	}

	// Run post-assign hook
	if e.config.Hooks.AfterAssign != nil {
		e.config.Hooks.AfterAssign(ctx, task, nil)
	}

	// Publish event
	e.em.PublishEvent("task.assigned", map[string]any{
		"task_id":   taskID,
		"assignees": assignees,
		"time":      time.Now(),
	})

	return nil
}

// ReassignTask reassigns a task
func (e *TaskExecutor) ReassignTask(ctx context.Context, taskID string, assignees []string) error {
	if !e.config.AllowReassign {
		return types.NewError(types.ErrInvalidParam, "task reassignment not allowed", nil)
	}
	return e.AssignTask(ctx, taskID, assignees)
}

// AutoAssignTask automatically assigns a task
func (e *TaskExecutor) AutoAssignTask(ctx context.Context, task *structs.ReadTask) error {
	if !e.config.EnableAutoAssign {
		return nil
	}

	// Get assignment rules
	rules := e.GetAssignmentRules(task)
	if len(rules) == 0 {
		return nil
	}

	// Find matching rule and assignees
	var selectedAssignees []string
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Evaluate rule
		match, err := e.evaluateAssignmentRule(rule, task)
		if err != nil || !match {
			continue
		}

		// Get assignees based on rule mode
		assignees, err := e.getAssigneesFromRule(ctx, rule, task)
		if err != nil || len(assignees) == 0 {
			continue
		}

		selectedAssignees = assignees
		break
	}

	if len(selectedAssignees) == 0 {
		e.metrics.AddCounter("auto_assign_failure", 1)
		return types.NewError(types.ErrNotFound, "no suitable assignees found", nil)
	}

	// Assign task
	err := e.AssignTask(ctx, task.ID, selectedAssignees)
	if err != nil {
		e.metrics.AddCounter("auto_assign_failure", 1)
		return err
	}

	e.metrics.AddCounter("auto_assign_success", 1)
	return nil
}

// Task query operations

// GetTaskInfo gets task execution info
func (e *TaskExecutor) GetTaskInfo(taskID string) (*TaskInfo, bool) {
	info, ok := e.tasks.Load(taskID)
	if !ok {
		return nil, false
	}
	return info.(*TaskInfo), true
}

// GetActiveTasks gets all active tasks
func (e *TaskExecutor) GetActiveTasks() []*TaskInfo {
	var active []*TaskInfo
	e.activeTasks.Range(func(_, value any) bool {
		info := value.(*TaskInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive {
			active = append(active, info)
		}
		info.mu.RUnlock()
		return true
	})
	return active
}

// GetUserTasks gets tasks assigned to a user
func (e *TaskExecutor) GetUserTasks(userID string) ([]*structs.ReadTask, error) {
	var userTasks []*structs.ReadTask

	// Get task IDs for user
	taskIDs, ok := e.userTasks.Load(userID)
	if !ok {
		return userTasks, nil
	}

	// Get task details
	for _, id := range taskIDs.([]string) {
		task, err := e.services.Task.Get(e.ctx, &structs.FindTaskParams{
			ID: id,
		})
		if err != nil {
			continue
		}
		userTasks = append(userTasks, task)
	}

	return userTasks, nil
}

// GetTasksByNode gets tasks for a node
func (e *TaskExecutor) GetTasksByNode(nodeID string) ([]*structs.ReadTask, error) {
	// Query tasks from service
	result, err := e.services.Task.List(e.ctx, &structs.ListTaskParams{
		NodeKey: nodeID,
	})
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// Task rules & policies

// ValidateTaskAssignment validates task assignment
func (e *TaskExecutor) ValidateTaskAssignment(task *structs.ReadTask, assignees []string) error {
	// Check task status
	if task.Status != string(types.ExecutionPending) {
		return types.NewError(types.ErrInvalidParam, "task must be pending", nil)
	}

	// Check assignees
	if len(assignees) == 0 {
		return types.NewError(types.ErrInvalidParam, "assignees cannot be empty", nil)
	}

	// Apply assignment rules
	for _, rule := range e.GetAssignmentRules(task) {
		if !rule.Enabled {
			continue
		}

		// Evaluate rule conditions
		match, err := e.evaluateAssignmentRule(rule, task)
		if err != nil {
			continue
		}

		if match {
			// Validate against rule requirements
			if err := e.validateAgainstRule(rule, assignees); err != nil {
				return err
			}
		}
	}

	return nil
}

// CheckTaskTimeout checks if task has timed out
func (e *TaskExecutor) CheckTaskTimeout(task *structs.ReadTask) bool {
	if task.DueTime == nil {
		return false
	}
	return time.Now().After(time.Unix(*task.DueTime, 0))
}

// GetAssignmentRules gets assignment rules for a task
func (e *TaskExecutor) GetAssignmentRules(task *structs.ReadTask) []AssignmentRule {
	var rules []AssignmentRule

	// Get configured rules
	if ruleConfigs, ok := task.Variables["assignmentRules"].([]any); ok {
		for _, cfg := range ruleConfigs {
			if rule, ok := cfg.(map[string]any); ok {
				rules = append(rules, AssignmentRule{
					Name:       rule["name"].(string),
					Priority:   rule["priority"].(int),
					Expression: rule["expression"].(string),
					Mode:       rule["mode"].(string),
					Enabled:    rule["enabled"].(bool),
				})
			}
		}
	}

	// Add default rules if none configured
	if len(rules) == 0 {
		rules = append(rules, AssignmentRule{
			Name:     "default",
			Priority: 0,
			Mode:     "any",
			Enabled:  true,
		})
	}

	return rules
}

// evaluateAssignmentRule evaluates an assignment rule
func (e *TaskExecutor) evaluateAssignmentRule(rule AssignmentRule, task *structs.ReadTask) (bool, error) {
	if rule.Expression == "" {
		return true, nil
	}

	result, err := e.expression.Evaluate(e.ctx, rule.Expression, map[string]any{
		"task":      task,
		"variables": task.Variables,
	})

	return result.(bool), err
}

// validateAgainstRule validates assignees against rule requirements
func (e *TaskExecutor) validateAgainstRule(rule AssignmentRule, assignees []string) error {
	switch rule.Mode {
	case "all":
		// All required assignees must be present
		for _, required := range rule.Assignees {
			found := false
			for _, assignee := range assignees {
				if assignee == required {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("required assignee %s missing", required)
			}
		}

	case "any":
		// At least one assignee must match
		if len(rule.Assignees) > 0 {
			matched := false
			for _, assignee := range assignees {
				for _, allowed := range rule.Assignees {
					if assignee == allowed {
						matched = true
						break
					}
				}
			}
			if !matched {
				return fmt.Errorf("no matching assignee found")
			}
		}

	case "percentage":
		// Check if required percentage of assignees are present
		if rule.Percentage > 0 {
			matchCount := 0
			for _, assignee := range assignees {
				for _, allowed := range rule.Assignees {
					if assignee == allowed {
						matchCount++
						break
					}
				}
			}
			percentage := float64(matchCount) / float64(len(rule.Assignees)) * 100
			if percentage < float64(rule.Percentage) {
				return fmt.Errorf("insufficient matching assignees: got %.1f%%, need %d%%",
					percentage, rule.Percentage)
			}
		}
	}

	return nil
}

// getAssigneesFromRule gets assignees based on rule mode
func (e *TaskExecutor) getAssigneesFromRule(ctx context.Context, rule AssignmentRule, _ *structs.ReadTask) ([]string, error) {
	if len(rule.Assignees) == 0 {
		return nil, nil
	}

	switch rule.Mode {
	case "all":
		return rule.Assignees, nil

	case "any":
		if len(rule.Assignees) > 0 {
			// Select least loaded assignee
			return e.selectLeastLoadedAssignee(ctx, rule.Assignees)
		}

	case "percentage":
		if rule.Percentage > 0 {
			// Calculate required number of assignees
			required := int(math.Ceil(float64(len(rule.Assignees)) * float64(rule.Percentage) / 100))
			return e.selectTopNAssignees(ctx, rule.Assignees, required)
		}
	}

	return nil, nil
}

// User task tracking helpers

func (e *TaskExecutor) trackUserTask(userID string, taskID string) {
	value, _ := e.userTasks.LoadOrStore(userID, []string{})
	taskIDs := value.([]string)
	taskIDs = append(taskIDs, taskID)
	e.userTasks.Store(userID, taskIDs)
}

func (e *TaskExecutor) removeUserTask(userID string, taskID string) {
	value, ok := e.userTasks.Load(userID)
	if !ok {
		return
	}

	taskIDs := value.([]string)
	for i, id := range taskIDs {
		if id == taskID {
			taskIDs = append(taskIDs[:i], taskIDs[i+1:]...)
			break
		}
	}

	if len(taskIDs) == 0 {
		e.userTasks.Delete(userID)
	} else {
		e.userTasks.Store(userID, taskIDs)
	}
}

// Load balancing helpers

// selectLeastLoadedAssignee selects the least loaded assignee
func (e *TaskExecutor) selectLeastLoadedAssignee(_ context.Context, candidates []string) ([]string, error) {
	type assigneeLoad struct {
		userID string
		tasks  int
	}

	var loads []assigneeLoad
	for _, userID := range candidates {
		value, _ := e.userTasks.Load(userID)
		taskIDs, _ := value.([]string)
		loads = append(loads, assigneeLoad{
			userID: userID,
			tasks:  len(taskIDs),
		})
	}

	// Sort by load
	sort.Slice(loads, func(i, j int) bool {
		return loads[i].tasks < loads[j].tasks
	})

	if len(loads) > 0 {
		return []string{loads[0].userID}, nil
	}
	return nil, nil
}

// selectTopNAssignees selects the top N assignees
func (e *TaskExecutor) selectTopNAssignees(_ context.Context, candidates []string, n int) ([]string, error) {
	if n > len(candidates) {
		n = len(candidates)
	}

	// For now just return first N candidates
	// Could implement more sophisticated selection based on load/performance metrics
	return candidates[:n], nil
}

// Background workers

// processTimeouts checks for task timeouts
func (e *TaskExecutor) processTimeouts() {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.checkTimeouts()
		}
	}
}

// processReminders checks for task reminders
func (e *TaskExecutor) processReminders() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.config.ReminderInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.checkReminders()
		}
	}
}

// processUnassignedTasks handles unassigned tasks
func (e *TaskExecutor) processUnassignedTasks() {
	e.tasks.Range(func(_, value any) bool {
		info := value.(*TaskInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive && info.Assignee == "" {
			e.handleTaskTimeout(info)
		}
		info.mu.RUnlock()
		return true
	})
}

// processAutoAssignment automatically assigns tasks
func (e *TaskExecutor) processAutoAssignment() {
	defer e.wg.Done()

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.processUnassignedTasks()
		}
	}
}

// checkTimeouts checks for task timeouts
func (e *TaskExecutor) checkTimeouts() {
	e.tasks.Range(func(_, value any) bool {
		info := value.(*TaskInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive && info.DueTime != nil {
			if time.Now().After(*info.DueTime) {
				e.handleTaskTimeout(info)
			}
		}
		info.mu.RUnlock()
		return true
	})
}

// checkReminders checks for task reminders
func (e *TaskExecutor) checkReminders() {
	if !e.config.OverdueReminder {
		return
	}

	e.tasks.Range(func(_, value any) bool {
		info := value.(*TaskInfo)
		info.mu.RLock()
		if info.Status == types.ExecutionActive && info.DueTime != nil {
			if time.Now().After(*info.DueTime) {
				e.sendTaskReminder(info)
			}
		}
		info.mu.RUnlock()
		return true
	})
}

// handleTaskTimeout handles task timeouts
func (e *TaskExecutor) handleTaskTimeout(info *TaskInfo) {
	// Get task configuration
	task, err := e.services.Task.Get(e.ctx, &structs.FindTaskParams{
		ID: info.ID,
	})
	if err != nil {
		return
	}

	// Call timeout hook if configured
	if e.config.Hooks.OnTimeout != nil {
		e.config.Hooks.OnTimeout(e.ctx, task)
	}

	// Publish timeout event
	e.em.PublishEvent(string(types.EventTaskTimeout), &types.Event{
		Type:      types.EventTaskTimeout,
		TaskID:    info.ID,
		Timestamp: time.Now(),
	})
}

// sendTaskReminder sends a task reminder
func (e *TaskExecutor) sendTaskReminder(info *TaskInfo) {
	task, err := e.services.Task.Get(e.ctx, &structs.FindTaskParams{
		ID: info.ID,
	})
	if err != nil {
		return
	}

	// Call overdue hook if configured
	if e.config.Hooks.OnOverdue != nil {
		e.config.Hooks.OnOverdue(e.ctx, task)
	}

	// Send reminder notification
	err = e.RemindTask(e.ctx, info.ID, "Task is overdue")
	if err != nil {
		e.logger.Error(e.ctx, "failed to send task overdue notification", "task_id", info.ID, "error", err)
		return
	}
}

// Execute task execution logic
func (e *TaskExecutor) Execute(ctx context.Context, req *types.Request) (*types.Response, error) {
	switch req.Type {
	case types.TaskExecutor:
		task, ok := req.Context["task"].(*structs.TaskBody)
		if !ok {
			return nil, types.NewError(types.ErrInvalidParam, "invalid task in request", nil)
		}
		result, err := e.CreateTask(ctx, task)
		if err != nil {
			return nil, err
		}
		return &types.Response{
			ID:        req.ID,
			Status:    types.ExecutionCompleted,
			Data:      result,
			StartTime: time.Now(),
		}, nil
	default:
		return nil, types.NewError(types.ErrInvalidParam, "unsupported request type", nil)
	}
}

// Cancel cancel task execution
func (e *TaskExecutor) Cancel(ctx context.Context, id string) error {
	return e.CancelTask(ctx, id)
}

// Rollback rollback task execution
func (e *TaskExecutor) Rollback(ctx context.Context, id string) error {
	info, ok := e.GetTaskInfo(id)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	// Update task status
	info.Status = types.ExecutionRollbacked
	now := time.Now()
	info.EndTime = &now

	// Update task in database
	_, err := e.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: id,
		TaskBody: structs.TaskBody{
			Status: string(types.ExecutionRollbacked),
		},
	})

	return err
}

// initMetrics initialize task executor metrics
func (e *TaskExecutor) initMetrics() error {
	// Task execution metrics
	e.metrics.RegisterCounter("task_total")
	e.metrics.RegisterCounter("task_success")
	e.metrics.RegisterCounter("task_error")
	e.metrics.RegisterCounter("task_timeout")
	e.metrics.RegisterCounter("task_cancelled")

	// Task state metrics
	e.metrics.RegisterGauge("tasks_active")
	e.metrics.RegisterGauge("tasks_pending")
	e.metrics.RegisterGauge("tasks_completed")

	// Assignment metrics
	e.metrics.RegisterCounter("task_assigned")
	e.metrics.RegisterCounter("task_delegated")
	e.metrics.RegisterCounter("task_transferred")
	e.metrics.RegisterCounter("auto_assign_success")
	e.metrics.RegisterCounter("auto_assign_failure")

	// Performance metrics
	e.metrics.RegisterHistogram("task_execution_time", 1000)
	e.metrics.RegisterHistogram("task_waiting_time", 1000)
	e.metrics.RegisterHistogram("assignment_time", 1000)

	return nil
}

// GetMetrics get task executor metrics
func (e *TaskExecutor) GetMetrics() map[string]any {
	if !e.config.EnableMetrics {
		return nil
	}
	return map[string]any{
		"task_total":          e.metrics.GetCounter("task_total"),
		"task_success":        e.metrics.GetCounter("task_success"),
		"task_error":          e.metrics.GetCounter("task_error"),
		"task_timeout":        e.metrics.GetCounter("task_timeout"),
		"task_cancelled":      e.metrics.GetCounter("task_cancelled"),
		"tasks_active":        e.metrics.GetGauge("tasks_active"),
		"tasks_pending":       e.metrics.GetGauge("tasks_pending"),
		"tasks_completed":     e.metrics.GetGauge("tasks_completed"),
		"task_assigned":       e.metrics.GetCounter("task_assigned"),
		"task_delegated":      e.metrics.GetCounter("task_delegated"),
		"task_transferred":    e.metrics.GetCounter("task_transferred"),
		"auto_assign_success": e.metrics.GetCounter("auto_assign_success"),
		"auto_assign_failure": e.metrics.GetCounter("auto_assign_failure"),
		"task_execution_time": e.metrics.GetHistogram("task_execution_time"),
		"task_waiting_time":   e.metrics.GetHistogram("task_waiting_time"),
		"assignment_time":     e.metrics.GetHistogram("assignment_time"),
	}
}

// GetCapabilities returns the task executor capabilities
func (e *TaskExecutor) GetCapabilities() *types.ExecutionCapabilities {
	return &types.ExecutionCapabilities{
		SupportsAsync:    true,
		SupportsRetry:    true,
		SupportsRollback: false,
		MaxConcurrency:   int(e.config.MaxConcurrent),
		MaxBatchSize:     int(e.config.MaxBatchSize),
		AllowedActions: []string{
			"create",
			"complete",
			"delegate",
			"transfer",
			"withdraw",
			"urge",
			"cancel",
		},
	}
}

// Status get task executor status
func (e *TaskExecutor) Status() types.ExecutionStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// IsHealthy checks if executor is healthy
func (e *TaskExecutor) IsHealthy() bool {
	return e.Status() == types.ExecutionActive
}
