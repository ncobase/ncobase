package structs

import (
	"ncobase/common/types"
	"time"
)

// GroupBody represents a group entity.
type GroupBody struct {
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"`
	TenantID    *string     `json:"tenant_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateGroupBody represents the body for creating or updating a group.
type CreateGroupBody struct {
	GroupBody
}

// UpdateGroupBody represents the body for updating a group.
type UpdateGroupBody struct {
	ID string `json:"id,omitempty"`
	GroupBody
}

// ReadGroup represents the output schema for retrieving a group.
type ReadGroup struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Disabled    bool        `json:"disabled"`
	Description string      `json:"description"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"`
	TenantID    *string     `json:"tenant_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *time.Time  `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *time.Time  `json:"updated_at,omitempty"`
}

// FindGroup represents the parameters for finding a group.
type FindGroup struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListGroupParams represents the query parameters for listing groups.
type ListGroupParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
}
