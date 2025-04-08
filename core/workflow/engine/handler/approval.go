package handler

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/metrics"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"sort"
	"sync"
	"time"
)

// ApprovalHandler handles approval nodes in workflow
type ApprovalHandler struct {
	*BaseHandler

	// Approval records tracking
	approvals sync.Map // nodeID -> []ApprovalRecord

	// Approval strategies
	strategies map[string]ApprovalStrategy

	// Active tasks
	activeTasks sync.Map // taskID -> *ApprovalTask

	// Metrics collector
	metrics *metrics.Collector

	// Handler configuration
	config *config.ApprovalHandlerConfig
}

// ApprovalRecord represents an approval action record
type ApprovalRecord struct {
	TaskID    string
	Approver  string
	Action    string
	Comment   string
	Timestamp time.Time
	Variables map[string]any
}

// ApprovalTask represents an active approval task
type ApprovalTask struct {
	ID           string
	NodeID       string
	Approver     string
	StartTime    time.Time
	DueTime      *time.Time
	Status       string
	Variables    map[string]any
	UrgeCount    int
	LastUrgeTime *time.Time
	mu           sync.RWMutex
}

// ApprovalStrategy represents approval strategy interface
type ApprovalStrategy interface {
	IsCompleted(completed, approved, total int) bool
	IsApproved(approved, total int) bool
	GetName() string
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(svc *service.Service, em *extension.Manager, cfg *config.Config) (*ApprovalHandler, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	handler := &ApprovalHandler{
		BaseHandler: NewBaseHandler("approval", "Approval Handler", svc, em, cfg.Handlers.Base),
		strategies:  make(map[string]ApprovalStrategy),
		config:      cfg.Handlers.Approval,
	}

	// Register default approval strategies
	handler.registerDefaultStrategies()

	// Initialize metrics
	if cfg.Handlers.Base.EnableMetrics {
		collector, err := metrics.NewCollector(cfg.Components.Metrics)
		if err != nil {
			return nil, fmt.Errorf("create metrics collector failed: %w", err)
		}
		handler.metrics = collector
		handler.initMetrics()
	}

	return handler, nil
}

// Type returns handler type
func (h *ApprovalHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *ApprovalHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *ApprovalHandler) Priority() int { return h.priority }

// Start starts the approval handler
func (h *ApprovalHandler) Start() error {
	if err := h.BaseHandler.Start(); err != nil {
		return err
	}

	return nil
}

func (h *ApprovalHandler) Stop() error {
	if err := h.BaseHandler.Stop(); err != nil {
		return err
	}

	// Clean up approvals
	h.approvals = sync.Map{}
	h.activeTasks = sync.Map{}

	return nil
}

// Reset resets the approval handler
func (h *ApprovalHandler) Reset() error {
	if err := h.BaseHandler.Reset(); err != nil {
		return err
	}
	// Clean up approvals
	h.approvals = sync.Map{}
	h.activeTasks = sync.Map{}

	return nil
}

// Execute executes the approval node
func (h *ApprovalHandler) Execute(ctx context.Context, node *structs.ReadNode) error {
	startTime := time.Now()

	// Parse approval configuration
	c, err := h.parseApprovalConfig(node)
	if err != nil {
		return err
	}

	// Resolve approvers
	approvers, err := h.resolveApprovers(ctx, node, c)
	if err != nil {
		return err
	}

	if len(approvers) == 0 {
		return types.NewError(types.ErrValidation, "no approvers resolved", nil)
	}

	// Create approval tasks
	if err := h.createApprovalTasks(ctx, node, approvers, c); err != nil {
		h.metrics.AddCounter("approval_create_failed", 1)
		return err
	}

	h.metrics.AddCounter("approval_created", 1)
	h.metrics.RecordValue("approval_create_duration", time.Since(startTime).Seconds())

	return nil
}

// Complete completes the approval node
func (h *ApprovalHandler) Complete(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error {
	startTime := time.Now()

	// Record approval action
	h.recordApproval(node.ID, ApprovalRecord{
		TaskID:    req.TaskID,
		Approver:  req.Operator,
		Action:    string(req.Action),
		Comment:   req.Comment,
		Timestamp: time.Now(),
		Variables: req.Variables,
	})

	// Get approval configuration
	c, err := h.parseApprovalConfig(node)
	if err != nil {
		return err
	}

	// Check completion status
	complete, err := h.checkCompletion(ctx, node, c)
	if err != nil {
		return err
	}

	if !complete {
		return nil // Wait for other approvals
	}

	// Determine approval result
	approved := h.isApproved(node.ID, c)

	// Update node status
	status := string(types.ExecutionRejected)
	if approved {
		status = string(types.ExecutionCompleted)
		h.metrics.AddCounter("approval_approved", 1)
	} else {
		h.metrics.AddCounter("approval_rejected", 1)
	}

	// Complete node
	if _, err = h.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: status,
		},
	}); err != nil {
		return err
	}

	// Record completion metrics
	h.metrics.RecordValue("approval_complete_duration", time.Since(startTime).Seconds())

	// Notify completion
	h.notifyCompletion(ctx, node, approved)

	return nil
}

// validateInternal validates the approval node
func (h *ApprovalHandler) validateInternal(node *structs.ReadNode) error {
	c, err := h.parseApprovalConfig(node)
	if err != nil {
		return err
	}

	// Validate approvers configuration
	if len(c.Candidates) == 0 {
		if approvers, ok := node.Properties["approvers"].([]string); !ok || len(approvers) == 0 {
			return types.NewError(types.ErrValidation, "no approvers configured", nil)
		}
	}

	// Validate approval strategy
	if _, ok := h.strategies[c.Strategy]; !ok {
		return types.NewError(types.ErrValidation, "invalid approval strategy", nil)
	}

	// Validate timeout settings
	if c.TimeoutHours < 0 {
		return types.NewError(types.ErrValidation, "invalid timeout hours", nil)
	}

	return nil
}

// Rollback rollbacks node execution
func (h *ApprovalHandler) Rollback(ctx context.Context, node *structs.ReadNode) error {
	// Get all approval tasks
	tasks, err := h.services.Task.List(ctx, &structs.ListTaskParams{
		ProcessID: node.ProcessID,
		NodeKey:   node.NodeKey,
	})
	if err != nil {
		return err
	}

	// Cancel pending tasks
	for _, task := range tasks.Items {
		if task.Status == string(types.StatusPending) {
			_, err := h.services.Task.Update(ctx, &structs.UpdateTaskBody{
				ID: task.ID,
				TaskBody: structs.TaskBody{
					Status:  string(types.ExecutionCancelled),
					Comment: "Node rollback",
				},
			})
			if err != nil {
				return err
			}
		}
	}

	// Clear approval records
	h.approvals.Delete(node.ID)

	h.metrics.AddCounter("approval_rollback", 1)

	return nil
}

// rollbackInternal rollbacks the approval node
func (h *ApprovalHandler) rollbackInternal(ctx context.Context, node *structs.ReadNode) error {
	// Cancel all active approval tasks
	h.activeTasks.Range(func(key, value any) bool {
		taskInfo := value.(*ApprovalTask)
		if taskInfo.NodeID == node.ID {
			if err := h.cancelApprovalTask(ctx, taskInfo); err != nil {
				h.logger.Errorf(ctx, "Failed to cancel approval task %s: %v", taskInfo.ID, err)
			}
		}
		return true
	})

	// Clear approval records
	h.approvals.Delete(node.ID)

	return nil
}

// Validate validates the approval node
func (h *ApprovalHandler) Validate(node *structs.ReadNode) error {
	c, err := h.parseApprovalConfig(node)
	if err != nil {
		return err
	}

	// Validate approvers configuration
	if len(c.Candidates) == 0 {
		if approvers, ok := node.Properties["approvers"].([]string); !ok || len(approvers) == 0 {
			return types.NewError(types.ErrValidation, "no approvers configured", nil)
		}
	}

	// Validate approval strategy
	if _, ok := h.strategies[c.Strategy]; !ok {
		return types.NewError(types.ErrValidation, "invalid approval strategy", nil)
	}

	// Validate timeout settings
	if c.TimeoutHours < 0 {
		return types.NewError(types.ErrValidation, "invalid timeout hours", nil)
	}

	return nil
}

// IsHealthy returns handler health status
func (h *ApprovalHandler) IsHealthy() bool {
	return h.Status() == types.HandlerRunning
}

// GetMetrics returns approval handler metrics
func (h *ApprovalHandler) GetMetrics() map[string]any {
	if h.metrics == nil {
		return nil
	}

	return map[string]any{
		"created":            h.metrics.GetCounter("approval_created"),
		"create_failed":      h.metrics.GetCounter("approval_create_failed"),
		"task_created":       h.metrics.GetCounter("approval_task_created"),
		"task_create_failed": h.metrics.GetCounter("approval_task_create_failed"),
		"approved":           h.metrics.GetCounter("approval_approved"),
		"rejected":           h.metrics.GetCounter("approval_rejected"),
		"timeout":            h.metrics.GetCounter("approval_timeout"),
		"urged":              h.metrics.GetCounter("approval_urged"),
		"delegated":          h.metrics.GetCounter("approval_delegated"),
		"transferred":        h.metrics.GetCounter("approval_transferred"),
		"rollback":           h.metrics.GetCounter("approval_rollback"),
		"create_duration":    h.metrics.GetHistogram("approval_create_duration"),
		"complete_duration":  h.metrics.GetHistogram("approval_complete_duration"),
	}
}

// initMetrics initializes metrics collectors
func (h *ApprovalHandler) initMetrics() {
	h.metrics.RegisterCounter("approval_created")
	h.metrics.RegisterCounter("approval_create_failed")
	h.metrics.RegisterCounter("approval_task_created")
	h.metrics.RegisterCounter("approval_task_create_failed")
	h.metrics.RegisterCounter("approval_approved")
	h.metrics.RegisterCounter("approval_rejected")
	h.metrics.RegisterCounter("approval_timeout")
	h.metrics.RegisterCounter("approval_urged")
	h.metrics.RegisterCounter("approval_delegated")
	h.metrics.RegisterCounter("approval_transferred")
	h.metrics.RegisterCounter("approval_rollback")

	h.metrics.RegisterHistogram("approval_create_duration", 1000)
	h.metrics.RegisterHistogram("approval_complete_duration", 1000)
}

// createApprovalTasks creates approval tasks for approvers
func (h *ApprovalHandler) createApprovalTasks(ctx context.Context, node *structs.ReadNode, approvers []string, config *config.ApprovalHandlerConfig) error {
	// Sort approvers by priority if configured
	if _, ok := node.Properties["priority_approvers"].([]string); ok {
		h.sortApproversByPriority(approvers, node.Properties["priority_approvers"].([]string))
	}

	for i, approver := range approvers {
		task := &structs.TaskBody{
			Name:      fmt.Sprintf("Approval Task: %s", node.Name),
			ProcessID: node.ProcessID,
			NodeKey:   node.NodeKey,
			NodeType:  string(node.Type),
			Status:    string(types.StatusPending),
			Assignees: []string{approver},
			Priority:  i, // Priority based on order
			Extras: map[string]any{
				"allowDelegate": config.AllowDelegate,
				"allowTransfer": config.AllowTransfer,
				"allowUrge":     config.AllowUrge,
				"timeoutHours":  config.TimeoutHours,
				"strategy":      config.Strategy,
			},
		}

		// Set due time if timeout configured
		if config.TimeoutHours > 0 {
			dueTime := time.Now().Add(time.Duration(config.TimeoutHours) * time.Hour).UnixMilli()
			task.DueTime = &dueTime
		}

		// Create task
		createdTask, err := h.services.Task.Create(ctx, task)
		if err != nil {
			h.metrics.AddCounter("approval_task_create_failed", 1)
			return fmt.Errorf("create approval task failed: %w", err)
		}

		// Track active task
		dt := time.UnixMilli(*task.DueTime)
		h.activeTasks.Store(createdTask.ID, &ApprovalTask{
			ID:        createdTask.ID,
			NodeID:    node.ID,
			Approver:  approver,
			StartTime: time.Now(),
			DueTime:   &dt,
			Status:    string(types.StatusPending),
		})

		h.metrics.AddCounter("approval_task_created", 1)
	}

	return nil
}

// parseApprovalConfig parses approval configuration from node properties
func (h *ApprovalHandler) parseApprovalConfig(node *structs.ReadNode) (*config.ApprovalHandlerConfig, error) {
	c, ok := node.Properties["approvalConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing approval configuration", nil)
	}

	cfg := config.DefaultApprovalHandlerConfig()

	// Parse configuration fields
	if strategy, ok := c["strategy"].(string); ok {
		cfg.Strategy = strategy
	}
	if timeout, ok := c["timeout_hours"].(int); ok {
		cfg.TimeoutHours = timeout
	}
	if candidates, ok := c["candidates"].([]string); ok {
		cfg.Candidates = candidates
	}
	if allowDelegate, ok := c["allow_delegate"].(bool); ok {
		cfg.AllowDelegate = allowDelegate
	}
	if allowTransfer, ok := c["allow_transfer"].(bool); ok {
		cfg.AllowTransfer = allowTransfer
	}
	if allowUrge, ok := c["allow_urge"].(bool); ok {
		cfg.AllowUrge = allowUrge
	}
	if autoPass, ok := c["auto_pass"].(bool); ok {
		cfg.AutoPass = autoPass
	}
	if autoReject, ok := c["auto_reject"].(bool); ok {
		cfg.AutoReject = autoReject
	}
	if maxUrges, ok := c["max_urges"].(int); ok {
		cfg.MaxUrges = maxUrges
	}
	if urgeInterval, ok := c["urge_interval"].(float64); ok {
		cfg.UrgeInterval = time.Duration(urgeInterval) * time.Hour
	}
	if autoEscalate, ok := c["auto_escalate"].(bool); ok {
		cfg.AutoEscalate = autoEscalate
	}
	if delegateRules, ok := c["delegate_rules"].(map[string]any); ok {
		cfg.DelegateRules = &config.DelegateRules{}
		if roles, ok := delegateRules["allowed_roles"].([]string); ok {
			cfg.DelegateRules.AllowedRoles = roles
		}
		if maxDelegates, ok := delegateRules["max_delegates"].(int); ok {
			cfg.DelegateRules.MaxDelegates = maxDelegates
		}
		if duration, ok := delegateRules["duration"].(float64); ok {
			cfg.DelegateRules.Duration = time.Duration(duration) * time.Hour
		}
	}

	return cfg, nil
}

// resolveApprovers resolves approval candidates
func (h *ApprovalHandler) resolveApprovers(ctx context.Context, node *structs.ReadNode, config *config.ApprovalHandlerConfig) ([]string, error) {
	var approvers []string

	// Add static candidates
	if len(config.Candidates) > 0 {
		approvers = append(approvers, config.Candidates...)
	}

	// Resolve dynamic approvers
	if dynamic, ok := node.Properties["approvers"].([]string); ok {
		approvers = append(approvers, dynamic...)
	}

	// Resolve role based approvers
	if roles, ok := node.Properties["approver_roles"].([]string); ok {
		roleApprovers, err := h.resolveRoleApprovers(ctx, roles)
		if err != nil {
			return nil, err
		}
		approvers = append(approvers, roleApprovers...)
	}

	// Resolve department based approvers
	if depts, ok := node.Properties["approver_departments"].([]string); ok {
		deptApprovers, err := h.resolveDepartmentApprovers(ctx, depts)
		if err != nil {
			return nil, err
		}
		approvers = append(approvers, deptApprovers...)
	}

	return h.deduplicateApprovers(approvers), nil
}

// Role and department resolution implementations

type RoleInfo struct {
	ID       string
	Name     string
	Users    []string
	Level    int    // For role hierarchy, higher means more senior
	ParentID string // For role hierarchy
}

type DepartmentInfo struct {
	ID       string
	Name     string
	Users    []string
	Level    int // For org hierarchy
	ParentID string
	Managers []string
}

// resolveRoleApprovers resolves role based approvers relationship
func (h *ApprovalHandler) resolveRoleApprovers(ctx context.Context, roles []string) ([]string, error) {
	var approvers []string
	// TODO: service implementation
	// // Get roles from identity service
	// for _, roleID := range roles {
	// 	role, err := h.services.Identity.GetRole(ctx, roleID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get role %s: %w", roleID, err)
	// 	}
	//
	// 	// Get users with this role
	// 	users, err := h.services.Identity.GetUsersWithRole(ctx, roleID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get users for role %s: %w", roleID, err)
	// 	}
	//
	// 	approvers = append(approvers, users...)
	//
	// 	// Also include users with higher roles if configured
	// 	if h.config.IncludeHigherRoles {
	// 		higherRoles, err := h.getHigherRoles(ctx, role)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		for _, higherRole := range higherRoles {
	// 			users, err := h.services.Identity.GetUsersWithRole(ctx, higherRole.ID)
	// 			if err != nil {
	// 				continue
	// 			}
	// 			approvers = append(approvers, users...)
	// 		}
	// 	}
	// }

	return approvers, nil
}

// getHigherRoles gets roles higher in the hierarchy
func (h *ApprovalHandler) getHigherRoles(ctx context.Context, role *RoleInfo) ([]*RoleInfo, error) {
	var higherRoles []*RoleInfo
	// TODO: service implementation
	// current := role
	// for current.ParentID != "" {
	// 	parent, err := h.services.Identity.GetRole(ctx, current.ParentID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if parent.Level > current.Level {
	// 		higherRoles = append(higherRoles, parent)
	// 	}
	// 	current = parent
	// }

	return higherRoles, nil
}

// resolveDepartmentApprovers resolves department based approvers relationship
func (h *ApprovalHandler) resolveDepartmentApprovers(ctx context.Context, departments []string) ([]string, error) {
	var approvers []string
	// TODO: service implementation
	// for _, deptID := range departments {
	// 	dept, err := h.services.Identity.GetDepartment(ctx, deptID)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get department %s: %w", deptID, err)
	// 	}
	//
	// 	// Add department managers
	// 	approvers = append(approvers, dept.Managers...)
	//
	// 	// Add higher level managers if configured
	// 	if h.config.IncludeHigherManagers {
	// 		higherDepts, err := h.getHigherDepartments(ctx, dept)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	//
	// 		for _, higherDept := range higherDepts {
	// 			approvers = append(approvers, higherDept.Managers...)
	// 		}
	// 	}
	// }

	return approvers, nil
}

// getHigherDepartments gets departments higher in the org hierarchy
func (h *ApprovalHandler) getHigherDepartments(ctx context.Context, dept *DepartmentInfo) ([]*DepartmentInfo, error) {
	var higherDepts []*DepartmentInfo

	// TODO: service implementation
	// current := dept
	// for current.ParentID != "" {
	// 	parent, err := h.services.Identity.GetDepartment(ctx, current.ParentID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if parent.Level > current.Level {
	// 		higherDepts = append(higherDepts, parent)
	// 	}
	// 	current = parent
	// }

	return higherDepts, nil
}

// deduplicateApprovers removes duplicate approvers while maintaining order
func (h *ApprovalHandler) deduplicateApprovers(approvers []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(approvers))

	for _, approver := range approvers {
		if !seen[approver] {
			seen[approver] = true
			result = append(result, approver)
		}
	}
	return result
}

// checkCompletion checks if approval is complete
func (h *ApprovalHandler) checkCompletion(ctx context.Context, node *structs.ReadNode, config *config.ApprovalHandlerConfig) (bool, error) {
	// Get all approval tasks
	tasks, err := h.services.Task.List(ctx, &structs.ListTaskParams{
		ProcessID: node.ProcessID,
		NodeKey:   node.NodeKey,
	})
	if err != nil {
		return false, err
	}

	// Calculate completion statistics
	completed := 0
	approved := 0
	total := len(tasks.Items)

	for _, task := range tasks.Items {
		if task.Status == string(types.ExecutionCompleted) {
			completed++
			if task.Action == string(structs.ActionApprove) {
				approved++
			}
		}
	}

	// Get approval strategy
	strategy, ok := h.strategies[config.Strategy]
	if !ok {
		return false, types.NewError(types.ErrValidation, "invalid approval strategy", nil)
	}

	return strategy.IsCompleted(completed, approved, total), nil
}

// recordApproval records approval action
func (h *ApprovalHandler) recordApproval(nodeID string, record ApprovalRecord) {
	value, _ := h.approvals.LoadOrStore(nodeID, []ApprovalRecord{})
	records := value.([]ApprovalRecord)
	records = append(records, record)
	h.approvals.Store(nodeID, records)
}

// isApproved checks if approval is approved based on strategy
func (h *ApprovalHandler) isApproved(nodeID string, config *config.ApprovalHandlerConfig) bool {
	strategy, ok := h.strategies[config.Strategy]
	if !ok {
		return false
	}

	approvals, ok := h.approvals.Load(nodeID)
	if !ok {
		return false
	}

	records := approvals.([]ApprovalRecord)
	approved := 0
	total := len(records)

	for _, record := range records {
		if record.Action == string(structs.ActionApprove) {
			approved++
		}
	}

	return strategy.IsApproved(approved, total)
}

// registerDefaultStrategies registers default approval strategies
func (h *ApprovalHandler) registerDefaultStrategies() {
	h.strategies["any"] = &AnyApprovalStrategy{}
	h.strategies["all"] = &AllApprovalStrategy{}
	h.strategies["majority"] = &MajorityApprovalStrategy{}
	h.strategies["percentage"] = &PercentageApprovalStrategy{}
	h.strategies["order"] = &OrderApprovalStrategy{}
}

// Strategy implementations

type AnyApprovalStrategy struct{}

func (s *AnyApprovalStrategy) GetName() string {
	return "any"
}

func (s *AnyApprovalStrategy) IsCompleted(completed, approved, total int) bool {
	return approved > 0 || completed == total
}

func (s *AnyApprovalStrategy) IsApproved(approved, total int) bool {
	return approved > 0
}

type AllApprovalStrategy struct{}

func (s *AllApprovalStrategy) GetName() string {
	return "all"
}

func (s *AllApprovalStrategy) IsCompleted(completed, approved, total int) bool {
	return completed == total
}

func (s *AllApprovalStrategy) IsApproved(approved, total int) bool {
	return approved == total
}

type MajorityApprovalStrategy struct{}

func (s *MajorityApprovalStrategy) GetName() string {
	return "majority"
}

func (s *MajorityApprovalStrategy) IsCompleted(completed, approved, total int) bool {
	return completed == total
}

func (s *MajorityApprovalStrategy) IsApproved(approved, total int) bool {
	return approved > total/2
}

// PercentageApprovalStrategy - specific percentage must approve
type PercentageApprovalStrategy struct {
	RequiredPercentage float64
}

func (s *PercentageApprovalStrategy) GetName() string {
	return "percentage"
}

func (s *PercentageApprovalStrategy) IsCompleted(completed, approved, total int) bool {
	return completed == total
}

func (s *PercentageApprovalStrategy) IsApproved(approved, total int) bool {
	if total == 0 {
		return false
	}
	percentage := float64(approved) / float64(total) * 100
	return percentage >= s.RequiredPercentage
}

// OrderApprovalStrategy - must approve in specific order
type OrderApprovalStrategy struct {
	CurrentIndex int
	Order        []string
}

func (s *OrderApprovalStrategy) GetName() string {
	return "order"
}

func (s *OrderApprovalStrategy) IsCompleted(completed, approved, total int) bool {
	return s.CurrentIndex >= len(s.Order) || completed == total
}

func (s *OrderApprovalStrategy) IsApproved(approved, total int) bool {
	// All previous approvers in order must have approved
	return approved >= s.CurrentIndex+1
}

// Task management methods

// handleTaskTimeout handles task timeout
func (h *ApprovalHandler) handleTaskTimeout(task *structs.ReadTask, config *config.ApprovalHandlerConfig) error {
	if config.AutoPass {
		// Auto approve on timeout
		return h.autoCompleteTask(task, string(structs.ActionApprove), "Auto approved due to timeout")
	} else if config.AutoReject {
		// Auto reject on timeout
		return h.autoCompleteTask(task, string(structs.ActionReject), "Auto rejected due to timeout")
	}

	// Just mark as timeout
	_, err := h.services.Task.Update(context.Background(), &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:  string(types.ExecutionTimeout),
			Comment: "Task timed out",
		},
	})

	if err == nil {
		h.metrics.AddCounter("approval_timeout", 1)
	}

	return err
}

// autoCompleteTask completes a task automatically
func (h *ApprovalHandler) autoCompleteTask(task *structs.ReadTask, action string, comment string) error {
	req := &structs.CompleteTaskRequest{
		TaskID:   task.ID,
		Action:   structs.ActionType(action),
		Operator: "system",
		Comment:  comment,
	}

	_, err := h.services.Task.Complete(context.Background(), req)
	if err != nil {
		return fmt.Errorf("auto complete task failed: %w", err)
	}

	return nil
}

// cancelApprovalTask cancels an approval task
func (h *ApprovalHandler) cancelApprovalTask(ctx context.Context, task *ApprovalTask) error {
	_, err := h.services.Task.Update(ctx, &structs.UpdateTaskBody{
		ID: task.ID,
		TaskBody: structs.TaskBody{
			Status:  string(types.ExecutionCancelled),
			Comment: "Task cancelled by system",
		},
	})

	if err == nil {
		h.activeTasks.Delete(task.ID)
	}

	return err
}

// Task urging and delegation

// urgeTask urges an approval task
func (h *ApprovalHandler) urgeTask(ctx context.Context, taskID string) error {
	taskInfo, ok := h.activeTasks.Load(taskID)
	if !ok {
		return types.NewError(types.ErrNotFound, "task not found", nil)
	}

	task := taskInfo.(*ApprovalTask)
	task.mu.Lock()
	defer task.mu.Unlock()

	// Check urge interval
	if task.LastUrgeTime != nil && time.Since(*task.LastUrgeTime) < h.config.UrgeInterval {
		return types.NewError(types.ErrValidation, "urge too frequent", nil)
	}

	// Check max urges
	if task.UrgeCount >= h.config.MaxUrges {
		if h.config.AutoEscalate {
			return h.escalateTask(ctx, task)
		}
		return types.NewError(types.ErrValidation, "max urges exceeded", nil)
	}

	// Send urge notification
	now := time.Now()
	task.UrgeCount++
	task.LastUrgeTime = &now

	req := &structs.UrgeTaskRequest{
		TaskID:   task.ID,
		Operator: "system",
		Comment:  fmt.Sprintf("Task urged for the %d time", task.UrgeCount),
	}

	err := h.services.Task.Urge(ctx, req)
	if err != nil {
		return fmt.Errorf("urge task failed: %w", err)
	}

	h.metrics.AddCounter("approval_urged", 1)
	return nil
}

// delegateTask delegates an approval task
func (h *ApprovalHandler) delegateTask(ctx context.Context, taskID string, delegator string, delegate string) error {
	// Validate delegate rules
	if err := h.validateDelegateRules(delegator, delegate); err != nil {
		return err
	}

	req := &structs.DelegateTaskRequest{
		TaskID:    taskID,
		Delegator: delegator,
		Delegate:  delegate,
		Comment:   "Task delegated",
	}

	if err := h.services.Task.Delegate(ctx, req); err != nil {
		return fmt.Errorf("delegate task failed: %w", err)
	}

	h.metrics.AddCounter("approval_delegated", 1)
	return nil
}

// escalateTask escalates an approval task
func (h *ApprovalHandler) escalateTask(ctx context.Context, task *ApprovalTask) error {
	// Get escalation approvers
	escalationApprovers, err := h.getEscalationApprovers(ctx, task)
	if err != nil {
		return err
	}

	// Create escalation task
	escalationTask := &structs.TaskBody{
		Name:      fmt.Sprintf("Escalated: %s", task.ID),
		ProcessID: task.NodeID, // Use same process
		NodeKey:   task.NodeID,
		Status:    string(types.StatusPending),
		Assignees: escalationApprovers,
		Priority:  0, // High priority
		Extras: map[string]any{
			"original_task": task.ID,
			"escalated":     true,
		},
	}

	_, err = h.services.Task.Create(ctx, escalationTask)
	if err != nil {
		return fmt.Errorf("create escalation task failed: %w", err)
	}

	// Cancel original task
	task.Status = string(types.ExecutionCancelled)
	h.metrics.AddCounter("approval_escalated", 1)

	return nil
}

// Delegate rules validation

type DelegationRule struct {
	MaxDelegates    int
	ValidDuration   time.Duration
	AllowedRoles    []string
	ExcludedUsers   []string
	RequireApproval bool
}

func (h *ApprovalHandler) validateDelegateRules(delegator string, delegate string) error {
	rules := h.config.DelegateRules
	if rules == nil {
		return types.NewError(types.ErrValidation, "delegation not configured", nil)
	}

	// TODO: service implementation
	// // Check if delegate is excluded
	// for _, excluded := range rules.ExcludedUsers {
	// 	if delegate == excluded {
	// 		return types.NewError(types.ErrValidation, fmt.Sprintf("user %s is excluded from delegation", delegate), nil)
	// 	}
	// }
	//
	// // Check role restrictions
	// if len(rules.AllowedRoles) > 0 {
	// 	delegateRoles, err := h.services.Identity.GetUserRoles(context.Background(), delegate)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to get delegate roles: %w", err)
	// 	}
	//
	// 	hasAllowedRole := false
	// 	for _, role := range delegateRoles {
	// 		for _, allowed := range rules.AllowedRoles {
	// 			if role == allowed {
	// 				hasAllowedRole = true
	// 				break
	// 			}
	// 		}
	// 	}
	//
	// 	if !hasAllowedRole {
	// 		return types.NewError(types.ErrValidation, "delegate does not have required role", nil)
	// 	}
	// }
	//
	// // Check max delegates
	// currentDelegates, err := h.getCurrentDelegates(delegator)
	// if err != nil {
	// 	return err
	// }
	// if len(currentDelegates) >= rules.MaxDelegates {
	// 	return types.NewError(types.ErrValidation, "max delegates exceeded", nil)
	// }

	return nil
}

func (h *ApprovalHandler) getCurrentDelegates(delegator string) ([]string, error) {
	// Query active delegations
	delegations, err := h.services.Delegation.GetActiveDelegations(context.Background(), delegator)
	if err != nil {
		return nil, fmt.Errorf("failed to get active delegations: %w", err)
	}

	delegates := make([]string, 0)
	for _, d := range delegations {
		delegates = append(delegates, d.DelegateeID)
	}

	return delegates, nil
}

// Escalation approvers resolution

type EscalationConfig struct {
	Roles       []string // Roles to escalate to
	Departments []string // Departments to escalate to
	SkipLevels  int      // Levels to skip in hierarchy
	MaxLevels   int      // Maximum levels to escalate
	AutoApprove bool     // Auto approve after max levels
}

func (h *ApprovalHandler) getEscalationApprovers(ctx context.Context, task *ApprovalTask) ([]string, error) {
	var approvers []string

	// Get escalation config
	c := h.getEscalationConfig(task)
	if c == nil {
		return nil, types.NewError(types.ErrValidation, "escalation not configured", nil)
	}

	// TODO: service implementation
	// // Get original task assignees
	// originalTask, err := h.services.Task.Get(ctx, &structs.FindTaskParams{
	// 	ID: task.ID,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	//
	// // Get roles/departments of original assignees
	// for _, assignee := range originalTask.Assignees {
	// 	userInfo, err := h.services.Identity.GetUser(ctx, assignee)
	// 	if err != nil {
	// 		continue
	// 	}
	//
	// 	// Get higher level roles
	// 	roles, err := h.services.Identity.GetUserRoles(ctx, assignee)
	// 	if err == nil {
	// 		for _, role := range roles {
	// 			roleInfo, err := h.services.Identity.GetRole(ctx, role)
	// 			if err != nil {
	// 				continue
	// 			}
	//
	// 			higherRoles, err := h.getHigherRolesWithSkip(ctx, roleInfo, c.SkipLevels, c.MaxLevels)
	// 			if err == nil {
	// 				for _, higherRole := range higherRoles {
	// 					users, _ := h.services.Identity.GetUsersWithRole(ctx, higherRole.ID)
	// 					approvers = append(approvers, users...)
	// 				}
	// 			}
	// 		}
	// 	}
	//
	// 	// Get higher level departments
	// 	deptID := userInfo.DepartmentID
	// 	if deptID != "" {
	// 		dept, err := h.services.Identity.GetDepartment(ctx, deptID)
	// 		if err == nil {
	// 			higherDepts, err := h.getHigherDepartmentsWithSkip(ctx, dept, c.SkipLevels, c.MaxLevels)
	// 			if err == nil {
	// 				for _, higherDept := range higherDepts {
	// 					approvers = append(approvers, higherDept.Managers...)
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	//
	// // Add configured roles
	// for _, roleID := range c.Roles {
	// 	users, err := h.services.Identity.GetUsersWithRole(ctx, roleID)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	approvers = append(approvers, users...)
	// }
	//
	// // Add configured departments
	// for _, deptID := range c.Departments {
	// 	dept, err := h.services.Identity.GetDepartment(ctx, deptID)
	// 	if err != nil {
	// 		continue
	// 	}
	// 	approvers = append(approvers, dept.Managers...)
	// }

	return h.deduplicateApprovers(approvers), nil
}

func (h *ApprovalHandler) getEscalationConfig(task *ApprovalTask) *EscalationConfig {
	// Try get from task config
	if c, ok := task.Variables["escalation_config"].(*EscalationConfig); ok {
		return c
	}

	// Return default config
	return &EscalationConfig{
		SkipLevels:  1,
		MaxLevels:   3,
		AutoApprove: true,
	}
}

func (h *ApprovalHandler) getHigherRolesWithSkip(ctx context.Context, role *RoleInfo, skip int, max int) ([]*RoleInfo, error) {
	var higherRoles []*RoleInfo

	// TODO: service implementation
	// current := role
	// skipped := 0
	// levels := 0
	//
	// for current.ParentID != "" && levels < max {
	// 	parent, err := h.services.Identity.GetRole(ctx, current.ParentID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if parent.Level > current.Level {
	// 		if skipped >= skip {
	// 			higherRoles = append(higherRoles, parent)
	// 			levels++
	// 		} else {
	// 			skipped++
	// 		}
	// 	}
	// 	current = parent
	// }

	return higherRoles, nil
}

func (h *ApprovalHandler) getHigherDepartmentsWithSkip(ctx context.Context, dept *DepartmentInfo, skip int, max int) ([]*DepartmentInfo, error) {
	var higherDepts []*DepartmentInfo

	// TODO: service implementation
	// current := dept
	// skipped := 0
	// levels := 0
	//
	// for current.ParentID != "" && levels < max {
	// 	parent, err := h.services.Identity.GetDepartment(ctx, current.ParentID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if parent.Level > current.Level {
	// 		if skipped >= skip {
	// 			higherDepts = append(higherDepts, parent)
	// 			levels++
	// 		} else {
	// 			skipped++
	// 		}
	// 	}
	// 	current = parent
	// }

	return higherDepts, nil
}

func (h *ApprovalHandler) sortApproversByPriority(approvers []string, priority []string) {
	// Create priority map for O(1) lookup
	priorityMap := make(map[string]int)
	for i, approver := range priority {
		priorityMap[approver] = i
	}

	// Sort approvers based on priority
	sort.Slice(approvers, func(i, j int) bool {
		pi := priorityMap[approvers[i]]
		pj := priorityMap[approvers[j]]

		// Higher priority (lower index) comes first
		if pi != pj {
			return pi < pj
		}
		// If same priority, maintain original order
		return i < j
	})
}

// Notification methods

func (h *ApprovalHandler) notifyCompletion(ctx context.Context, node *structs.ReadNode, approved bool) {
	details := map[string]any{
		"node_id":    node.ID,
		"process_id": node.ProcessID,
		"approved":   approved,
		"time":       time.Now(),
	}

	if approved {
		h.em.PublishEvent("approval.approved", details)
	} else {
		h.em.PublishEvent("approval.rejected", details)
	}
}
