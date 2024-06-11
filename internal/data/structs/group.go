package structs

import (
	"stocms/pkg/types"
)

// Group represents a group entity.
type Group struct {
	ID          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"` // Use pointer for nullable field
	DomainID    *string     `json:"domain_id,omitempty"` // Use pointer for nullable field
	BaseEntity
}

// CreateGroupBody represents the body for creating or updating a group.
type CreateGroupBody struct {
	Group
}

// UpdateGroupBody represents the body for updating a group.
type UpdateGroupBody struct {
	Group
	ID string `json:"id,omitempty"`
}

// FindGroup represents the parameters for finding a group.
type FindGroup struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListGroupParams represents the query parameters for listing groups.
type ListGroupParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty"`
}
