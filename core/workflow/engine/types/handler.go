package types

import (
	"context"
	"encoding/json"
	"ncobase/core/workflow/structs"
	"time"
)

// HandlerType represents type of handler
type HandlerType string

const (
	ServiceHandler      HandlerType = "service"
	ApprovalHandler     HandlerType = "approval"
	NotificationHandler HandlerType = "notification"
	ScriptHandler       HandlerType = "script"
	TimerHandler        HandlerType = "timer"
	ParallelHandler     HandlerType = "parallel"
	SubProcessHandler   HandlerType = "subprocess"
	ExclusiveHandler    HandlerType = "exclusive"
)

func (h HandlerType) String() string {
	return string(h)
}

// Handler represents base handler interface
type Handler interface {
	// Basic info

	Type() HandlerType
	Name() string
	Priority() int

	// Lifecycle

	Start() error
	Stop() error
	Reset() error

	// Core operations

	Execute(ctx context.Context, node *structs.ReadNode) error
	Complete(ctx context.Context, node *structs.ReadNode, req *structs.CompleteTaskRequest) error
	Validate(node *structs.ReadNode) error
	Rollback(ctx context.Context, node *structs.ReadNode) error
	Cancel(ctx context.Context, nodeID string) error

	// Status & metrics

	Status() HandlerStatus
	IsHealthy() bool
	GetMetrics() map[string]any
	GetState() *HandlerState
	GetCapabilities() *HandlerCapabilities
}

// HandlerConfig represents handler configuration
type HandlerConfig struct {
	// Basic settings
	MaxRetries    int           // Maximum retry attempts
	RetryInterval time.Duration // Retry interval
	Timeout       time.Duration // Execution timeout

	// Execution
	AsyncMode  bool // Enable async execution
	BatchMode  bool // Enable batch mode
	BatchSize  int  // Maximum batch size
	MaxWorkers int  // Maximum workers

	// Resource limits
	MaxMemory int64 // Maximum memory usage
	MaxCPU    int   // Maximum CPU usage

	// Features
	StrictMode    bool // Enable strict mode
	EnableMetrics bool // Enable metrics
	EnableLogging bool // Enable logging

	// Capabilities
	Capabilities HandlerCapabilities
}

// HandlerCapabilities represents handler capabilities
type HandlerCapabilities struct {
	SupportsRollback bool     `json:"supports_rollback"`
	SupportsRetry    bool     `json:"supports_retry"`
	SupportsAsync    bool     `json:"supports_async"`
	SupportsBatch    bool     `json:"supports_batch"`
	MaxConcurrency   int      `json:"max_concurrency"`
	MaxBatchSize     int      `json:"max_batch_size"`
	AllowedActions   []string `json:"allowed_actions"`
	AllowedModes     []string `json:"allowed_modes"`
}

// HandlerState represents handler state
type HandlerState map[string]any

// ToMap converts HandlerState to map[string]any
func (s HandlerState) ToMap() map[string]any {
	m := make(map[string]any)
	for k, v := range s {
		m[k] = v
	}
	return m
}

// ToJSON converts HandlerState to JSON
func (s HandlerState) ToJSON() []byte {
	data, _ := json.Marshal(s)
	return data
}

type TimerType string

const (
	TimerDelay TimerType = "delay" // delay execution
	TimerCron  TimerType = "cron"  // CRON expression
	TimerCycle TimerType = "cycle" // cycle execution
	TimerDate  TimerType = "date"  // designed date execution
)

// String implements fmt.Stringer
func (t TimerType) String() string {
	return string(t)
}

// TimeoutAction represents a timeout action
type TimeoutAction string

const (
	TimeoutActionNone     TimeoutAction = "none"
	TimeoutActionPass     TimeoutAction = "pass"
	TimeoutActionReject   TimeoutAction = "reject"
	TimeoutActionEscalate TimeoutAction = "escalate"
)

// String implements fmt.Stringer
func (t TimeoutAction) String() string {
	return string(t)
}

// ApprovalStrategy represents approval strategy
type ApprovalStrategy string

const (
	ApprovalAny      ApprovalStrategy = "any"      // any approval
	ApprovalAll      ApprovalStrategy = "all"      // all approval
	ApprovalMajority ApprovalStrategy = "majority" // majority approval
	ApprovalOrder    ApprovalStrategy = "order"    // approval order
)

// String implements fmt.Stringer
func (s ApprovalStrategy) String() string {
	return string(s)
}
