package structs

import (
	"time"

	"github.com/ncobase/common/types"
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
	Keywords    []string    `json:"keywords,omitempty"`
	Description string      `json:"description,omitempty"`
	Status      int         `json:"status,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id,omitempty"`
	TenantID    string      `json:"tenant_id,omitempty"`
	BaseEntity
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
	Keywords    []string    `json:"keywords"`
	Description string      `json:"description"`
	Status      int         `json:"status"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ParentID    string      `json:"parent_id"`
	TenantID    string      `json:"tenant_id"`
	BaseEntity
}

// ListTaxonomyParams represents the query parameters for listing taxonomies.
type ListTaxonomyParams struct {
	Cursor   string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit    int    `form:"limit,omitempty" json:"limit,omitempty"`
	ParentID string `form:"parent_id,omitempty" json:"parent_id,omitempty"`
	TenantID string `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty" validate:"required"`
}

// TaxonomyRelationBody represents the common fields for creating and updating a taxonomy relation.
type TaxonomyRelationBody struct {
	TaxonomyID string     `json:"taxonomy_id,omitempty"`
	Type       string     `json:"type,omitempty"`
	ObjectID   string     `json:"object_id,omitempty"`
	Order      *int       `json:"order,omitempty"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

// CreateTaxonomyRelationBody represents the request body for creating a taxonomy relation.
type CreateTaxonomyRelationBody struct {
	TaxonomyRelationBody
}

// UpdateTaxonomyRelationBody represents the request body for updating a taxonomy relation.
type UpdateTaxonomyRelationBody struct {
	ID string `json:"id"`
	TaxonomyRelationBody
}

// ListTaxonomyRelationParams represents the parameters for listing taxonomy relations.
type ListTaxonomyRelationParams struct {
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
}

// FindTaxonomyRelation represents the parameters for finding a single taxonomy relation.
type FindTaxonomyRelation struct {
	ObjectID   string `json:"object_id,omitempty"`
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	Type       string `json:"type,omitempty"`
}

// FindTaxonomyRelationParams represents the parameters for finding multiple taxonomy relations.
type FindTaxonomyRelationParams struct {
	ObjectID   string `json:"object_id,omitempty"`
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	Type       string `json:"type,omitempty"`
}
