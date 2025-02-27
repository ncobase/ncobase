package config

import (
	"context"
	"ncobase/core/workflow/engine/types"
	"ncobase/core/workflow/structs"
	"time"
)

// HandlerConfig represents configuration for all handlers
type HandlerConfig struct {
	// Base handler config
	Base *BaseHandlerConfig `json:"base,omitempty"`
	// Approval handler config
	Approval *ApprovalHandlerConfig `json:"approval,omitempty"`
	// Exclusive handler config
	Exclusive *ExclusiveHandlerConfig `json:"exclusion,omitempty"`
	// Notification handler config
	Notification *NotificationHandlerConfig `json:"notification,omitempty"`
	// Parallel handler config
	Parallel *ParallelHandlerConfig `json:"parallel,omitempty"`
	// SubProcess handler config
	SubProcess *SubProcessHandlerConfig `json:"sub_process,omitempty"`
	// Service handler config
	Service *ServiceHandlerConfig `json:"service,omitempty"`
	// Timer handler config
	Timer *TimerHandlerConfig `json:"timer,omitempty"`
	// Script handler config
	Script *ScriptHandlerConfig `json:"script,omitempty"`
}

// Base handler config

// BaseHandlerConfig represents base handler configuration
type BaseHandlerConfig struct {
	// Basic execution settings
	MaxWorkers    int           `json:"max_workers"`
	Timeout       time.Duration `json:"timeout"`
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"`

	// Concurrency control
	MaxBatchSize  int32 `json:"max_batch_size"`
	MaxConcurrent int32 `json:"max_concurrent"`
	QueueSize     int32 `json:"queue_size"`
	Workers       int32 `json:"workers"`

	// Features flags
	StrictMode     bool `json:"strict_mode"`
	EnableMetrics  bool `json:"enable_metrics"`
	EnableAsync    bool `json:"enable_async"`
	EnableLogging  bool `json:"enable_logging"`
	EnableRollback bool `json:"enable_rollback"`

	// Resource limits
	MaxMemory int64 `json:"max_memory"`
	MaxCPU    int   `json:"max_cpu"`

	// Validation
	ValidateInput bool `json:"validate_input"`

	// Error handling
	ErrorMode     string `json:"error_mode"`     // continue/fail
	FailurePolicy string `json:"failure_policy"` // ignore/retry/rollback

	// Hooks
	Hooks *HandlerHooks `json:"hooks"`

	// Executor capabilities
	Capabilities *types.HandlerCapabilities `json:"capabilities"`
}

// HandlerHooks defines handler lifecycle hooks
type HandlerHooks struct {
	BeforeExecute  func(ctx context.Context, node *structs.ReadNode) error
	AfterExecute   func(ctx context.Context, node *structs.ReadNode, err error)
	BeforeComplete func(ctx context.Context, node *structs.ReadNode) error
	AfterComplete  func(ctx context.Context, node *structs.ReadNode, err error)
	OnError        func(ctx context.Context, node *structs.ReadNode, err error)
}

// DefaultBaseHandlerConfig returns default handler configuration
func DefaultBaseHandlerConfig() *BaseHandlerConfig {
	return &BaseHandlerConfig{
		MaxWorkers:    10,
		QueueSize:     1000,
		Timeout:       time.Minute * 5,
		MaxRetries:    3,
		RetryInterval: time.Second * 5,
		StrictMode:    false,
		EnableMetrics: true,
		EnableAsync:   false,
		EnableLogging: true,
		MaxMemory:     512 * 1024 * 1024, // 512MB
		MaxCPU:        80,                // 80% CPU usage
		ValidateInput: true,
		ErrorMode:     "fail",
		FailurePolicy: "retry",
		Hooks:         &HandlerHooks{},
	}
}

// Node type specific configurations

// ApprovalHandlerConfig represents approval node configuration
type ApprovalHandlerConfig struct {
	// Approval strategy
	Strategy     string   `json:"strategy"`   // any/all/majority
	Candidates   []string `json:"candidates"` // Static approvers
	TimeoutHours int      `json:"timeout_hours"`

	// Task features
	AllowDelegate bool `json:"allow_delegate"`
	AllowTransfer bool `json:"allow_transfer"`
	AllowUrge     bool `json:"allow_urge"`

	// Timeout behavior
	AutoPass   bool `json:"auto_pass"`
	AutoReject bool `json:"auto_reject"`

	// Urge settings
	MaxUrges     int           `json:"max_urges"`
	UrgeInterval time.Duration `json:"urge_interval"`
	AutoEscalate bool          `json:"auto_escalate"`

	// Delegate rules
	DelegateRules *DelegateRules `json:"delegate_rules"`
}

type DelegateRules struct {
	AllowedRoles []string      `json:"allowed_roles"`
	MaxDelegates int           `json:"max_delegates"`
	Duration     time.Duration `json:"duration"`
}

// DefaultApprovalHandlerConfig returns default approval node configuration
func DefaultApprovalHandlerConfig() *ApprovalHandlerConfig {
	return &ApprovalHandlerConfig{
		Strategy:      "any",
		TimeoutHours:  24,
		AllowDelegate: true,
		AllowTransfer: true,
		AllowUrge:     true,
		AutoPass:      false,
		AutoReject:    false,
		MaxUrges:      3,
		UrgeInterval:  time.Hour * 24,
		AutoEscalate:  false,
		DelegateRules: &DelegateRules{},
	}
}

// Exclusive handler

// ExclusiveHandlerConfig represents exclusive gateway configuration
type ExclusiveHandlerConfig struct {
	Conditions  []Condition `json:"conditions"`   // Prioritized conditions
	DefaultPath string      `json:"default_path"` // Default path if no condition matches
	FailureMode string      `json:"failure_mode"` // continue/error
}

// Condition represents a gateway condition
type Condition struct {
	Expression string `json:"expression"` // Condition expression
	NextNode   string `json:"next_node"`  // Next node if condition matches
	Priority   int    `json:"priority"`   // Evaluation priority
}

// DefaultExclusiveHandlerConfig returns default exclusive gateway configuration
func DefaultExclusiveHandlerConfig() *ExclusiveHandlerConfig {
	return &ExclusiveHandlerConfig{
		Conditions:  []Condition{},
		DefaultPath: "",
		FailureMode: "error",
	}
}

// Notification handler

// NotificationHandlerConfig represents notification configuration
type NotificationHandlerConfig struct {
	Type       string         `json:"type"`       // email/sms/webhook etc
	Template   string         `json:"template"`   // Template ID/name
	Recipients []string       `json:"recipients"` // Static recipients
	Variables  map[string]any `json:"variables"`  // Template variables
	Options    map[string]any `json:"options"`    // Additional options
}

// DefaultNotificationHandlerConfig returns default notification configuration
func DefaultNotificationHandlerConfig() *NotificationHandlerConfig {
	return &NotificationHandlerConfig{
		Type:       "email",
		Template:   "",
		Recipients: []string{},
		Variables:  map[string]any{},
		Options:    map[string]any{},
	}
}

// Parallel handler

// ParallelHandlerConfig represents parallel gateway configuration
type ParallelHandlerConfig struct {
	Branches      []ParallelHandlerBranch `json:"branches"`       // Parallel branches
	CompleteMode  string                  `json:"complete_mode"`  // all/any/majority
	ErrorMode     string                  `json:"error_mode"`     // continue/fail
	Timeout       int                     `json:"timeout"`        // in seconds
	MaxConcurrent int                     `json:"max_concurrent"` // max concurrent branches
}

// ParallelHandlerBranch represents a parallel branch
type ParallelHandlerBranch struct {
	NodeKey   string         `json:"node_key"`  // Branch node key
	Priority  int            `json:"priority"`  // Execution priority
	Required  bool           `json:"required"`  // Is branch required
	Condition string         `json:"condition"` // Optional condition
	Variables map[string]any `json:"variables"` // Branch specific variables
}

// DefaultParallelHandlerConfig returns default parallel gateway configuration
func DefaultParallelHandlerConfig() *ParallelHandlerConfig {
	return &ParallelHandlerConfig{
		Branches:      []ParallelHandlerBranch{},
		CompleteMode:  "all",
		ErrorMode:     "fail",
		Timeout:       0,
		MaxConcurrent: 10,
	}
}

// SubProcess handler

// SubProcessHandlerConfig represents subprocess configuration
type SubProcessHandlerConfig struct {
	TemplateID   string         `json:"template_id"`   // Subprocess template ID
	ProcessCode  string         `json:"process_code"`  // Process code
	Variables    map[string]any `json:"variables"`     // Initial variables
	WaitComplete bool           `json:"wait_complete"` // Wait for completion
	Timeout      int            `json:"timeout"`       // Timeout in seconds
}

// DefaultSubProcessHandlerConfig returns default subprocess configuration
func DefaultSubProcessHandlerConfig() *SubProcessHandlerConfig {
	return &SubProcessHandlerConfig{
		TemplateID:   "",
		ProcessCode:  "",
		Variables:    map[string]any{},
		WaitComplete: true,
		Timeout:      3600,
	}
}

// Service handler

// ServiceHandlerConfig represents service node configuration
type ServiceHandlerConfig struct {
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Endpoint   string            `json:"endpoint"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	InputVars  map[string]any    `json:"input_vars"`
	OutputVars map[string]any    `json:"output_vars"`
	Timeout    time.Duration     `json:"timeout"`

	RetryPolicy *RetryConfig `json:"retry_policy"`

	FailureMode string `json:"failure_mode"`
	Async       bool   `json:"async"`

	// Circuit breaker
	CBMaxRequests int           `json:"cb_max_requests"`
	CBInterval    time.Duration `json:"cb_interval"`
	CBTimeout     time.Duration `json:"cb_timeout"`

	// Cache
	EnableCache  bool          `json:"enable_cache"`
	CacheTTL     time.Duration `json:"cache_ttl"`
	MaxCacheSize int           `json:"max_cache_size"`
}

// DefaultServiceHandlerConfig returns default service node configuration
func DefaultServiceHandlerConfig() *ServiceHandlerConfig {
	return &ServiceHandlerConfig{
		Type:          "http",
		Name:          "",
		Endpoint:      "",
		Method:        "GET",
		Headers:       map[string]string{},
		InputVars:     map[string]any{},
		OutputVars:    map[string]any{},
		Timeout:       time.Second * 10,
		RetryPolicy:   DefaultRetryConfig(),
		FailureMode:   "error",
		Async:         false,
		CBMaxRequests: 100,
		CBInterval:    time.Second * 30,
		CBTimeout:     time.Second * 60,
		EnableCache:   false,
		CacheTTL:      time.Minute,
		MaxCacheSize:  1000,
	}
}

// Timer handler

// TimerHandlerConfig represents timer node configuration
type TimerHandlerConfig struct {
	Type          string        `json:"type"`           // delay/cron/cycle/date
	Duration      string        `json:"duration"`       // For delay type
	CronExpr      string        `json:"cron"`           // For cron type
	CycleCount    int           `json:"cycle_count"`    // For cycle type
	CycleInterval time.Duration `json:"cycle_interval"` // For cycle type
	TriggerDate   string        `json:"date"`           // For date type
	TimeoutHours  int           `json:"timeout_hours"`
	IsWorkingDay  bool          `json:"is_working_day"`
	FailureMode   string        `json:"failure_mode"`
}

// DefaultTimerHandlerConfig returns default timer node configuration
func DefaultTimerHandlerConfig() *TimerHandlerConfig {
	return &TimerHandlerConfig{
		Type:         string(types.TimerDelay),
		Duration:     "1h",
		TimeoutHours: 24,
		IsWorkingDay: true,
		FailureMode:  "fail",
	}
}

// Script handler

// ScriptHandlerConfig represents script node configuration
type ScriptHandlerConfig struct {
	Script      string         `json:"script"`
	Language    string         `json:"language"`
	Timeout     time.Duration  `json:"timeout"`
	MaxMemory   int64          `json:"max_memory"`
	AllowedAPIs []string       `json:"allowed_apis"`
	InputVars   map[string]any `json:"input_vars"`
	OutputVars  []string       `json:"output_vars"`
	FailureMode string         `json:"failure_mode"`
	Sandbox     *Sandbox       `json:"sandbox"`
}

// DefaultScriptHandlerConfig returns default script node configuration
func DefaultScriptHandlerConfig() *ScriptHandlerConfig {
	return &ScriptHandlerConfig{
		Script:      "",
		Language:    "javascript",
		Timeout:     time.Second * 30,
		MaxMemory:   50 * 1024 * 1024,
		AllowedAPIs: []string{"math", "date"},
		InputVars:   map[string]any{},
		OutputVars:  []string{},
		FailureMode: "fail",
		Sandbox:     DefaultSandbox(),
	}
}

// Sandbox represents script execution sandbox
type Sandbox struct {
	// Resource limits
	MaxMemory       int64         // Maximum memory usage in bytes
	MaxCPU          int           // Maximum CPU usage percentage
	Timeout         time.Duration // Maximum execution time
	MaxInstructions int64         // Maximum instructions
	MaxEventLoopLag time.Duration // Maximum event loop lag

	// Security settings
	AllowedAPIs     []string                  // Allowed API packages
	BlockedPackages []string                  // Blocked packages/imports
	IsStrict        bool                      // Strict mode enabled
	Globals         map[string]any            // Global variables
	Modules         map[string]map[string]any // Built-in modules
	Console         map[string]any            // Custom console methods
}

// DefaultSandbox returns the default sandbox configuration
func DefaultSandbox() *Sandbox {
	return &Sandbox{
		MaxMemory:       50 * 1024 * 1024, // 50MB
		MaxCPU:          10,               // 10% CPU
		Timeout:         time.Second * 30, // 30s timeout
		MaxInstructions: 1000000,          // 1M instructions
		MaxEventLoopLag: time.Millisecond * 100,
		AllowedAPIs:     []string{"math", "date", "string"},
		BlockedPackages: []string{"os", "net", "syscall"},
		IsStrict:        true,
	}
}
