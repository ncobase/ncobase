package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// MenuBody represents a menu entity.
type MenuBody struct {
	Name      string      `json:"name,omitempty"`
	Label     string      `json:"label,omitempty"`
	Slug      string      `json:"slug,omitempty"`
	Type      string      `json:"type,omitempty"`
	Path      string      `json:"path,omitempty"`
	Target    string      `json:"target,omitempty"`
	Icon      string      `json:"icon,omitempty"`
	Perms     string      `json:"perms,omitempty"`
	Hidden    *bool       `json:"hidden,omitempty"`
	Order     *int        `json:"order,omitempty"`
	Disabled  *bool       `json:"disabled,omitempty"`
	Extras    *types.JSON `json:"extras,omitempty"`
	ParentID  string      `json:"parent_id,omitempty"`
	TenantID  string      `json:"tenant_id,omitempty"`
	CreatedBy *string     `json:"created_by,omitempty"`
	UpdatedBy *string     `json:"updated_by,omitempty"`
}

// CreateMenuBody represents the body for creating or updating a menu.
type CreateMenuBody struct {
	MenuBody
}

// UpdateMenuBody represents the body for updating a menu.
type UpdateMenuBody struct {
	ID string `json:"id,omitempty"`
	MenuBody
}

// ReadMenu represents the output schema for retrieving a menu.
type ReadMenu struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Label     string           `json:"label"`
	Slug      string           `json:"slug"`
	Type      string           `json:"type"`
	Path      string           `json:"path"`
	Target    string           `json:"target"`
	Icon      string           `json:"icon"`
	Perms     string           `json:"perms"`
	Hidden    bool             `json:"hidden"`
	Order     int              `json:"order"`
	Disabled  bool             `json:"disabled"`
	Extras    *types.JSON      `json:"extras,omitempty"`
	ParentID  string           `json:"parent_id,omitempty"`
	TenantID  string           `json:"tenant_id,omitempty"`
	Children  []types.TreeNode `json:"children,omitempty"`
	CreatedBy *string          `json:"created_by,omitempty"`
	CreatedAt *int64           `json:"created_at,omitempty"`
	UpdatedBy *string          `json:"updated_by,omitempty"`
	UpdatedAt *int64           `json:"updated_at,omitempty"`
}

// GetID returns the ID of the menu.
func (r *ReadMenu) GetID() string {
	return r.ID
}

// GetParentID returns the parent ID of the menu.
func (r *ReadMenu) GetParentID() string {
	return r.ParentID
}

// SetChildren sets the children of the menu.
func (r *ReadMenu) SetChildren(children []types.TreeNode) {
	r.Children = children
}

// GetChildren returns the children of the menu.
func (r *ReadMenu) GetChildren() []types.TreeNode {
	return r.Children
}

// GetCursorValue returns the cursor value.
func (r *ReadMenu) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadMenu) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return convert.ToValue(r.CreatedAt)
	case SortByOrder:
		return r.Order
	default:
		return convert.ToValue(r.CreatedAt)
	}
}

// FindMenu represents the parameters for finding a menu.
type FindMenu struct {
	Menu     string `form:"menu,omitempty" json:"menu,omitempty"`
	Parent   string `form:"parent,omitempty" json:"parent,omitempty"`
	Tenant   string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty"`
	Children bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy   string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListMenuParams represents the query parameters for listing menus.
type ListMenuParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Parent    string `form:"parent,omitempty" json:"parent,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Perms     string `form:"perms,omitempty" json:"perms,omitempty"`
	Children  bool   `form:"children,omitempty" json:"children,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
