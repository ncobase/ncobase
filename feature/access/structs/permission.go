package structs

import (
	"fmt"
	"ncobase/common/types"
)

// PermissionBody represents a permission entity.
type PermissionBody struct {
	Name        string      `json:"name,omitempty"`
	Action      string      `json:"action,omitempty"`
	Subject     string      `json:"subject,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     *bool       `json:"default,omitempty"`
	Disabled    *bool       `json:"disabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
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
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Action      string      `json:"action"`
	Subject     string      `json:"subject"`
	Description string      `json:"description"`
	Default     *bool       `json:"default"`
	Disabled    *bool       `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

func (r *ReadPermission) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// FindPermission represents the parameters for finding a permission.
type FindPermission struct {
	ID      string `form:"id,omitempty" json:"id,omitempty"`
	Action  string `form:"action,omitempty" json:"action,omitempty"`
	Subject string `form:"subject,omitempty" json:"subject,omitempty"`
}

// ListPermissionParams represents the query parameters for listing permissions.
type ListPermissionParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
}
