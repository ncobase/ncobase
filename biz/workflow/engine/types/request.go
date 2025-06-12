package types

// StartProcessRequest represents process start request
type StartProcessRequest struct {
	TemplateID  string         // Template ID
	BusinessKey string         // Business key
	Variables   map[string]any // Initial variables
	Priority    int            // Process priority
	Initiator   string         // Process initiator
	SpaceID     string         // Space ID
}

// StartProcessResponse represents process start response
type StartProcessResponse struct {
	ProcessID string          // Process ID
	Status    ExecutionStatus // Process status
	StartTime *int64          // Start time
	Variables map[string]any  // Process variables
}

// TerminateRequest represents process terminate request
type TerminateRequest struct {
	ProcessID string // Process ID
	Operator  string // Operator
	Reason    string // Termination reason
	Comment   string // Comment
}

// CompleteRequest represents node/task completion request
type CompleteRequest struct {
	Action    string         // Complete action
	Variables map[string]any // Variables
	Comment   string         // Comment
	Operator  string         // Operator
}
