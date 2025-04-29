package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// TransformerBody defines the structure for request body used to create or update transformers.
type TransformerBody struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description"`
	Type        string      `json:"type" validate:"required,oneof=template script mapping"`
	Content     string      `json:"content" validate:"required"`
	ContentType string      `json:"content_type" validate:"required"`
	Disabled    bool        `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateTransformerBody represents the body for creating a transformer.
type CreateTransformerBody struct {
	TransformerBody
}

// UpdateTransformerBody represents the body for updating a transformer.
type UpdateTransformerBody struct {
	ID string `json:"id,omitempty"`
	TransformerBody
}

// ReadTransformer represents the output schema for retrieving a transformer.
type ReadTransformer struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Content     string      `json:"content"`
	ContentType string      `json:"content_type"`
	Disabled    bool        `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadTransformer) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListTransformerParams represents the query parameters for listing transformers.
type ListTransformerParams struct {
	Name        string `form:"name,omitempty" json:"name,omitempty"`
	Type        string `form:"type,omitempty" json:"type,omitempty"`
	ContentType string `form:"content_type,omitempty" json:"content_type,omitempty"`
	Disabled    *bool  `form:"disabled,omitempty" json:"disabled,omitempty"`
	Cursor      string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit       int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction   string `form:"direction,omitempty" json:"direction,omitempty"`
}
