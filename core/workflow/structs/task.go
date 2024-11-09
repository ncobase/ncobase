package structs

import (
	"fmt"
	"ncobase/common/types"
	"time"
)

// TaskBody represents a task entity base fields
type TaskBody struct {
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	Status       string         `json:"status,omitempty"`
	TaskKey      string         `json:"task_key,omitempty"`
	ProcessID    string         `json:"process_id,omitempty"`
	NodeKey      string         `json:"node_key,omitempty"`
	NodeType     string         `json:"node_type,omitempty"`
	Assignee     string         `json:"assignee,omitempty"`
	AssigneeDept string         `json:"assignee_dept,omitempty"`
	Candidates   []string       `json:"candidates,omitempty"`
	Action       string         `json:"action,omitempty"`
	Comment      string         `json:"comment,omitempty"`
	FormData     map[string]any `json:"form_data,omitempty"`
	Variables    map[string]any `json:"variables,omitempty"`
	Priority     int            `json:"priority,omitempty"`
	IsUrged      bool           `json:"is_urged,omitempty"`
	UrgeCount    int            `json:"urge_count,omitempty"`
	IsTimeout    bool           `json:"is_timeout,omitempty"`
	Extras       map[string]any `json:"extras,omitempty"`
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
	ID string `json:"id"`
	TaskBody
	TenantID  string     `json:"tenant_id,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	ClaimTime *time.Time `json:"claim_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	DueTime   *time.Time `json:"due_time,omitempty"`
	Duration  *int       `json:"duration,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt *int64     `json:"created_at,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	UpdatedAt *int64     `json:"updated_at,omitempty"`
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
	switch types.SortField(field) {
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
	ProcessID   string          `form:"process_id,omitempty" json:"process_id,omitempty"`
	NodeKey     string          `form:"node_key,omitempty" json:"node_key,omitempty"`
	NodeType    string          `form:"node_type,omitempty" json:"node_type,omitempty"`
	Status      string          `form:"status,omitempty" json:"status,omitempty"`
	Assignee    string          `form:"assignee,omitempty" json:"assignee,omitempty"`
	IsUrged     *bool           `form:"is_urged,omitempty" json:"is_urged,omitempty"`
	IsTimeout   *bool           `form:"is_timeout,omitempty" json:"is_timeout,omitempty"`
	Priority    *int            `form:"priority,omitempty" json:"priority,omitempty"`
	DueTimeFrom *time.Time      `form:"due_time_from,omitempty" json:"due_time_from,omitempty"`
	DueTimeTo   *time.Time      `form:"due_time_to,omitempty" json:"due_time_to,omitempty"`
	Tenant      string          `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy      types.SortField `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListTaskParams represents list parameters for tasks
type ListTaskParams struct {
	Cursor    string          `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int             `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string          `form:"direction,omitempty" json:"direction,omitempty"`
	ProcessID string          `form:"process_id,omitempty" json:"process_id,omitempty"`
	NodeType  string          `form:"node_type,omitempty" json:"node_type,omitempty"`
	Status    string          `form:"status,omitempty" json:"status,omitempty"`
	Assignee  string          `form:"assignee,omitempty" json:"assignee,omitempty"`
	IsTimeout *bool           `form:"is_timeout,omitempty" json:"is_timeout,omitempty"`
	Priority  *int            `form:"priority,omitempty" json:"priority,omitempty"`
	Tenant    string          `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    types.SortField `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
