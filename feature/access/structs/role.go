package structs

import (
	"fmt"
	"ncobase/common/types"
)

// RoleBody represents a role entity.
type RoleBody struct {
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateRoleBody represents the body for creating or updating a role.
type CreateRoleBody struct {
	RoleBody
}

// UpdateRoleBody represents the body for updating a role.
type UpdateRoleBody struct {
	ID string `json:"id,omitempty"`
	RoleBody
}

// ReadRole represents the output schema for retrieving a role.
type ReadRole struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Disabled    bool        `json:"disabled"`
	Description string      `json:"description"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

func (r *ReadRole) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, r.CreatedAt)
}

// FindRole represents the parameters for finding a role.
type FindRole struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
	Name string `form:"name,omitempty" json:"name,omitempty"`
}

// ListRoleParams represents the query parameters for listing roles.
type ListRoleParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset    int    `form:"offset,omitempty" json:"offset,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
}
