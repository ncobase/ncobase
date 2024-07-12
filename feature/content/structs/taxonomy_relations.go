package structs

import "time"

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

// ReadTaxonomyRelation represents the output schema for retrieving a taxonomy relation.
type ReadTaxonomyRelation struct {
	ID         string     `json:"id"`
	ObjectID   string     `json:"object_id"`
	TaxonomyID string     `json:"taxonomy_id"`
	Type       string     `json:"type"`
	Order      *int       `json:"order"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	UpdatedBy  *string    `json:"updated_by,omitempty"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
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
