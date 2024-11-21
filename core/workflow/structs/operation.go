package structs

import (
	"ncobase/common/types"
)

// StartProcessRequest represents process start request
type StartProcessRequest struct {
	TemplateID    string     `json:"template_id" binding:"required"`
	BusinessKey   string     `json:"business_key" binding:"required"`
	ModuleCode    string     `json:"module_code,omitempty"`
	FormCode      string     `json:"form_code,omitempty"`
	Variables     types.JSON `json:"variables,omitempty"`
	FormData      types.JSON `json:"form_data,omitempty"`
	Priority      int        `json:"priority,omitempty"`
	Initiator     string     `json:"initiator" binding:"required"`
	InitiatorDept string     `json:"initiator_dept,omitempty"`
	TenantID      string     `json:"tenant_id,omitempty"`
}

// StartProcessResponse represents process start response
type StartProcessResponse struct {
	ProcessID string     `json:"process_id"`
	Status    Status     `json:"status"`
	StartTime *int64     `json:"start_time"`
	Variables types.JSON `json:"variables,omitempty"`
}

// CompleteTaskRequest represents task completion request
type CompleteTaskRequest struct {
	TaskID      string     `json:"task_id" binding:"required"`
	Action      ActionType `json:"action" binding:"required"`
	Comment     string     `json:"comment,omitempty"`
	Variables   types.JSON `json:"variables,omitempty"`
	FormData    types.JSON `json:"form_data,omitempty"`
	Attachments types.JSON `json:"attachments,omitempty"`
	Operator    string     `json:"operator" binding:"required"`
}

// CompleteTaskResponse represents task completion response
type CompleteTaskResponse struct {
	TaskID    string            `json:"task_id"`
	ProcessID string            `json:"process_id"`
	Action    ActionType        `json:"action"`
	EndTime   *int64            `json:"end_time"`
	NextNodes types.StringArray `json:"next_nodes,omitempty"`
}

// DelegateTaskRequest represents task delegation request
type DelegateTaskRequest struct {
	TaskID    string `json:"task_id" binding:"required"`
	Delegator string `json:"delegator" binding:"required"`
	Delegate  string `json:"delegate" binding:"required"`
	Reason    string `json:"reason,omitempty"`
	Comment   string `json:"comment,omitempty"`
}

// TransferTaskRequest represents task transfer request
type TransferTaskRequest struct {
	TaskID     string `json:"task_id" binding:"required"`
	Transferor string `json:"transferor" binding:"required"`
	Transferee string `json:"transferee" binding:"required"`
	Reason     string `json:"reason,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

// WithdrawTaskRequest represents task withdrawal request
type WithdrawTaskRequest struct {
	TaskID   string `json:"task_id" binding:"required"`
	Operator string `json:"operator" binding:"required"`
	Reason   string `json:"reason,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// UrgeTaskRequest represents task urge request
type UrgeTaskRequest struct {
	TaskID    string     `json:"task_id" binding:"required"`
	Operator  string     `json:"operator" binding:"required"`
	Comment   string     `json:"comment,omitempty"`
	Variables types.JSON `json:"variables,omitempty"`
}

// TerminateProcessRequest represents process termination request
type TerminateProcessRequest struct {
	ProcessID string `json:"process_id" binding:"required"`
	Operator  string `json:"operator" binding:"required"`
	Reason    string `json:"reason,omitempty"`
	Comment   string `json:"comment,omitempty"`
}
