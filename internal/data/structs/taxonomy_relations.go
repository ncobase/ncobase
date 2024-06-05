package structs

import "time"

// CreateTaxonomyRelationsBody represents the request body for creating a taxonomy relation.
type CreateTaxonomyRelationsBody struct {
	TaxonomyID string    `json:"taxonomy_id"`
	Type       string    `json:"type"`
	ObjectID   string    `json:"object_id"`
	Order      int32     `json:"order"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// UpdateTaxonomyRelationsBody represents the request body for updating a taxonomy relation.
type UpdateTaxonomyRelationsBody struct {
	ObjectID   string    `json:"object_id"`
	TaxonomyID string    `json:"taxonomy_id"`
	Type       string    `json:"type"`
	Order      int32     `json:"order"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListTaxonomyRelationsParams represents the parameters for listing taxonomy relations.
type ListTaxonomyRelationsParams struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
	Domain string `json:"domain"`
}

// FindTaxonomyRelations represents the parameters for finding a single taxonomy relation.
type FindTaxonomyRelations struct {
	Object   string `json:"object"`
	Taxonomy string `json:"taxonomy"`
	Type     string `json:"type"`
}

// FindTaxonomyRelationsParams represents the parameters for finding multiple taxonomy relations.
type FindTaxonomyRelationsParams struct {
	Object   string `json:"object"`
	Taxonomy string `json:"taxonomy"`
	Type     string `json:"type"`
}
