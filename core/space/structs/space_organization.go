package structs

import (
	orgStructs "ncobase/organization/structs"
)

// SpaceOrganization represents the space-organization relationship
type SpaceOrganization struct {
	SpaceID string `json:"space_id,omitempty"`
	OrgID   string `json:"org_id,omitempty"`
}

// AddSpaceOrganizationRequest represents the request to add a organization to space
type AddSpaceOrganizationRequest struct {
	OrgID string `json:"org_id" binding:"required"`
}

// SpaceOrganizationRelation represents a space-organization relationship with metadata
type SpaceOrganizationRelation struct {
	ID      string `json:"id,omitempty"`
	SpaceID string `json:"space_id,omitempty"`
	OrgID   string `json:"org_id,omitempty"`
	AddedAt int64  `json:"added_at,omitempty"`
}

// ListOrganizationParams represents the query parameters for listing orgs in space.
type ListOrganizationParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Parent    string `form:"parent,omitempty" json:"parent,omitempty"`
	Children  bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ReadOrganization represents the output schema for retrieving a organization
type ReadOrganization = orgStructs.ReadOrganization
