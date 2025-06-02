package structs

import (
	"fmt"

	"github.com/ncobase/ncore/utils/convert"
)

// OptionBody represents an option entity.
type OptionBody struct {
	Name      string  `json:"name,omitempty"`
	Type      string  `json:"type,omitempty"`
	Value     string  `json:"value,omitempty"`
	Autoload  bool    `json:"autoload,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
}

// CreateOptionBody represents the body for creating option.
type CreateOptionBody struct {
	OptionBody
}

// UpdateOptionBody represents the body for updating option.
type UpdateOptionBody struct {
	ID string `json:"id,omitempty"`
	OptionBody
}

// ReadOption represents the output schema for retrieving option.
type ReadOption struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Value     string  `json:"value"`
	Autoload  bool    `json:"autoload"`
	CreatedBy *string `json:"created_by,omitempty"`
	CreatedAt *int64  `json:"created_at,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
	UpdatedAt *int64  `json:"updated_at,omitempty"`
}

// GetID returns the ID of the option.
func (r *ReadOption) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value.
func (r *ReadOption) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadOption) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return convert.ToValue(r.CreatedAt)
	case SortByName:
		return r.Name
	default:
		return convert.ToValue(r.CreatedAt)
	}
}

// FindOptions represents the parameters for finding option.
type FindOptions struct {
	Option string `form:"option,omitempty" json:"option,omitempty"`
	Type   string `form:"type,omitempty" json:"type,omitempty"`
	SortBy string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListOptionParams represents the query parameters for listing options.
type ListOptionParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Autoload  *bool  `form:"autoload,omitempty" json:"autoload,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
	Prefix    string `form:"prefix,omitempty" json:"prefix,omitempty"`
}
