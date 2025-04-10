package structs

import (
	"fmt"
	"github.com/ncobase/ncore/pkg/types"
)

// HistoryBody represents a history entity base fields
type HistoryBody struct {
	Type         string     `json:"type,omitempty"`
	ProcessID    string     `json:"process_id,omitempty"`
	TemplateID   string     `json:"template_id,omitempty"`
	NodeID       string     `json:"node_id,omitempty"`
	NodeName     string     `json:"node_name,omitempty"`
	NodeType     string     `json:"node_type,omitempty"`
	TaskID       string     `json:"task_id,omitempty"`
	Operator     string     `json:"operator,omitempty"`
	OperatorDept string     `json:"operator_dept,omitempty"`
	Action       string     `json:"action,omitempty"`
	Comment      string     `json:"comment,omitempty"`
	Variables    types.JSON `json:"variables,omitempty"`
	FormData     types.JSON `json:"form_data,omitempty"`
	NodeConfig   types.JSON `json:"node_config,omitempty"`
	Details      types.JSON `json:"details,omitempty"`
	TenantID     string     `json:"tenant_id,omitempty"`
}

// CreateHistoryBody represents the body for creating history
type CreateHistoryBody struct {
	HistoryBody
	TenantID string `json:"tenant_id,omitempty"`
}

// ReadHistory represents the output schema for retrieving history
type ReadHistory struct {
	ID           string     `json:"id"`
	Type         string     `json:"type,omitempty"`
	ProcessID    string     `json:"process_id,omitempty"`
	TemplateID   string     `json:"template_id,omitempty"`
	NodeID       string     `json:"node_id,omitempty"`
	NodeName     string     `json:"node_name,omitempty"`
	NodeType     string     `json:"node_type,omitempty"`
	TaskID       string     `json:"task_id,omitempty"`
	Operator     string     `json:"operator,omitempty"`
	OperatorDept string     `json:"operator_dept,omitempty"`
	Action       string     `json:"action,omitempty"`
	Comment      string     `json:"comment,omitempty"`
	Variables    types.JSON `json:"variables,omitempty"`
	FormData     types.JSON `json:"form_data,omitempty"`
	NodeConfig   types.JSON `json:"node_config,omitempty"`
	Details      types.JSON `json:"details,omitempty"`
	TenantID     string     `json:"tenant_id,omitempty"`
	CreatedBy    *string    `json:"created_by,omitempty"`
	CreatedAt    *int64     `json:"created_at,omitempty"`
	UpdatedBy    *string    `json:"updated_by,omitempty"`
	UpdatedAt    *int64     `json:"updated_at,omitempty"`
}

// GetID returns the ID of the history
func (r *ReadHistory) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadHistory) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadHistory) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindHistoryParams represents query parameters for finding histories
type FindHistoryParams struct {
	ProcessID  string `form:"process_id,omitempty" json:"process_id,omitempty"`
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	NodeID     string `form:"node_id,omitempty" json:"node_id,omitempty"`
	TaskID     string `form:"task_id,omitempty" json:"task_id,omitempty"`
	Operator   string `form:"operator,omitempty" json:"operator,omitempty"`
	Action     string `form:"action,omitempty" json:"action,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListHistoryParams represents list parameters for histories
type ListHistoryParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	NodeID    string `form:"node_id,omitempty" json:"node_id,omitempty"`
	ProcessID string `form:"process_id,omitempty" json:"process_id,omitempty"`
	TaskID    string `form:"task_id,omitempty" json:"task_id,omitempty"`
	Operator  string `form:"operator,omitempty" json:"operator,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
