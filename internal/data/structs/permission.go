package structs

import (
	"stocms/pkg/types"
)

// PermissionBody represents a permission entity.
type PermissionBody struct {
	BaseEntity
	Name        string      `json:"name,omitempty"`
	Action      string      `json:"action,omitempty"`
	Subject     string      `json:"subject,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     *bool       `json:"default,omitempty"`
	Disabled    *bool       `json:"disabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
}

// CreatePermissionBody represents the body for creating a permission.
type CreatePermissionBody struct {
	PermissionBody
}

// UpdatePermissionBody represents the body for updating a permission.
type UpdatePermissionBody struct {
	ID string `json:"id,omitempty"`
	PermissionBody
}

// ReadPermission represents the output schema for retrieving a permission.
type ReadPermission struct {
	BaseEntity
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Action      string      `json:"action"`
	Subject     string      `json:"subject"`
	Description string      `json:"description"`
	Default     *bool       `json:"default"`
	Disabled    *bool       `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
}

// FindPermission represents the parameters for finding a permission.
type FindPermission struct {
	ID      string `form:"id,omitempty" json:"id,omitempty"`
	Action  string `form:"action,omitempty" json:"action,omitempty"`
	Subject string `form:"subject,omitempty" json:"subject,omitempty"`
}

// ListPermissionParams represents the query parameters for listing permissions.
type ListPermissionParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty"`
}

// RolePermission represents the role permission relationship.
type RolePermission struct {
	RoleID       string `json:"role_id,omitempty"`
	PermissionID string `json:"permission_id,omitempty"`
}
