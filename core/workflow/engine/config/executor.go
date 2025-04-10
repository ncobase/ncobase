package config

import (
	"context"
	"github.com/ncobase/ncore/pkg/queue"
	"ncobase/core/workflow/engine/scheduler"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/structs"
	"time"
)

// ExecutorConfig represents configuration for all executors
type ExecutorConfig struct {
	// Base executor config
	Base *BaseExecutorConfig `json:"base,omitempty"`

	// Process executor config
	Process *ProcessExecutorConfig `json:"process,omitempty"`

	// Node executor config
	Node *NodeExecutorConfig `json:"node,omitempty"`

	// Task executor config
	Task *TaskExecutorConfig `json:"task,omitempty"`

	// Service executor config
	Service *ServiceExecutorConfig `json:"service,omitempty"`

	// Retry executor config
	Retry *RetryExecutorConfig `json:"retry,omitempty"`
}

// Base executor

// BaseExecutorConfig provides common configuration for all executors
type BaseExecutorConfig struct {
	// Basic settings
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`
	Timeout       time.Duration `json:"timeout"`

	// Concurrency control
	MaxBatchSize  int32 `json:"max_batch_size"`
	MaxConcurrent int32 `json:"max_concurrent"`
	QueueSize     int32 `json:"queue_size"`
	Workers       int32 `json:"workers"`

	// Features
	EnableMetrics bool   `json:"enable_metrics"`
	EnableAsync   bool   `json:"enable_async"`
	StrictMode    bool   `json:"strict_mode"`
	LogLevel      string `json:"log_level"`

	// Executor capabilities
	Capabilities *types.ExecutionCapabilities `json:"capabilities"`
}

// DefaultBaseExecutorConfig returns default base executor configuration
func DefaultBaseExecutorConfig() *BaseExecutorConfig {
	return &BaseExecutorConfig{
		MaxRetries:    3,
		RetryInterval: time.Second * 5,
		Timeout:       time.Minute * 5,
		MaxConcurrent: 100,
		QueueSize:     1000,
		Workers:       10,
		EnableMetrics: true,
		EnableAsync:   false,
		StrictMode:    false,
		LogLevel:      "info",
		Capabilities: &types.ExecutionCapabilities{
			SupportsAsync:    true,
			SupportsRetry:    true,
			SupportsRollback: false,
			MaxConcurrency:   100,
			MaxBatchSize:     1000,
			AllowedActions:   []string{"execute", "cancel", "rollback"},
		},
	}
}

// Process executor

// ProcessExecutorConfig represents process executor configuration
type ProcessExecutorConfig struct {
	*BaseExecutorConfig

	// Process execution settings
	AutoStartSubProcess bool          `json:"auto_start_subprocess"`
	WaitSubProcess      bool          `json:"wait_subprocess"`
	RollbackOnError     bool          `json:"rollback_on_error"`
	ProcessTimeout      time.Duration `json:"process_timeout"`
	SubProcessTimeout   time.Duration `json:"subprocess_timeout"`
	NodeTimeout         time.Duration `json:"node_timeout"`

	// Recovery settings
	EnableRecovery   bool          `json:"enable_recovery"`
	EnableSnapshot   bool          `json:"enable_snapshot"`
	SnapshotInterval time.Duration `json:"snapshot_interval"`

	// Hooks
	Hooks *ProcessHooks `json:"hooks"`
}

// ProcessHooks defines process execution hooks
type ProcessHooks struct {
	// Called before process execution
	BeforeExecute func(ctx context.Context, process *structs.ReadProcess) error

	// Called after process execution
	AfterExecute func(ctx context.Context, process *structs.ReadProcess, err error)

	// Called on process error
	OnError func(ctx context.Context, process *structs.ReadProcess, err error)
}

// DefaultProcessExecutorConfig returns default process executor configuration
func DefaultProcessExecutorConfig() *ProcessExecutorConfig {
	return &ProcessExecutorConfig{
		BaseExecutorConfig:  DefaultBaseExecutorConfig(),
		AutoStartSubProcess: true,
		WaitSubProcess:      true,
		RollbackOnError:     true,
		ProcessTimeout:      time.Hour,
		SubProcessTimeout:   time.Hour,
		NodeTimeout:         time.Minute * 30,
		EnableRecovery:      true,
		EnableSnapshot:      true,
		SnapshotInterval:    time.Minute * 5,
	}
}

// Node executor

// NodeExecutorConfig represents node executor configuration
type NodeExecutorConfig struct {
	*BaseExecutorConfig

	// Node execution settings
	ConcurrentNodes int  `json:"concurrent_nodes"`
	BufferSize      int  `json:"buffer_size"`
	AutoComplete    bool `json:"auto_complete"`
	ValidateConfig  bool `json:"validate_config"`
	TrackJoinStatus bool `json:"track_join_status"`

	// Hooks
	Hooks *NodeHooks `json:"hooks"`
}

// NodeHooks defines node execution hooks
type NodeHooks struct {
	BeforeExecute  func(ctx context.Context, node *structs.ReadNode) error
	AfterExecute   func(ctx context.Context, node *structs.ReadNode, err error)
	OnError        func(ctx context.Context, node *structs.ReadNode, err error)
	BeforeComplete func(ctx context.Context, node *structs.ReadNode) error
	AfterComplete  func(ctx context.Context, node *structs.ReadNode, err error)
	BeforeRollback func(ctx context.Context, node *structs.ReadNode) error
	AfterRollback  func(ctx context.Context, node *structs.ReadNode, err error)
}

// DefaultNodeExecutorConfig returns default node executor configuration
func DefaultNodeExecutorConfig() *NodeExecutorConfig {
	return &NodeExecutorConfig{
		BaseExecutorConfig: DefaultBaseExecutorConfig(),
		ConcurrentNodes:    50,
		BufferSize:         1000,
		AutoComplete:       true,
		ValidateConfig:     true,
		TrackJoinStatus:    true,
	}
}

// Task executor

// TaskExecutorConfig represents task executor configuration
type TaskExecutorConfig struct {
	*BaseExecutorConfig

	// Task assignment settings
	EnableAutoAssign  bool          `json:"enable_auto_assign"`
	MaxAssignRetries  int           `json:"max_assign_retries"`
	AssignmentTimeout time.Duration `json:"assignment_timeout"`
	LoadBalanceMode   string        `json:"load_balance_mode"`

	// Task operation permissions
	AllowDelegate bool `json:"allow_delegate"`
	AllowTransfer bool `json:"allow_transfer"`
	AllowWithdraw bool `json:"allow_withdraw"`
	AllowUrge     bool `json:"allow_urge"`
	AllowReassign bool `json:"allow_reassign"`

	// Task timing settings
	DefaultDueTime   time.Duration `json:"default_due_time"`
	ReminderInterval time.Duration `json:"reminder_interval"`
	OverdueReminder  bool          `json:"overdue_reminder"`

	// Advanced features
	EnablePriority  bool `json:"enable_priority"`
	EnableDeadline  bool `json:"enable_deadline"`
	EnableWorkload  bool `json:"enable_workload"`
	BatchAssignment bool `json:"batch_assignment"`

	// Hooks
	Hooks *TaskHooks `json:"hooks"`
}

// TaskHooks defines task execution hooks
type TaskHooks struct {
	BeforeCreate   func(ctx context.Context, task *structs.TaskBody) error
	AfterCreate    func(ctx context.Context, task *structs.ReadTask) error
	BeforeComplete func(ctx context.Context, req *structs.CompleteTaskRequest) error
	AfterComplete  func(ctx context.Context, resp *structs.CompleteTaskResponse) error
	BeforeAssign   func(ctx context.Context, task *structs.ReadTask, assignees []string) error
	AfterAssign    func(ctx context.Context, task *structs.ReadTask, err error)
	OnTimeout      func(ctx context.Context, task *structs.ReadTask)
	OnOverdue      func(ctx context.Context, task *structs.ReadTask)
}

// DefaultTaskExecutorConfig returns default task executor configuration
func DefaultTaskExecutorConfig() *TaskExecutorConfig {
	return &TaskExecutorConfig{
		BaseExecutorConfig: DefaultBaseExecutorConfig(),
		EnableAutoAssign:   true,
		MaxAssignRetries:   3,
		AssignmentTimeout:  time.Minute * 5,
		LoadBalanceMode:    "round-robin",
		AllowDelegate:      true,
		AllowTransfer:      true,
		AllowWithdraw:      true,
		AllowUrge:          true,
		DefaultDueTime:     time.Hour * 24,
		ReminderInterval:   time.Hour,
		EnablePriority:     true,
		EnableDeadline:     true,
		EnableWorkload:     true,
		BatchAssignment:    false,
	}
}

// Service executor

// ServiceExecutorConfig represents service executor configuration
type ServiceExecutorConfig struct {
	*BaseExecutorConfig

	// Service discovery
	EnableDiscovery bool `json:"enable_discovery"`

	// Circuit breaker settings
	CBMaxRequests int           `json:"cb_max_requests"`
	CBInterval    time.Duration `json:"cb_interval"`
	CBTimeout     time.Duration `json:"cb_timeout"`

	// Cache settings
	EnableCache  bool          `json:"enable_cache"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	MaxCacheSize int           `json:"max_cache_size"`
}

// DefaultServiceExecutorConfig returns default service executor configuration
func DefaultServiceExecutorConfig() *ServiceExecutorConfig {
	return &ServiceExecutorConfig{
		BaseExecutorConfig: DefaultBaseExecutorConfig(),
		EnableDiscovery:    false,
		CBMaxRequests:      100,
		CBInterval:         time.Second * 30,
		CBTimeout:          time.Second * 60,
		EnableCache:        false,
		CacheTTL:           time.Minute,
		MaxCacheSize:       1000,
	}
}

// Retry executor

// RetryExecutorConfig represents retry executor configuration
type RetryExecutorConfig struct {
	Policy         *RetryConfig         `json:"policy"`
	Queue          *queue.TaskQueue     `json:"queue"`
	Scheduler      *scheduler.Scheduler `json:"scheduler"`
	MetricsEnabled bool                 `json:"metrics_enabled"`
	LogLevel       string               `json:"log_level"`
}

// DefaultRetryExecutorConfig returns default retry executor config
func DefaultRetryExecutorConfig() *RetryExecutorConfig {
	return &RetryExecutorConfig{
		Policy:         DefaultRetryConfig(),
		Queue:          queue.NewTaskQueue(context.Background()),
		Scheduler:      scheduler.NewScheduler(nil, nil, nil),
		MetricsEnabled: true,
		LogLevel:       "info",
	}
}
