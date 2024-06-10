package structs

import (
	"stocms/pkg/types"
	"time"
)

// Permission represents a permission entity.
type Permission struct {
	ID          string     `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	Action      string     `json:"action,omitempty"`
	Subject     string     `json:"subject,omitempty"`
	Description string     `json:"description,omitempty"`
	Default     bool       `json:"default,omitempty"`
	Disabled    bool       `json:"disabled,omitempty"`
	ExtraProps  types.JSON `json:"extras,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
}

// CreatePermissionBody represents the body for creating or updating a permission.
type CreatePermissionBody struct {
	Permission
}

// UpdatePermissionBody represents the body for updating a permission.
type UpdatePermissionBody struct {
	Permission
	ID string `json:"id,omitempty"`
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
