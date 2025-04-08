package handler

import (
	"context"
	"fmt"
	"ncobase/core/workflow/engine/config"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"ncore/extension"
	"sort"
	"strings"
	"sync"
	"time"
)

// BranchStatus represents branch execution status
type BranchStatus struct {
	NodeKey   string
	Status    structs.Status
	StartTime time.Time
	EndTime   *time.Time
	Error     error
}

// ParallelHandler handles parallel gateway nodes
type ParallelHandler struct {
	*BaseHandler
	// Configuration
	config *config.ParallelHandlerConfig

	// Active branches
	activeNodes sync.Map // Tracks active parallel branches
}

// NewParallelHandler creates a new parallel handler
func NewParallelHandler(svc *service.Service, em *extension.Manager, cfg *config.Config) *ParallelHandler {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &ParallelHandler{
		BaseHandler: NewBaseHandler("parallel", "Parallel Gateway Handler", svc, em, cfg.Handlers.Base),
		config:      cfg.Handlers.Parallel,
		activeNodes: sync.Map{},
	}
}

// Type returns handler type
func (h *ParallelHandler) Type() types.HandlerType { return h.handlerType }

// Name returns handler name
func (h *ParallelHandler) Name() string { return h.name }

// Priority returns handler priority
func (h *ParallelHandler) Priority() int { return h.priority }

// executeInternal executes the parallel node
func (h *ParallelHandler) executeInternal(ctx context.Context, node *structs.ReadNode) error {
	// Parse config
	c, err := h.parseParallelConfig(node)
	if err != nil {
		return err
	}

	// Create execution context with timeout
	timeoutCtx := ctx
	if c.Timeout > 0 {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, time.Duration(c.Timeout)*time.Second)
		defer cancel()
	}

	// Execute branches
	return h.executeBranches(timeoutCtx, node, c)
}

// completeInternal completes the parallel node
func (h *ParallelHandler) completeInternal(ctx context.Context, node *structs.ReadNode, _ *structs.CompleteTaskRequest) error {
	// Parse config
	c, err := h.parseParallelConfig(node)
	if err != nil {
		return err
	}

	// Check if all required branches are completed
	completed, err := h.checkBranchesCompletion(ctx, node, c)
	if err != nil {
		return err
	}

	if !completed {
		return nil // Wait for other branches
	}

	// Complete gateway node
	return h.completeGateway(ctx, node)
}

// validateInternal validates the parallel node
func (h *ParallelHandler) validateInternal(node *structs.ReadNode) error {
	// Parse and validate config
	c, err := h.parseParallelConfig(node)
	if err != nil {
		return err
	}

	// Validate branches
	for _, branch := range c.Branches {
		// Validate node exists
		if err := h.validateNodeExists(node.ProcessID, branch.NodeKey); err != nil {
			return fmt.Errorf("invalid branch node: %w", err)
		}
	}

	return nil
}

// rollbackInternal rollbacks the parallel node
func (h *ParallelHandler) rollbackInternal(ctx context.Context, node *structs.ReadNode) error {
	// Get all active branches
	branches := make([]string, 0)
	h.activeNodes.Range(func(key, value any) bool {
		if strings.HasPrefix(key.(string), node.ID) {
			branches = append(branches, value.(string))
		}
		return true
	})

	// Rollback each active branch
	for _, branchKey := range branches {
		if err := h.rollbackBranch(ctx, node.ProcessID, branchKey); err != nil {
			return err
		}
	}

	return nil
}

// parseParallelConfig parses the parallel gateway config
func (h *ParallelHandler) parseParallelConfig(node *structs.ReadNode) (*config.ParallelHandlerConfig, error) {
	c, ok := node.Properties["gatewayConfig"].(map[string]any)
	if !ok {
		return nil, types.NewError(types.ErrValidation, "missing gateway configuration", nil)
	}

	result := &config.ParallelHandlerConfig{
		CompleteMode:  "all",  // Default all branches must complete
		ErrorMode:     "fail", // Default fail on error
		MaxConcurrent: 10,     // Default max concurrent branches
	}

	// Parse branches
	if branches, ok := c["branches"].([]any); ok {
		result.Branches = make([]config.ParallelHandlerBranch, 0, len(branches))
		for _, b := range branches {
			branch, ok := b.(map[string]any)
			if !ok {
				continue
			}

			nodeKey, ok := branch["node_key"].(string)
			if !ok {
				continue
			}

			newBranch := config.ParallelHandlerBranch{
				NodeKey:  nodeKey,
				Required: true, // Default required
			}

			// Parse optional fields
			if priority, ok := branch["priority"].(int); ok {
				newBranch.Priority = priority
			}
			if required, ok := branch["required"].(bool); ok {
				newBranch.Required = required
			}
			if condition, ok := branch["condition"].(string); ok {
				newBranch.Condition = condition
			}
			if variables, ok := branch["variables"].(map[string]any); ok {
				newBranch.Variables = variables
			}

			result.Branches = append(result.Branches, newBranch)
		}
	}

	// Parse other options
	if mode := c["complete_mode"].(string); mode != "" {
		result.CompleteMode = mode
	}
	if mode := c["error_mode"].(string); mode != "" {
		result.ErrorMode = mode
	}
	if timeout := c["timeout"].(int); timeout > 0 {
		result.Timeout = timeout
	}
	if maxConcurrent := c["max_concurrent"].(int); maxConcurrent > 0 {
		result.MaxConcurrent = maxConcurrent
	}

	return result, nil
}

// executeBranches executes the parallel branches
func (h *ParallelHandler) executeBranches(ctx context.Context, node *structs.ReadNode, cfg *config.ParallelHandlerConfig) error {
	// Sort branches by priority
	branches := make([]config.ParallelHandlerBranch, len(cfg.Branches))
	copy(branches, cfg.Branches)
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Priority > branches[j].Priority
	})

	// Create semaphore for concurrency control
	sem := make(chan struct{}, cfg.MaxConcurrent)
	var wg sync.WaitGroup
	errChan := make(chan error, len(branches))

	// Execute branches
	for _, branch := range branches {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Acquire semaphore
		sem <- struct{}{}
		wg.Add(1)

		go func(b config.ParallelHandlerBranch) {
			defer wg.Done()
			defer func() { <-sem }()

			// Track active branch
			branchKey := fmt.Sprintf("%s-%s", node.ID, b.NodeKey)
			h.activeNodes.Store(branchKey, b.NodeKey)
			defer h.activeNodes.Delete(branchKey)

			if err := h.executeBranch(ctx, node.ProcessID, b); err != nil {
				errChan <- err
			}
		}(branch)
	}

	// Wait for all branches
	wg.Wait()
	close(errChan)

	// Check errors based on error mode
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 && cfg.ErrorMode == "fail" {
		return fmt.Errorf("branch execution failed: %v", errs)
	}

	return nil
}

// executeBranch executes a parallel branch
func (h *ParallelHandler) executeBranch(ctx context.Context, processID string, branch config.ParallelHandlerBranch) error {
	// Create branch task
	task := &structs.TaskBody{
		ProcessID: processID,
		NodeKey:   branch.NodeKey,
		Status:    string(structs.StatusPending),
		Variables: branch.Variables,
	}

	if _, err := h.services.Task.Create(ctx, task); err != nil {
		return fmt.Errorf("failed to create branch task: %w", err)
	}

	// Execute branch node
	req := &structs.NodeBody{
		ProcessID: processID,
		NodeKey:   branch.NodeKey,
	}

	_, err := h.services.Node.Create(ctx, req)
	return err
}

// checkBranchesCompletion checks if all branches are completed
func (h *ParallelHandler) checkBranchesCompletion(ctx context.Context, node *structs.ReadNode, config *config.ParallelHandlerConfig) (bool, error) {
	tasks, err := h.services.Task.List(ctx, &structs.ListTaskParams{
		ProcessID: node.ProcessID,
		NodeKey:   node.NodeKey,
	})
	if err != nil {
		return false, err
	}

	completed := 0
	successful := 0
	failed := 0
	total := len(config.Branches)

	for _, task := range tasks.Items {
		if task.Status == string(structs.StatusCompleted) {
			completed++
			successful++
		} else if task.Status == string(structs.StatusError) {
			completed++
			failed++
		}
	}

	// Check completion based on mode
	switch config.CompleteMode {
	case "all":
		return completed == total, nil
	case "any":
		return successful > 0, nil
	case "majority":
		return successful > total/2, nil
	default:
		return false, types.NewError(types.ErrValidation, "invalid completion mode", nil)
	}
}

func (h *ParallelHandler) completeGateway(ctx context.Context, node *structs.ReadNode) error {
	// Update node status
	_, err := h.services.Node.Update(ctx, &structs.UpdateNodeBody{
		ID: node.ID,
		NodeBody: structs.NodeBody{
			Status: string(structs.StatusCompleted),
		},
	})

	return err
}

// rollbackBranch rollbacks a parallel branch
func (h *ParallelHandler) rollbackBranch(ctx context.Context, processID string, nodeKey string) error {
	// Get branch node
	node, err := h.services.Node.Get(ctx, &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	if err != nil {
		return err
	}

	// Execute rollback
	return h.services.Node.UpdateStatus(ctx, node.ID, node.Status)
}

// validateNodeExists checks if a node exists
func (h *ParallelHandler) validateNodeExists(processID string, nodeKey string) error {
	_, err := h.services.Node.Get(context.Background(), &structs.FindNodeParams{
		ProcessID: processID,
		NodeKey:   nodeKey,
	})
	return err
}
