package structs

import (
	"fmt"
	"ncobase/common/types"
)

// FindTaxonomy represents the parameters for finding a taxonomy.
type FindTaxonomy struct {
	ID       string `json:"id,omitempty"`
	Slug     string `json:"slug,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	Type     string `json:"type,omitempty"`
}

// TaxonomyBody represents the common fields for creating and updating a taxonomy.
type TaxonomyBody struct {
	Name        string      `json:"name,omitempty"`
	Type        string      `json:"type,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Cover       string      `json:"cover,omitempty"`
	Thumbnail   string      `json:"thumbnail,omitempty"`
	Color       string      `json:"color,omitempty"`
	Icon        string      `json:"icon,omitempty"`
	URL         string      `json:"url,omitempty"`
	Keywords    string      `json:"keywords,omitempty"`
	Description string      `json:"description,omitempty"`
	Status      int         `json:"status,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id,omitempty"`
	TenantID    string      `json:"tenant_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateTaxonomyBody represents the body for creating a taxonomy.
type CreateTaxonomyBody struct {
	TaxonomyBody
}

// UpdateTaxonomyBody represents the body for updating a taxonomy.
type UpdateTaxonomyBody struct {
	ID string `json:"id"`
	TaxonomyBody
}

// ReadTaxonomy represents the output schema for retrieving a taxonomy.
type ReadTaxonomy struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Slug        string      `json:"slug"`
	Cover       string      `json:"cover"`
	Thumbnail   string      `json:"thumbnail"`
	Color       string      `json:"color"`
	Icon        string      `json:"icon"`
	URL         string      `json:"url"`
	Keywords    string      `json:"keywords"`
	Description string      `json:"description"`
	Status      int         `json:"status"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id"`
	TenantID    string      `json:"tenant_id"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

func (r *ReadTaxonomy) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// ListTaxonomyParams represents the query parameters for listing taxonomies.
type ListTaxonomyParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset    int    `form:"offset,omitempty" json:"offset,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	ParentID  string `form:"parent_id,omitempty" json:"parent_id,omitempty"`
	TenantID  string `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty" validate:"required"`
}
