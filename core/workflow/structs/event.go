package structs

import "ncobase/common/types"

// WorkflowEvent represents workflow event constants
const (
	EventProcessStarted   = "workflow.process.started"
	EventProcessCompleted = "workflow.process.completed"
	EventProcessRejected  = "workflow.process.rejected"
	EventProcessCancelled = "workflow.process.cancelled"
	EventNodeStarted      = "workflow.node.started"
	EventNodeCompleted    = "workflow.node.completed"
	EventTaskCreated      = "workflow.task.created"
	EventTaskCompleted    = "workflow.task.completed"
	EventTaskDelegated    = "workflow.task.delegated"
	EventTaskTransferred  = "workflow.task.transferred"
	EventTaskUrged        = "workflow.task.urged"
	EventTaskOverdue      = "workflow.task.overdue"
)

// EventData represents workflow event data
type EventData struct {
	ProcessID     string      `json:"process_id,omitempty"`
	ProcessName   string      `json:"process_name,omitempty"`
	NodeID        string      `json:"node_id,omitempty"`
	NodeName      string      `json:"node_name,omitempty"`
	TaskID        string      `json:"task_id,omitempty"`
	TaskName      string      `json:"task_name,omitempty"`
	Operator      string      `json:"operator,omitempty"`
	Action        ActionType  `json:"action,omitempty"`
	Variables     *types.JSON `json:"variables,omitempty"`
	BusinessData  *types.JSON `json:"business_data,omitempty"`
	Timestamp     int64       `json:"timestamp"`
	ModuleCode    string      `json:"module_code,omitempty"`
	FormCode      string      `json:"form_code,omitempty"`
	TemplateID    string      `json:"template_id,omitempty"`
	TemplateName  string      `json:"template_name,omitempty"`
	ProcessStatus Status      `json:"process_status,omitempty"`
	NodeType      NodeType    `json:"node_type,omitempty"`
	PrevNodeID    string      `json:"prev_node_id,omitempty"`
	NextNodeID    string      `json:"next_node_id,omitempty"`
	Duration      int64       `json:"duration,omitempty"`
	ErrorInfo     string      `json:"error_info,omitempty"`
}
