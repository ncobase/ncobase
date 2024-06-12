package structs

import "stocms/pkg/types"

// GroupBody represents a group entity.
type GroupBody struct {
	BaseEntity
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"`
	DomainID    *string     `json:"domain_id,omitempty"`
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
	BaseEntity
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Disabled    bool        `json:"disabled"`
	Description string      `json:"description"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"`
	DomainID    *string     `json:"domain_id,omitempty"`
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

// GroupRole represents the group role.
type GroupRole struct {
	GroupID string `json:"group_id,omitempty"`
	RoleID  string `json:"role_id,omitempty"`
}

// UserGroup represents the user group.
type UserGroup struct {
	UserID  string `json:"user_id,omitempty"`
	GroupID string `json:"group_id,omitempty"`
}
