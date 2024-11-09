package structs

import "ncobase/common/types"

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

// ProcessPriority represents process priority levels
type ProcessPriority int

const (
	PriorityLow    ProcessPriority = 0
	PriorityNormal ProcessPriority = 1
	PriorityHigh   ProcessPriority = 2
	PriorityUrgent ProcessPriority = 3
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
	StatusDraft      Status = "draft"      // Draft
	StatusActive     Status = "active"     // Active
	StatusPending    Status = "pending"    // Pending
	StatusProcessing Status = "processing" // Processing
	StatusCompleted  Status = "completed"  // Completed
	StatusRejected   Status = "rejected"   // Rejected
	StatusCancelled  Status = "cancelled"  // Cancelled
	StatusTerminated Status = "terminated" // Terminated
	StatusError      Status = "error"      // Error
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
	SortByCreatedAt types.SortField = "created_at"
	SortByPriority  types.SortField = "priority"
	SortByDueTime   types.SortField = "due_time"
	SortByName      types.SortField = "name"
)
