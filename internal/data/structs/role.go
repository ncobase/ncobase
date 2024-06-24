package structs

import (
	"ncobase/common/types"
)

// RoleBody represents a role entity.
type RoleBody struct {
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	OperatorBy
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
	BaseEntity
}

// FindRole represents the parameters for finding a role.
type FindRole struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListRoleParams represents the query parameters for listing roles.
type ListRoleParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
}

// UserRole represents the user role.
type UserRole struct {
	UserID string `json:"user_id,omitempty"`
	RoleID string `json:"role_id,omitempty"`
}

// UserTenantRole represents the user tenant role.
type UserTenantRole struct {
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
}
