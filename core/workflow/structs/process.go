package structs

import (
	"encoding/json"
	"fmt"

	"github.com/ncobase/ncore/types"
)

// ProcessBody represents a process entity base fields
type ProcessBody struct {
	Title         string            `json:"title,omitempty"`
	Category      string            `json:"category,omitempty"`
	StartNodeID   string            `json:"start_node_id,omitempty"`
	EndNodeID     string            `json:"end_node_id,omitempty"`
	MaxDuration   int               `json:"max_duration,omitempty"`
	Retryable     bool              `json:"retryable,omitempty"`
	Callbacks     types.JSON        `json:"callbacks,omitempty"`
	ProcessKey    string            `json:"process_key,omitempty"`
	Status        string            `json:"status,omitempty"`
	TemplateID    string            `json:"template_id,omitempty"`
	BusinessKey   string            `json:"business_key,omitempty"`
	ModuleCode    string            `json:"module_code,omitempty"`
	FormCode      string            `json:"form_code,omitempty"`
	Initiator     string            `json:"initiator,omitempty"`
	InitiatorDept string            `json:"initiator_dept,omitempty"`
	ProcessCode   string            `json:"process_code,omitempty"`
	Variables     types.JSON        `json:"variables,omitempty"`
	CurrentNode   string            `json:"current_node,omitempty"`
	ActiveNodes   types.StringArray `json:"active_nodes,omitempty"`
	FlowStatus    string            `json:"flow_status,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	IsSuspended   bool              `json:"is_suspended,omitempty"`
	SuspendReason string            `json:"suspend_reason,omitempty"`
	AllowCancel   bool              `json:"allow_cancel,omitempty"`
	AllowUrge     bool              `json:"allow_urge,omitempty"`
	UrgeCount     int               `json:"urge_count,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	StartTime     *int64            `json:"start_time,omitempty"`
	EndTime       *int64            `json:"end_time,omitempty"`
	DueDate       *int64            `json:"due_date,omitempty"`
	Duration      *int              `json:"duration,omitempty"`
	ParentID      *string           `json:"parent_id,omitempty"`
	Extras        types.JSON        `json:"extras,omitempty"`
}

// CreateProcessBody represents the body for creating process
type CreateProcessBody struct {
	ProcessBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateProcessBody represents the body for updating process
type UpdateProcessBody struct {
	ID string `json:"id,omitempty"`
	ProcessBody
}

// ReadProcess represents the output schema for retrieving process
type ReadProcess struct {
	ID            string            `json:"id"`
	Title         string            `json:"title,omitempty"`
	Category      string            `json:"category,omitempty"`
	StartNodeID   string            `json:"start_node_id,omitempty"`
	EndNodeID     string            `json:"end_node_id,omitempty"`
	MaxDuration   int               `json:"max_duration,omitempty"`
	Retryable     bool              `json:"retryable,omitempty"`
	Callbacks     types.JSON        `json:"callbacks,omitempty"`
	ProcessKey    string            `json:"process_key,omitempty"`
	Status        string            `json:"status,omitempty"`
	TemplateID    string            `json:"template_id,omitempty"`
	BusinessKey   string            `json:"business_key,omitempty"`
	ModuleCode    string            `json:"module_code,omitempty"`
	FormCode      string            `json:"form_code,omitempty"`
	Initiator     string            `json:"initiator,omitempty"`
	InitiatorDept string            `json:"initiator_dept,omitempty"`
	ProcessCode   string            `json:"process_code,omitempty"`
	Variables     types.JSON        `json:"variables,omitempty"`
	CurrentNode   string            `json:"current_node,omitempty"`
	ActiveNodes   types.StringArray `json:"active_nodes,omitempty"`
	FlowStatus    string            `json:"flow_status,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	IsSuspended   bool              `json:"is_suspended,omitempty"`
	SuspendReason string            `json:"suspend_reason,omitempty"`
	AllowCancel   bool              `json:"allow_cancel,omitempty"`
	AllowUrge     bool              `json:"allow_urge,omitempty"`
	UrgeCount     int               `json:"urge_count,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	StartTime     *int64            `json:"start_time,omitempty"`
	EndTime       *int64            `json:"end_time,omitempty"`
	DueDate       *int64            `json:"due_date,omitempty"`
	Duration      *int              `json:"duration,omitempty"`
	ParentID      *string           `json:"parent_id,omitempty"`
	Extras        types.JSON        `json:"extras,omitempty"`
	CreatedBy     *string           `json:"created_by,omitempty"`
	CreatedAt     *int64            `json:"created_at,omitempty"`
	UpdatedBy     *string           `json:"updated_by,omitempty"`
	UpdatedAt     *int64            `json:"updated_at,omitempty"`
}

// GetID returns the ID of the process
func (r *ReadProcess) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadProcess) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadProcess) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	case SortByPriority:
		return r.Priority
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// ProcessSnapshot represents a process entity base fields
type ProcessSnapshot struct {
	ID           string          `json:"id"`
	ProcessID    string          `json:"process_id"`
	ProcessData  json.RawMessage `json:"process_data"`
	NodeData     json.RawMessage `json:"node_data"`
	TaskData     json.RawMessage `json:"task_data"`
	BusinessData json.RawMessage `json:"business_data"`
	CreatedAt    int64           `json:"created_at"`
	CreatedBy    string          `json:"created_by"`
	Comment      string          `json:"comment"`
}

// FindProcessParams represents query parameters for finding processes
type FindProcessParams struct {
	ProcessKey  string `form:"process_key,omitempty" json:"process_key,omitempty"`
	TemplateID  string `form:"template_id,omitempty" json:"template_id,omitempty"`
	BusinessKey string `form:"business_key,omitempty" json:"business_key,omitempty"`
	ModuleCode  string `form:"module_code,omitempty" json:"module_code,omitempty"`
	FormCode    string `form:"form_code,omitempty" json:"form_code,omitempty"`
	Status      string `form:"status,omitempty" json:"status,omitempty"`
	FlowStatus  string `form:"flow_status,omitempty" json:"flow_status,omitempty"`
	Initiator   string `form:"initiator,omitempty" json:"initiator,omitempty"`
	Priority    *int   `form:"priority,omitempty" json:"priority,omitempty"`
	IsSuspended *bool  `form:"is_suspended,omitempty" json:"is_suspended,omitempty"`
	StartFrom   *int64 `form:"start_from,omitempty" json:"start_from,omitempty"`
	StartTo     *int64 `form:"start_to,omitempty" json:"start_to,omitempty"`
	Tenant      string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy      string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListProcessParams represents list parameters for processes
type ListProcessParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Status    string `form:"status,omitempty" json:"status,omitempty"`
	Initiator string `form:"initiator,omitempty" json:"initiator,omitempty"`
	Priority  *int   `form:"priority,omitempty" json:"priority,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
