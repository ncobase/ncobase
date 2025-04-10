package structs

import (
	"fmt"
	"github.com/ncobase/ncore/pkg/types"
)

// TaskBody represents a task entity base fields
type TaskBody struct {
	Name           string            `json:"name,omitempty"`
	Category       CategoryType      `json:"category,omitempty"`
	Weight         PriorityStrategy  `json:"weight,omitempty"`
	ParentTaskID   string            `json:"parent_task_id,omitempty"`
	SubTasks       types.StringArray `json:"sub_tasks,omitempty"`
	AllowedActions types.StringArray `json:"allowed_actions,omitempty"`
	Restrictions   types.JSON        `json:"restrictions,omitempty"`
	Description    string            `json:"description,omitempty"`
	Status         string            `json:"status,omitempty"`
	TaskKey        string            `json:"task_key,omitempty"`
	ProcessID      string            `json:"process_id,omitempty"`
	NodeKey        string            `json:"node_key,omitempty"`
	NodeType       string            `json:"node_type,omitempty"`
	Assignees      types.StringArray `json:"assignees,omitempty"`
	Candidates     types.StringArray `json:"candidates,omitempty"`
	Action         string            `json:"action,omitempty"`
	Comment        string            `json:"comment,omitempty"`
	FormData       types.JSON        `json:"form_data,omitempty"`
	Variables      types.JSON        `json:"variables,omitempty"`
	Priority       int               `json:"priority,omitempty"`
	IsUrged        bool              `json:"is_urged,omitempty"`
	UrgeCount      int               `json:"urge_count,omitempty"`
	IsTransferred  bool              `json:"is_transferred,omitempty"`
	IsResubmit     bool              `json:"is_resubmit,omitempty"`
	IsDelegated    bool              `json:"is_delegated,omitempty"`
	DelegatedFrom  types.JSON        `json:"delegated_from,omitempty"`
	IsTimeout      bool              `json:"is_timeout,omitempty"`
	TenantID       string            `json:"tenant_id,omitempty"`
	StartTime      *int64            `json:"start_time,omitempty"`
	ClaimTime      *int64            `json:"claim_time,omitempty"`
	AssignStrategy string            `json:"assign_strategy,omitempty"`
	EndTime        *int64            `json:"end_time,omitempty"`
	DueTime        *int64            `json:"due_time,omitempty"`
	Duration       *int              `json:"duration,omitempty"`
	Extras         types.JSON        `json:"extras,omitempty"`
}

// CreateTaskBody represents the body for creating task
type CreateTaskBody struct {
	TaskBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateTaskBody represents the body for updating task
type UpdateTaskBody struct {
	ID string `json:"id,omitempty"`
	TaskBody
}

// ReadTask represents the output schema for retrieving task
type ReadTask struct {
	ID             string            `json:"id"`
	Name           string            `json:"name,omitempty"`
	Category       CategoryType      `json:"category,omitempty"`
	Weight         PriorityStrategy  `json:"weight,omitempty"`
	ParentTaskID   string            `json:"parent_task_id,omitempty"`
	SubTasks       types.StringArray `json:"sub_tasks,omitempty"`
	AllowedActions types.StringArray `json:"allowed_actions,omitempty"`
	Restrictions   types.JSON        `json:"restrictions,omitempty"`
	Description    string            `json:"description,omitempty"`
	Status         string            `json:"status,omitempty"`
	TaskKey        string            `json:"task_key,omitempty"`
	ProcessID      string            `json:"process_id,omitempty"`
	NodeKey        string            `json:"node_key,omitempty"`
	NodeType       string            `json:"node_type,omitempty"`
	Assignees      types.StringArray `json:"assignees,omitempty"`
	Candidates     types.StringArray `json:"candidates,omitempty"`
	Action         string            `json:"action,omitempty"`
	Comment        string            `json:"comment,omitempty"`
	FormData       types.JSON        `json:"form_data,omitempty"`
	Variables      types.JSON        `json:"variables,omitempty"`
	Priority       int               `json:"priority,omitempty"`
	IsUrged        bool              `json:"is_urged,omitempty"`
	UrgeCount      int               `json:"urge_count,omitempty"`
	IsTransferred  bool              `json:"is_transferred,omitempty"`
	IsResubmit     bool              `json:"is_resubmit,omitempty"`
	IsDelegated    bool              `json:"is_delegated,omitempty"`
	DelegatedFrom  types.JSON        `json:"delegated_from,omitempty"`
	IsTimeout      bool              `json:"is_timeout,omitempty"`
	TenantID       string            `json:"tenant_id,omitempty"`
	StartTime      *int64            `json:"start_time,omitempty"`
	ClaimTime      *int64            `json:"claim_time,omitempty"`
	AssignStrategy string            `json:"assign_strategy,omitempty"`
	EndTime        *int64            `json:"end_time,omitempty"`
	DueTime        *int64            `json:"due_time,omitempty"`
	Duration       *int              `json:"duration,omitempty"`
	Extras         types.JSON        `json:"extras,omitempty"`
	CreatedBy      *string           `json:"created_by,omitempty"`
	CreatedAt      *int64            `json:"created_at,omitempty"`
	UpdatedBy      *string           `json:"updated_by,omitempty"`
	UpdatedAt      *int64            `json:"updated_at,omitempty"`
}

// GetID returns the ID of the task
func (r *ReadTask) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadTask) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadTask) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	case SortByPriority:
		return r.Priority
	case SortByDueTime:
		return r.DueTime
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindTaskParams represents query parameters for finding tasks
type FindTaskParams struct {
	ID          string            `form:"id,omitempty" json:"id,omitempty"`
	ProcessID   string            `form:"process_id,omitempty" json:"process_id,omitempty"`
	NodeKey     string            `form:"node_key,omitempty" json:"node_key,omitempty"`
	NodeType    string            `form:"node_type,omitempty" json:"node_type,omitempty"`
	Status      string            `form:"status,omitempty" json:"status,omitempty"`
	Assignees   types.StringArray `form:"assignees,omitempty" json:"assignees,omitempty"`
	IsUrged     *bool             `form:"is_urged,omitempty" json:"is_urged,omitempty"`
	IsTimeout   *bool             `form:"is_timeout,omitempty" json:"is_timeout,omitempty"`
	Priority    *int              `form:"priority,omitempty" json:"priority,omitempty"`
	DueTimeFrom *int64            `form:"due_time_from,omitempty" json:"due_time_from,omitempty"`
	DueTimeTo   *int64            `form:"due_time_to,omitempty" json:"due_time_to,omitempty"`
	Tenant      string            `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy      string            `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// TaskCondition represents the condition for finding tasks
type TaskCondition struct {
	ProcessIDs []string `json:"process_ids,omitempty"`
	NodeTypes  []string `json:"node_types,omitempty"`
	Statuses   []string `json:"statuses,omitempty"`
	Assignees  []string `json:"assignees,omitempty"`
	Priority   *int     `json:"priority,omitempty"`
	StartTime  *int64   `json:"start_time,omitempty"`
	EndTime    *int64   `json:"end_time,omitempty"`
	IsTimeout  *bool    `json:"is_timeout,omitempty"`
	IsUrged    *bool    `json:"is_urged,omitempty"`
	SortBy     []string `json:"sort_by,omitempty"`
	OrderBy    []string `json:"order_by,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// ListTaskParams represents list parameters for tasks
type ListTaskParams struct {
	Cursor    string            `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int               `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string            `form:"direction,omitempty" json:"direction,omitempty"`
	ProcessID string            `form:"process_id,omitempty" json:"process_id,omitempty"`
	NodeType  string            `form:"node_type,omitempty" json:"node_type,omitempty"`
	NodeKey   string            `form:"node_key,omitempty" json:"node_key,omitempty"`
	Status    string            `form:"status,omitempty" json:"status,omitempty"`
	Assignees types.StringArray `form:"assignees,omitempty" json:"assignees,omitempty"`
	IsTimeout *bool             `form:"is_timeout,omitempty" json:"is_timeout,omitempty"`
	Priority  *int              `form:"priority,omitempty" json:"priority,omitempty"`
	Tenant    string            `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    string            `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
