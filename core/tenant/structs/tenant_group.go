package structs

import (
	spaceStructs "ncobase/space/structs"
)

// TenantGroup represents the tenant-group relationship
type TenantGroup struct {
	TenantID string `json:"tenant_id,omitempty"`
	GroupID  string `json:"group_id,omitempty"`
}

// AddTenantGroupRequest represents the request to add a group to tenant
type AddTenantGroupRequest struct {
	GroupID string `json:"group_id" binding:"required"`
}

// TenantGroupRelation represents a tenant-group relationship with metadata
type TenantGroupRelation struct {
	ID       string `json:"id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	GroupID  string `json:"group_id,omitempty"`
	AddedAt  int64  `json:"added_at,omitempty"`
}

// ListGroupParams represents the query parameters for listing groups in tenant.
type ListGroupParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Parent    string `form:"parent,omitempty" json:"parent,omitempty"`
	Children  bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ReadGroup represents the output schema for retrieving a group
type ReadGroup = spaceStructs.ReadGroup
