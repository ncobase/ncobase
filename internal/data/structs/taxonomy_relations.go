package structs

import "time"

// TaxonomyRelationBody represents the common fields for creating and updating a taxonomy relation.
type TaxonomyRelationBody struct {
	TaxonomyID string     `json:"taxonomy_id,omitempty"`
	Type       string     `json:"type,omitempty"`
	ObjectID   string     `json:"object_id,omitempty"`
	Order      *int32     `json:"order,omitempty"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
}

// CreateTaxonomyRelationBody represents the request body for creating a taxonomy relation.
type CreateTaxonomyRelationBody struct {
	TaxonomyRelationBody
}

// UpdateTaxonomyRelationBody represents the request body for updating a taxonomy relation.
type UpdateTaxonomyRelationBody struct {
	TaxonomyRelationBody
}

// ListTaxonomyRelationParams represents the parameters for listing taxonomy relations.
type ListTaxonomyRelationParams struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Domain string `json:"domain,omitempty"`
}

// FindTaxonomyRelation represents the parameters for finding a single taxonomy relation.
type FindTaxonomyRelation struct {
	Object   string `json:"object,omitempty"`
	Taxonomy string `json:"taxonomy,omitempty"`
	Type     string `json:"type,omitempty"`
}

// FindTaxonomyRelationParams represents the parameters for finding multiple taxonomy relations.
type FindTaxonomyRelationParams struct {
	Object   string `json:"object,omitempty"`
	Taxonomy string `json:"taxonomy,omitempty"`
	Type     string `json:"type,omitempty"`
}
