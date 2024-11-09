package structs

import (
	"fmt"
	"ncobase/common/types"
)

// NodeBody represents a node entity base fields
type NodeBody struct {
	Name           string         `json:"name,omitempty"`
	Type           string         `json:"type,omitempty"`
	Description    string         `json:"description,omitempty"`
	Status         string         `json:"status,omitempty"`
	NodeKey        string         `json:"node_key,omitempty"`
	ProcessID      string         `json:"process_id,omitempty"`
	PrevNodes      []string       `json:"prev_nodes,omitempty"`
	NextNodes      []string       `json:"next_nodes,omitempty"`
	ParallelNodes  []string       `json:"parallel_nodes,omitempty"`
	Conditions     []any          `json:"conditions,omitempty"`
	Properties     map[string]any `json:"properties,omitempty"`
	FormConfig     map[string]any `json:"form_config,omitempty"`
	Permissions    map[string]any `json:"permissions,omitempty"`
	AssigneeConfig map[string]any `json:"assignee_config,omitempty"`
	Handlers       map[string]any `json:"handlers,omitempty"`
	RetryTimes     int            `json:"retry_times,omitempty"`
	RetryInterval  int            `json:"retry_interval,omitempty"`
	IsWorkingDay   bool           `json:"is_working_day,omitempty"`
	Extras         map[string]any `json:"extras,omitempty"`
}

// CreateNodeBody represents the body for creating node
type CreateNodeBody struct {
	NodeBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateNodeBody represents the body for updating node
type UpdateNodeBody struct {
	ID string `json:"id,omitempty"`
	NodeBody
}

// ReadNode represents the output schema for retrieving node
type ReadNode struct {
	ID string `json:"id"`
	NodeBody
	TenantID  string  `json:"tenant_id,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	CreatedAt *int64  `json:"created_at,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
	UpdatedAt *int64  `json:"updated_at,omitempty"`
}

// GetID returns the ID of the node
func (r *ReadNode) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadNode) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadNode) GetSortValue(field string) any {
	switch types.SortField(field) {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	case SortByName:
		return r.Name
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindNodeParams represents query parameters for finding nodes
type FindNodeParams struct {
	ProcessID string          `form:"process_id,omitempty" json:"process_id,omitempty"`
	Type      string          `form:"type,omitempty" json:"type,omitempty"`
	Status    string          `form:"status,omitempty" json:"status,omitempty"`
	NodeKey   string          `form:"node_key,omitempty" json:"node_key,omitempty"`
	Name      string          `form:"name,omitempty" json:"name,omitempty"`
	Tenant    string          `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    types.SortField `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListNodeParams represents list parameters for nodes
type ListNodeParams struct {
	Cursor    string          `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int             `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string          `form:"direction,omitempty" json:"direction,omitempty"`
	ProcessID string          `form:"process_id,omitempty" json:"process_id,omitempty"`
	Type      string          `form:"type,omitempty" json:"type,omitempty"`
	Status    string          `form:"status,omitempty" json:"status,omitempty"`
	Tenant    string          `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    types.SortField `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
