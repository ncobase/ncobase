package types

import (
	"context"
)

// Engine represents the workflow engine interface
type Engine interface {
	// Lifecycle

	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Status() EngineStatus

	// Process operations

	StartProcess(ctx context.Context, req *StartProcessRequest) (*StartProcessResponse, error)
	CompleteProcess(ctx context.Context, processID string) error
	TerminateProcess(ctx context.Context, req *TerminateRequest) error
	SuspendProcess(ctx context.Context, processID string, reason string) error
	ResumeProcess(ctx context.Context, processID string) error

	// Query operations

	GetProcessStatus(ctx context.Context, processID string) (*ProcessStatus, error)
	GetActiveNodes(ctx context.Context, processID string) ([]*Node, error)
	GetNextNodes(ctx context.Context, processID string, nodeID string) ([]*Node, error)

	// Metrics & monitoring

	GetMetrics() map[string]any
}

// ProcessStatus represents process execution status
type ProcessStatus struct {
	ProcessID   string          // Process ID
	Status      ExecutionStatus // Current status
	CurrentNode string          // Current node
	ActiveNodes []any           // Active nodes
	Variables   map[string]any  // Process variables
	StartTime   *int64          // Start time
	EndTime     *int64          // End time
}

// ResourceUsage represents executor resource usage
type ResourceUsage struct {
	MemoryUsage    int64   // Memory usage in bytes
	CPUUsage       float64 // CPU usage percentage
	GoroutineCount int     // Number of goroutines
	ActiveJobs     int64   // Number of active jobs
}
