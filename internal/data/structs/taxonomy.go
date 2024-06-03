package structs

import (
	"stocms/pkg/types"
	"time"
)

// FindTaxonomy represents the parameters for finding a taxonomy.
type FindTaxonomy struct {
	ID       string `json:"id,omitempty"`
	Slug     string `json:"slug,omitempty"`
	DomainID string `json:"domain_id,omitempty"`
	Type     string `json:"type,omitempty"`
}

// TaxonomyBody - Common fields for creating and updating taxonomies
type TaxonomyBody struct {
	Name        string     `json:"name,omitempty"`
	Type        string     `json:"type,omitempty"` // type, default 'node', options: 'node', 'plane', 'event', 'page', 'tag', 'link'
	Slug        string     `json:"slug,omitempty"`
	Cover       string     `json:"cover,omitempty"`
	Thumbnail   string     `json:"thumbnail,omitempty"`
	Color       string     `json:"color,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	URL         string     `json:"url,omitempty"`
	Keywords    []string   `json:"keywords,omitempty"`
	Description string     `json:"description,omitempty"`
	Status      int32      `json:"status,omitempty"` // status, 0: enabled, 1: disabled, ...
	ExtraProps  types.JSON `json:"extra_props,omitempty"`
	ParentID    string     `json:"parent_id,omitempty"`
	DomainID    string     `json:"domain_id,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
}

// CreateTaxonomyBody - Create taxonomy body
type CreateTaxonomyBody struct {
	TaxonomyBody
}

// UpdateTaxonomyBody - Update taxonomy body
type UpdateTaxonomyBody struct {
	TaxonomyBody
	ID string `json:"id"`
}

// ReadTaxonomy - Output taxonomy schema
type ReadTaxonomy struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Slug        string     `json:"slug"`
	Cover       string     `json:"cover"`
	Thumbnail   string     `json:"thumbnail"`
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
	URL         string     `json:"url"`
	Keywords    []string   `json:"keywords"`
	Description string     `json:"description"`
	Status      int32      `json:"status"`
	ExtraProps  types.JSON `json:"extra_props"`
	ParentID    string     `json:"parent_id"`
	DomainID    string     `json:"domain_id"`
	CreatedBy   string     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedBy   string     `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
}

// ListTaxonomyParams - Query taxonomy list params
type ListTaxonomyParams struct {
	Cursor string `form:"cursor" json:"cursor"`
	Limit  int64  `form:"limit" json:"limit"`
	Parent string `form:"parent,omitempty" json:"parent,omitempty"`
	Domain string `form:"domain,omitempty" json:"domain,omitempty"`
	Type   string `form:"type,omitempty" json:"type,omitempty" binding:"required"`
}
