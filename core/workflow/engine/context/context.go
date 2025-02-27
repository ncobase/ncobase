package context

import (
	"context"
	"fmt"
	"ncobase/core/workflow/structs"
	"sync"
	"time"
)

// Context represents the context for execution
type Context struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	// Basic info
	ProcessID  string
	NodeID     string
	TaskID     string
	OperatorID string
	StartTime  time.Time

	// Runtime data
	variables *Variables
	Metadata  map[string]string

	// Process data
	Process *structs.ReadProcess
	Node    *structs.ReadNode
	Task    *structs.ReadTask

	// State
	state  State
	errors []error
}

// State represents context state
type State struct {
	IsRetry    bool
	RetryCount int
	IsTimeout  bool
	IsCanceled bool
	Phase      ExecutionPhase
}

// ExecutionPhase represents execution phase
type ExecutionPhase string

const (
	PhaseInitializing ExecutionPhase = "initializing"
	PhaseExecuting    ExecutionPhase = "executing"
	PhaseCompleting   ExecutionPhase = "completing"
	PhaseRollback     ExecutionPhase = "rollback"
	PhaseCompleted    ExecutionPhase = "completed"
	PhaseFailed       ExecutionPhase = "failed"
)

// NewContext creates a new context
//
// Usage:
//
//	// Create a basic context
//	ctx := NewContext(context.Background())
//
//	// Set metadata
//	ctx.SetMetadata("initiator", "user-123")
//	ctx.SetMetadata("priority", "high")
//
//	// Set variables
//	ctx.SetVariable("orderId", "ORD-123")
//	ctx.SetVariable("approvalRules", map[string]interface{}{
//	    "minAmount": 1000,
//	    "requiredApprovers": 2,
//	})
//
//	// Process execution
//	ctx.SetPhase(PhaseExecuting)
//	if err := processApprovalRequest(); err != nil {
//	    ctx.AddError(err)
//	    ctx.SetPhase(PhaseFailed)
//	}
//
//	// Handle timeout
//	select {
//	case <-ctx.Done():
//	    // Workflow was cancelled
//	    if err := ctx.Err(); err != nil {
//	        // Handle cancellation
//	    }
//	case <-time.After(approvalTimeout):
//	    ctx.MarkTimeout()
//	}
//
//	// Retry task
//	if needsRetry {
//	    ctx.MarkRetry()
//	}
//
//	// Clone context for parallel tasks
//	clonedCtx, err := ctx.Clone()
//	if err != nil {
//	    // Handle clone error
//	}
//
//	// Access current process
//	if processID := ctx.ProcessID; processID != "" {
//	    // Process task
//	}
func NewContext(parent context.Context) *Context {
	ctx, cancel := context.WithCancel(parent)

	return &Context{
		ctx:       ctx,
		cancel:    cancel,
		variables: NewVariables(),
		Metadata:  make(map[string]string),
		StartTime: time.Now(),
		state: State{
			Phase: PhaseInitializing,
		},
	}
}

// Context operations

func (wc *Context) Context() context.Context {
	return wc.ctx
}

func (wc *Context) Cancel() {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.state.IsCanceled = true
	wc.cancel()
}

func (wc *Context) Done() <-chan struct{} {
	return wc.ctx.Done()
}

func (wc *Context) Err() error {
	return wc.ctx.Err()
}

// WithProcess sets the process and merges its variables
func (wc *Context) WithProcess(process *structs.ReadProcess) *Context {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.Process = process
	wc.ProcessID = process.ID
	if process.Variables != nil {
		if err := wc.variables.MergeJSON(process.Variables); err != nil {
			fmt.Printf("Failed to merge process variables: %v, keeping existing variables\n", err)
		}
	}
	return wc
}

func (wc *Context) WithNode(node *structs.ReadNode) *Context {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.Node = node
	wc.NodeID = node.ID
	return wc
}

func (wc *Context) WithTask(task *structs.ReadTask) *Context {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.Task = task
	wc.TaskID = task.ID
	return wc
}

func (wc *Context) WithOperator(operatorID string) *Context {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.OperatorID = operatorID
	return wc
}

// State operations

func (wc *Context) SetPhase(phase ExecutionPhase) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.state.Phase = phase
}

func (wc *Context) GetPhase() ExecutionPhase {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	return wc.state.Phase
}

func (wc *Context) MarkRetry() {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.state.IsRetry = true
	wc.state.RetryCount++
}

func (wc *Context) MarkTimeout() {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.state.IsTimeout = true
}

func (wc *Context) AddError(err error) {
	if err == nil {
		return
	}

	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.errors = append(wc.errors, err)
}

func (wc *Context) GetErrors() []error {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	return wc.errors
}

func (wc *Context) HasErrors() bool {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	return len(wc.errors) > 0
}

// Metadata operations

func (wc *Context) GetMetadata(key string) string {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	return wc.Metadata[key]
}

func (wc *Context) SetMetadata(key string, value string) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	wc.Metadata[key] = value
}

// Variable operations

func (wc *Context) GetVariable(key string) (any, error) {
	return wc.variables.Get(key)
}

func (wc *Context) SetVariable(key string, value any) error {
	return wc.variables.Set(key, value)
}

// Clone creates a copy of the context
func (wc *Context) Clone() (*Context, error) {
	wc.mu.RLock()
	defer wc.mu.RUnlock()

	newCtx := NewContext(wc.ctx)

	// Copy basic info
	newCtx.ProcessID = wc.ProcessID
	newCtx.NodeID = wc.NodeID
	newCtx.TaskID = wc.TaskID
	newCtx.OperatorID = wc.OperatorID
	newCtx.StartTime = wc.StartTime

	// Copy runtime data
	clonedVars, err := wc.variables.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone variables: %w", err)
	}
	newCtx.variables = clonedVars

	// Copy metadata
	for k, v := range wc.Metadata {
		newCtx.Metadata[k] = v
	}

	// Copy process data
	newCtx.Process = wc.Process
	newCtx.Node = wc.Node
	newCtx.Task = wc.Task

	// Copy state
	newCtx.state = wc.state
	newCtx.errors = append([]error{}, wc.errors...)

	return newCtx, nil
}
