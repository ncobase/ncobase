package structs

import (
	"fmt"

	"github.com/ncobase/ncore/utils/convert"
)

// TaxonomyRelationBody represents the common fields for creating and updating a taxonomy relation.
type TaxonomyRelationBody struct {
	TaxonomyID string  `json:"taxonomy_id,omitempty"`
	Type       string  `json:"type,omitempty"`
	ObjectID   string  `json:"object_id,omitempty"`
	Order      *int    `json:"order,omitempty"`
	CreatedBy  *string `json:"created_by,omitempty"`
	CreatedAt  *int64  `json:"created_at,omitempty"`
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
	ID         string  `json:"id"`
	ObjectID   string  `json:"object_id"`
	TaxonomyID string  `json:"taxonomy_id"`
	Type       string  `json:"type"`
	Order      *int    `json:"order"`
	CreatedBy  *string `json:"created_by,omitempty"`
	CreatedAt  *int64  `json:"created_at,omitempty"`
	UpdatedBy  *string `json:"updated_by,omitempty"`
	UpdatedAt  *int64  `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadTaxonomyRelation) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListTaxonomyRelationParams represents the parameters for listing taxonomy relations.
type ListTaxonomyRelationParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	SpaceID   string `json:"space_id,omitempty"`
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
