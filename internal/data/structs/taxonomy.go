package structs

import (
	"stocms/pkg/types"
)

// FindTaxonomy represents the parameters for finding a taxonomy.
type FindTaxonomy struct {
	ID       string `json:"id,omitempty"`
	Slug     string `json:"slug,omitempty"`
	DomainID string `json:"domain_id,omitempty"`
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
	Keywords    []string    `json:"keywords,omitempty"`
	Description string      `json:"description,omitempty"`
	Status      int32       `json:"status,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id,omitempty"`
	DomainID    string      `json:"domain_id,omitempty"`
	BaseEntity
}

// CreateTaxonomyBody represents the body for creating a taxonomy.
type CreateTaxonomyBody struct {
	TaxonomyBody
}

// UpdateTaxonomyBody represents the body for updating a taxonomy.
type UpdateTaxonomyBody struct {
	TaxonomyBody
	ID string `json:"id"`
}

// GetTaxonomy represents the output schema for retrieving a taxonomy.
type GetTaxonomy struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Slug        string      `json:"slug"`
	Cover       string      `json:"cover"`
	Thumbnail   string      `json:"thumbnail"`
	Color       string      `json:"color"`
	Icon        string      `json:"icon"`
	URL         string      `json:"url"`
	Keywords    []string    `json:"keywords"`
	Description string      `json:"description"`
	Status      int32       `json:"status"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id"`
	DomainID    string      `json:"domain_id"`
	BaseEntity
}

// ListTaxonomyParams represents the query parameters for listing taxonomies.
type ListTaxonomyParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty"`
	Parent string `form:"parent,omitempty" json:"parent,omitempty"`
	Domain string `form:"domain,omitempty" json:"domain,omitempty"`
	Type   string `form:"type,omitempty" json:"type,omitempty" binding:"required"`
}
