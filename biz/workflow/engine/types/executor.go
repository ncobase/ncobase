package types

import (
	"context"
	"time"
)

// ExecutorType represents type of executor
type ExecutorType string

const (
	ProcessExecutor ExecutorType = "process"
	NodeExecutor    ExecutorType = "node"
	TaskExecutor    ExecutorType = "task"
	RetryExecutor   ExecutorType = "retry"
	ServiceExecutor ExecutorType = "service"
)

// String implements fmt.Stringer
func (t ExecutorType) String() string {
	return string(t)
}

// Executor represents base executor interface
type Executor interface {
	// Basic info

	Type() ExecutorType
	Name() string
	Priority() int

	// Lifecycle

	Start() error
	Stop() error
	Pause() error
	Resume() error
	Reset() error

	// Core operations

	Execute(ctx context.Context, req *Request) (*Response, error)
	Cancel(ctx context.Context, id string) error
	Rollback(ctx context.Context, id string) error

	// Status & capabilities

	Status() ExecutionStatus
	IsHealthy() bool
	GetMetrics() map[string]any
	GetCapabilities() *ExecutionCapabilities

	// Resource management

	SetMaxConcurrent(max int32)
	SetTimeout(timeout time.Duration)
	ResourceUsage() *ResourceUsage
}

// Request represents an execution request
type Request struct {
	ID        string            // Request ID
	Type      ExecutorType      // Request type
	Priority  int               // Request priority
	Timeout   time.Duration     // Request timeout
	Context   map[string]any    // Request context
	Variables map[string]any    // Request variables
	Metadata  map[string]string // Request metadata
}

// Response represents an execution response
type Response struct {
	ID        string            // Response ID
	Status    ExecutionStatus   // Response status
	Data      any               // Response data
	Error     error             // Response error
	StartTime time.Time         // Start time
	EndTime   *time.Time        // End time
	Duration  time.Duration     // Execution duration
	Metadata  map[string]string // Response metadata
}

// ExecutionCapabilities represents executor capabilities
type ExecutionCapabilities struct {
	SupportsAsync    bool     // Support async execution
	SupportsRetry    bool     // Support retry
	SupportsRollback bool     // Support rollback
	MaxConcurrency   int      // Max concurrent executions
	MaxBatchSize     int      // Max batch size
	AllowedActions   []string // Allowed actions
}
