package structs

import (
	"stocms/pkg/types"
)

// Role represents a role entity.
type Role struct {
	ID          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	BaseEntity
}

// CreateRoleBody represents the body for creating or updating a role.
type CreateRoleBody struct {
	Role
}

// UpdateRoleBody represents the body for updating a role.
type UpdateRoleBody struct {
	Role
	ID string `json:"id,omitempty"`
}

// FindRole represents the parameters for finding a role.
type FindRole struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListRoleParams represents the query parameters for listing roles.
type ListRoleParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty"`
}
