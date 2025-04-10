package structs

import (
	"fmt"

	"github.com/ncobase/ncore/pkg/types"
)

// NodeBody represents a node entity base fields
type NodeBody struct {
	Name             string            `json:"name,omitempty"`
	Type             string            `json:"type,omitempty"`
	Required         bool              `json:"required,omitempty"`
	Skippable        bool              `json:"skippable,omitempty"`
	BranchConditions []string          `json:"branch_conditions,omitempty"`
	DefaultBranch    string            `json:"default_branch,omitempty"`
	Description      string            `json:"description,omitempty"`
	Status           string            `json:"status,omitempty"`
	NodeKey          string            `json:"node_key,omitempty"`
	ProcessID        string            `json:"process_id,omitempty"`
	TemplateID       string            `json:"template_id,omitempty"`
	PrevNodes        types.StringArray `json:"prev_nodes,omitempty"`
	NextNodes        types.StringArray `json:"next_nodes,omitempty"`
	ParallelNodes    types.StringArray `json:"parallel_nodes,omitempty"`
	BranchNodes      types.StringArray `json:"branch_nodes,omitempty"`
	Conditions       types.StringArray `json:"conditions,omitempty"`
	Properties       types.JSON        `json:"properties,omitempty"`
	FormConfig       types.JSON        `json:"form_config,omitempty"`
	Permissions      types.JSON        `json:"permissions,omitempty"`
	Assignees        types.JSON        `json:"assignees,omitempty"`
	Handlers         types.JSON        `json:"handlers,omitempty"`
	RetryTimes       int               `json:"retry_times,omitempty"`
	RetryInterval    int               `json:"retry_interval,omitempty"`
	IsWorkingDay     bool              `json:"is_working_day,omitempty"`
	TimeoutConfig    types.JSON        `json:"timeout_config,omitempty"`
	TimeoutDuration  int               `json:"timeout_duration,omitempty"`
	TenantID         string            `json:"tenant_id,omitempty"`
	Variables        types.JSON        `json:"variables,omitempty"`
	Extras           types.JSON        `json:"extras,omitempty"`
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
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	Type             string            `json:"type,omitempty"`
	Required         bool              `json:"required,omitempty"`
	Skippable        bool              `json:"skippable,omitempty"`
	BranchConditions types.StringArray `json:"branch_conditions,omitempty"`
	DefaultBranch    string            `json:"default_branch,omitempty"`
	Description      string            `json:"description,omitempty"`
	Status           string            `json:"status,omitempty"`
	NodeKey          string            `json:"node_key,omitempty"`
	ProcessID        string            `json:"process_id,omitempty"`
	TemplateID       string            `json:"template_id,omitempty"`
	PrevNodes        types.StringArray `json:"prev_nodes,omitempty"`
	NextNodes        types.StringArray `json:"next_nodes,omitempty"`
	ParallelNodes    types.StringArray `json:"parallel_nodes,omitempty"`
	BranchNodes      types.StringArray `json:"branch_nodes,omitempty"`
	Conditions       types.StringArray `json:"conditions,omitempty"`
	Properties       types.JSON        `json:"properties,omitempty"`
	FormConfig       types.JSON        `json:"form_config,omitempty"`
	Permissions      types.JSON        `json:"permissions,omitempty"`
	Assignees        types.StringArray `json:"assignees,omitempty"`
	Handlers         types.JSON        `json:"handlers,omitempty"`
	RetryTimes       int               `json:"retry_times,omitempty"`
	RetryInterval    int               `json:"retry_interval,omitempty"`
	IsWorkingDay     bool              `json:"is_working_day,omitempty"`
	TimeoutConfig    types.JSON        `json:"timeout_config,omitempty"`
	TimeoutDuration  int               `json:"timeout_duration,omitempty"`
	TenantID         string            `json:"tenant_id,omitempty"`
	Variables        types.JSON        `json:"variables,omitempty"`
	Extras           types.JSON        `json:"extras,omitempty"`
	CreatedBy        *string           `json:"created_by,omitempty"`
	CreatedAt        *int64            `json:"created_at,omitempty"`
	UpdatedBy        *string           `json:"updated_by,omitempty"`
	UpdatedAt        *int64            `json:"updated_at,omitempty"`
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
	switch field {
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
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	ProcessID  string `form:"process_id,omitempty" json:"process_id,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	NodeKey    string `form:"node_key,omitempty" json:"node_key,omitempty"`
	Name       string `form:"name,omitempty" json:"name,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListNodeParams represents list parameters for nodes
type ListNodeParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	ProcessID string `form:"process_id,omitempty" json:"process_id,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Status    string `form:"status,omitempty" json:"status,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
