package structs

// Error codes
const (
	ErrCodeInvalidParam = "INVALID_PARAM"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeSystemError  = "SYSTEM_ERROR"
	ErrCodeTimeout      = "TIMEOUT"
	ErrCodeBizError     = "BIZ_ERROR"
)

// AssigneeType represents assignee types
type AssigneeType string

const (
	AssigneeUser  AssigneeType = "user"
	AssigneeRole  AssigneeType = "role"
	AssigneeDept  AssigneeType = "dept"
	AssigneeGroup AssigneeType = "group"
)

// Assignee represents assignee
type Assignee struct {
	ID   string       `json:"id"`
	Name string       `json:"name"`
	Type AssigneeType `json:"type"`
}

// AssignStrategy represents assign strategy
type AssignStrategy string

const (
	AssignStrategyAuto   = "auto"
	AssignStrategyClaim  = "claim"
	AssignStrategyManual = "manual"
)

// PriorityStrategy represents priority handling strategies
type PriorityStrategy int

const (
	PriorityLow    PriorityStrategy = 0
	PriorityNormal PriorityStrategy = 5
	PriorityHigh   PriorityStrategy = 10
	PriorityUrgent PriorityStrategy = 15
)

// CategoryType represents category types
type CategoryType string

const (
	CategoryProcess = "process"
	CategoryNode    = "node"
	CategoryTask    = "task"
	CategorySystem  = "system"
)

// TimeoutStrategy represents timeout handling strategies
type TimeoutStrategy string

const (
	TimeoutNone     TimeoutStrategy = "none"     // No timeout handling
	TimeoutAutoPass TimeoutStrategy = "autoPass" // Auto pass when timeout
	TimeoutAutoFail TimeoutStrategy = "autoFail" // Auto fail when timeout
	TimeoutAlert    TimeoutStrategy = "alert"    // Only alert when timeout
)

// CountersignStrategy represents countersign strategies
type CountersignStrategy string

const (
	CountersignAny      CountersignStrategy = "any"      // Any one approval passes
	CountersignAll      CountersignStrategy = "all"      // All approvals required
	CountersignMajority CountersignStrategy = "majority" // Majority approval passes
)

// Status represents workflow status
type Status string

const (
	StatusDraft       Status = "draft"       // Draft
	StatusReady       Status = "ready"       // Ready
	StatusActive      Status = "active"      // Active
	StatusPending     Status = "pending"     // Pending
	StatusSuspended   Status = "suspended"   // Suspended
	StatusProcessing  Status = "processing"  // Processing
	StatusCompleted   Status = "completed"   // Completed
	StatusCompensated Status = "compensated" // Compensated
	StatusRejected    Status = "rejected"    // Rejected
	StatusCancelled   Status = "cancelled"   // Cancelled
	StatusTerminated  Status = "terminated"  // Terminated
	StatusRollbacked  Status = "rollbacked"  // Rollbacked
	StatusWithdrawn   Status = "withdrawn"   // Withdrawn
	StatusError       Status = "error"       // Error
)

// NodeType represents workflow node types
type NodeType string

const (
	NodeStart     NodeType = "start"     // Start node
	NodeApproval  NodeType = "approval"  // Approval node
	NodeService   NodeType = "service"   // Service node
	NodeExclusive NodeType = "exclusive" // Exclusive node
	NodeParallel  NodeType = "parallel"  // Parallel node
	NodeCc        NodeType = "cc"        // CC node
	NodeEnd       NodeType = "end"       // End node
)

// ActionType represents workflow actions
type ActionType string

const (
	ActionSubmit    ActionType = "submit"    // Submit form
	ActionSave      ActionType = "save"      // Save draft
	ActionRevoke    ActionType = "revoke"    // Revoke approval
	ActionReassign  ActionType = "reassign"  // Reassign task
	ActionAddSign   ActionType = "addSign"   // Add countersign
	ActionRemind    ActionType = "remind"    // Send reminder
	ActionApprove   ActionType = "approve"   // Approve
	ActionReject    ActionType = "reject"    // Reject
	ActionDelegate  ActionType = "delegate"  // Delegate
	ActionTransfer  ActionType = "transfer"  // Transfer
	ActionWithdraw  ActionType = "withdraw"  // Withdraw
	ActionTerminate ActionType = "terminate" // Terminate
	ActionSuspend   ActionType = "suspend"   // Suspend
	ActionResume    ActionType = "resume"    // Resume
	ActionUrge      ActionType = "urge"      // Urge
)

// CompareOperator represents comparison operators
type CompareOperator string

const (
	CompareEq     CompareOperator = "eq"     // Equal
	CompareNe     CompareOperator = "ne"     // Not equal
	CompareGt     CompareOperator = "gt"     // Greater than
	CompareGe     CompareOperator = "ge"     // Greater than or equal
	CompareLt     CompareOperator = "lt"     // Less than
	CompareLe     CompareOperator = "le"     // Less than or equal
	CompareIn     CompareOperator = "in"     // In
	CompareNin    CompareOperator = "nin"    // Not in
	CompareExists CompareOperator = "exists" // Exists
	CompareNull   CompareOperator = "null"   // Null
)

// HandlerType represents handler types
type HandlerType string

const (
	HandlerBefore HandlerType = "before" // Before handler
	HandlerAfter  HandlerType = "after"  // After handler
	HandlerRule   HandlerType = "rule"   // Rule handler
	HandlerAction HandlerType = "action" // Action handler
)

// LogicOperator represents logical operators
type LogicOperator string

const (
	LogicAnd LogicOperator = "and" // And
	LogicOr  LogicOperator = "or"  // Or
)

// SortField represents sort fields
const (
	SortByCreatedAt string = "created_at"
	SortByPriority  string = "priority"
	SortByDueTime   string = "due_time"
	SortByName      string = "name"
)
