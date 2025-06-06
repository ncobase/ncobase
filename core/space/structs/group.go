package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// GroupBody represents a group entity.
type GroupBody struct {
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Type        string      `json:"type,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Description string      `json:"description,omitempty"`
	Leader      *types.JSON `json:"leader,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    *string     `json:"parent_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateGroupBody represents the body for creating a group.
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
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Type        string           `json:"type"`
	Disabled    bool             `json:"disabled"`
	Description string           `json:"description"`
	Leader      *types.JSON      `json:"leader,omitempty"`
	Extras      *types.JSON      `json:"extras,omitempty"`
	ParentID    *string          `json:"parent_id,omitempty"`
	Children    []types.TreeNode `json:"children,omitempty"`
	CreatedBy   *string          `json:"created_by,omitempty"`
	CreatedAt   *int64           `json:"created_at,omitempty"`
	UpdatedBy   *string          `json:"updated_by,omitempty"`
	UpdatedAt   *int64           `json:"updated_at,omitempty"`
}

// GetID returns the ID of the group.
func (r *ReadGroup) GetID() string {
	return r.ID
}

// GetParentID returns the parent ID of the group.
func (r *ReadGroup) GetParentID() string {
	return convert.ToValue(r.ParentID)
}

// SetChildren sets the children of the group.
func (r *ReadGroup) SetChildren(children []types.TreeNode) {
	r.Children = children
}

// GetChildren returns the children of the group.
func (r *ReadGroup) GetChildren() []types.TreeNode {
	return r.Children
}

// GetCursorValue returns the cursor value.
func (r *ReadGroup) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadGroup) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return convert.ToValue(r.CreatedAt)
	default:
		return convert.ToValue(r.CreatedAt)
	}
}

// FindGroup represents the parameters for finding a group.
type FindGroup struct {
	Group    string `form:"group,omitempty" json:"group,omitempty"`
	Parent   string `form:"parent,omitempty" json:"parent,omitempty"`
	Children bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy   string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListGroupParams represents the query parameters for listing groups.
type ListGroupParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Parent    string `form:"parent,omitempty" json:"parent,omitempty"`
	Children  bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
