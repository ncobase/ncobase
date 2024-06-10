package structs

import "time"

// CreateTaxonomyRelationBody represents the request body for creating a taxonomy relation.
type CreateTaxonomyRelationBody struct {
	TaxonomyID string    `json:"taxonomy_id"`
	Type       string    `json:"type"`
	ObjectID   string    `json:"object_id"`
	Order      int32     `json:"order"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpdateTaxonomyRelationBody represents the request body for updating a taxonomy relation.
type UpdateTaxonomyRelationBody struct {
	ObjectID   string    `json:"object_id"`
	TaxonomyID string    `json:"taxonomy_id"`
	Type       string    `json:"type"`
	Order      int32     `json:"order"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListTaxonomyRelationParams represents the parameters for listing taxonomy relations.
type ListTaxonomyRelationParams struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
	Domain string `json:"domain"`
}

// FindTaxonomyRelation represents the parameters for finding a single taxonomy relation.
type FindTaxonomyRelation struct {
	Object   string `json:"object"`
	Taxonomy string `json:"taxonomy"`
	Type     string `json:"type"`
}

// FindTaxonomyRelationParams represents the parameters for finding multiple taxonomy relations.
type FindTaxonomyRelationParams struct {
	Object   string `json:"object"`
	Taxonomy string `json:"taxonomy"`
	Type     string `json:"type"`
}
