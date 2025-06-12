package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// OrganizationBody represents an organization entity.
type OrganizationBody struct {
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

// CreateOrganizationBody represents the body for creating an organization.
type CreateOrganizationBody struct {
	OrganizationBody
}

// UpdateOrganizationBody represents the body for updating an organization.
type UpdateOrganizationBody struct {
	ID string `json:"id,omitempty"`
	OrganizationBody
}

// ReadOrganization represents the output schema for retrieving an organization.
type ReadOrganization struct {
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

// GetID returns the ID of the organization.
func (r *ReadOrganization) GetID() string {
	return r.ID
}

// GetParentID returns the parent ID of the organization.
func (r *ReadOrganization) GetParentID() string {
	return convert.ToValue(r.ParentID)
}

// SetChildren sets the children of the organization.
func (r *ReadOrganization) SetChildren(children []types.TreeNode) {
	r.Children = children
}

// GetChildren returns the children of the organization.
func (r *ReadOrganization) GetChildren() []types.TreeNode {
	return r.Children
}

// GetCursorValue returns the cursor value.
func (r *ReadOrganization) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadOrganization) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return convert.ToValue(r.CreatedAt)
	default:
		return convert.ToValue(r.CreatedAt)
	}
}

// FindOrganization represents the parameters for finding an organization.
type FindOrganization struct {
	Organization string `form:"organization,omitempty" json:"organization,omitempty"`
	Parent       string `form:"parent,omitempty" json:"parent,omitempty"`
	Children     bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy       string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListOrganizationParams represents the query parameters for listing organizations.
type ListOrganizationParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Parent    string `form:"parent,omitempty" json:"parent,omitempty"`
	Children  bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
