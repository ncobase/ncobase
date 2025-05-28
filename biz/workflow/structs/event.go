package structs

import "github.com/ncobase/ncore/types"

// WorkflowEvent represents workflow event constants
const (
	EventProcessError      = "workflow.process.error"
	EventProcessStarted    = "workflow.process.started"
	EventProcessTerminated = "workflow.process.terminated"
	EventProcessCompleted  = "workflow.process.completed"
	EventProcessRejected   = "workflow.process.rejected"
	EventProcessSuspended  = "workflow.process.suspended"
	EventProcessWithdraw   = "workflow.process.withdraw"
	EventProcessResumed    = "workflow.process.resumed"
	EventProcessCancelled  = "workflow.process.cancelled"
	EventNodeError         = "workflow.node.error"
	EventNodeStarted       = "workflow.node.started"
	EventNodeTimeout       = "workflow.node.timeout"
	EventNodeCompleted     = "workflow.node.completed"
	EventNodeCancelled     = "workflow.node.cancelled"
	EventTaskError         = "workflow.task.error"
	EventTaskCreated       = "workflow.task.created"
	EventTaskCompleted     = "workflow.task.completed"
	EventTaskDelegated     = "workflow.task.delegated"
	EventTaskTransferred   = "workflow.task.transferred"
	EventTaskUrged         = "workflow.task.urged"
	EventTaskTimeout       = "workflow.task.timeout"
	EventTaskWithdrawn     = "workflow.task.withdrawn"
	EventTaskAssigned      = "workflow.task.assigned"
	EventTaskCancelled     = "workflow.task.cancelled"
	EventTaskOverdue       = "workflow.task.overdue"
)

// EventData represents workflow event data
type EventData struct {
	Type          string     `json:"type"`
	ProcessID     string     `json:"process_id,omitempty"`
	ProcessName   string     `json:"process_name,omitempty"`
	NodeID        string     `json:"node_id,omitempty"`
	NodeName      string     `json:"node_name,omitempty"`
	TaskID        string     `json:"task_id,omitempty"`
	TaskName      string     `json:"task_name,omitempty"`
	Operator      string     `json:"operator,omitempty"`
	Action        ActionType `json:"action,omitempty"`
	Variables     types.JSON `json:"variables,omitempty"`
	BusinessData  types.JSON `json:"business_data,omitempty"`
	Timestamp     int64      `json:"timestamp"`
	ModuleCode    string     `json:"module_code,omitempty"`
	FormCode      string     `json:"form_code,omitempty"`
	TemplateID    string     `json:"template_id,omitempty"`
	TemplateName  string     `json:"template_name,omitempty"`
	ProcessStatus Status     `json:"process_status,omitempty"`
	NodeType      NodeType   `json:"node_type,omitempty"`
	PrevNodeID    string     `json:"prev_node_id,omitempty"`
	NextNodeID    string     `json:"next_node_id,omitempty"`
	Duration      int64      `json:"duration,omitempty"`
	ErrorInfo     string     `json:"error_info,omitempty"`
	Assignees     []string   `json:"assignees,omitempty"`
	Comment       string     `json:"comment,omitempty"`
	Details       types.JSON `json:"details,omitempty"`
}
