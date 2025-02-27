package types

// GeneralStatus defines common statuses across all components.
type GeneralStatus string

const (
	// General common statuses

	StatusPending    GeneralStatus = "pending"    // Waiting to be processed.
	StatusRunning    GeneralStatus = "running"    // Actively processing.
	StatusPaused     GeneralStatus = "paused"     // Temporarily halted.
	StatusCompleted  GeneralStatus = "completed"  // Successfully finished.
	StatusError      GeneralStatus = "error"      // An error occurred.
	StatusCancelled  GeneralStatus = "cancelled"  // Cancelled before completion.
	StatusStopped    GeneralStatus = "stopped"    // Stopped manually.
	StatusTerminated GeneralStatus = "terminated" // Forcefully terminated.
	StatusTimeout    GeneralStatus = "timeout"    // Timed out.
)

// String implements fmt.Stringer
func (s GeneralStatus) String() string {
	return string(s)
}

// EngineStatus represents the operational status of the workflow engine.
type EngineStatus string

const (
	EngineInitializing EngineStatus = "initializing" // Engine is starting up.
	EngineReady        EngineStatus = "ready"        // Engine is initialized and ready.
	EngineRunning      EngineStatus = "running"      // Engine is actively processing tasks.
	EnginePaused       EngineStatus = "paused"       // Engine is temporarily paused.
	EngineStopped      EngineStatus = "stopped"      // Engine is stopped and not running.
	EngineError        EngineStatus = "error"        // Engine encountered a critical error.
)

// String implements fmt.Stringer
func (s EngineStatus) String() string {
	return string(s)
}

// ExecutionStatus represents the status of a workflow instance execution.
type ExecutionStatus string

const (
	ExecutionPending    ExecutionStatus = "pending"    // Waiting to start execution.
	ExecutionActive     ExecutionStatus = "active"     // Actively executing.
	ExecutionCompleted  ExecutionStatus = "completed"  // Execution finished successfully.
	ExecutionError      ExecutionStatus = "error"      // Error occurred during execution.
	ExecutionCancelled  ExecutionStatus = "cancelled"  // Cancelled by user or system.
	ExecutionStopped    ExecutionStatus = "stopped"    // Execution stopped manually.
	ExecutionSuspended  ExecutionStatus = "suspended"  // Temporarily halted, can be resumed.
	ExecutionTerminated ExecutionStatus = "terminated" // Terminated and cannot be resumed.
	ExecutionWithdrawn  ExecutionStatus = "withdrawn"  // Withdrawn by the initiator.
	ExecutionWaiting    ExecutionStatus = "waiting"    // Waiting for external events or conditions.
	ExecutionEscalated  ExecutionStatus = "escalated"  // Escalated for higher-level attention.
	ExecutionRejected   ExecutionStatus = "rejected"   // Rejected by the responsible party.
	ExecutionRollbacked ExecutionStatus = "rollbacked" // Rolled back to a previous state.
	ExecutionTimeout    ExecutionStatus = "timeout"    // Execution timed out.
)

// String implements fmt.Stringer
func (s ExecutionStatus) String() string {
	return string(s)
}

// HandlerStatus represents the status of a workflow handler or task processor.
type HandlerStatus string

const (
	HandlerReady   HandlerStatus = "ready"   // Ready to process tasks.
	HandlerRunning HandlerStatus = "running" // Actively processing tasks.
	HandlerPaused  HandlerStatus = "paused"  // Temporarily paused.
	HandlerStopped HandlerStatus = "stopped" // Stopped and not processing tasks.
	HandlerError   HandlerStatus = "error"   // Encountered an error during task processing.
	HandlerTimeout HandlerStatus = "timeout" // Task processing timed out.
	HandlerBlocked HandlerStatus = "blocked" // Blocked due to dependency or other issues.
)

// String implements fmt.Stringer
func (s HandlerStatus) String() string {
	return string(s)
}

// NotificationStatus represents the status of notifications within the workflow.
type NotificationStatus string

const (
	NotificationPending   NotificationStatus = "pending"   // Notification is queued and waiting.
	NotificationSending   NotificationStatus = "sending"   // Notification is in the process of being sent.
	NotificationSent      NotificationStatus = "sent"      // Notification has been successfully sent.
	NotificationRead      NotificationStatus = "read"      // Notification has been read by the recipient.
	NotificationFailed    NotificationStatus = "failed"    // Sending notification failed.
	NotificationTimeout   NotificationStatus = "timeout"   // Notification delivery timed out.
	NotificationDismissed NotificationStatus = "dismissed" // Notification was dismissed by the recipient.
)

// String implements fmt.Stringer
func (s NotificationStatus) String() string {
	return string(s)
}

// ApprovalStatus represents the status of an approval process.
type ApprovalStatus string

const (
	ApprovalPending   ApprovalStatus = "pending"   // Awaiting approval decision.
	ApprovalApproved  ApprovalStatus = "approved"  // Approved by the responsible party.
	ApprovalRejected  ApprovalStatus = "rejected"  // Rejected by the responsible party.
	ApprovalEscalated ApprovalStatus = "escalated" // Escalated to higher-level approval.
	ApprovalWithdrawn ApprovalStatus = "withdrawn" // Withdrawn by the request initiator.
	ApprovalTimeout   ApprovalStatus = "timeout"   // Approval timed out.
)

// String implements fmt.Stringer
func (s ApprovalStatus) String() string {
	return string(s)
}

// TaskStatus represents the status of individual workflow tasks.
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"     // Task is waiting to be started.
	TaskInProgress TaskStatus = "in_progress" // Task is currently being worked on.
	TaskCompleted  TaskStatus = "completed"   // Task completed successfully.
	TaskFailed     TaskStatus = "failed"      // Task failed.
	TaskCancelled  TaskStatus = "cancelled"   // Task was cancelled.
	TaskSkipped    TaskStatus = "skipped"     // Task was skipped.
)

// String implements fmt.Stringer
func (s TaskStatus) String() string {
	return string(s)
}
