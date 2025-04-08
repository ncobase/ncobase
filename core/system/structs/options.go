package structs

import (
	"fmt"
	"ncore/pkg/types"
)

// OptionsBody represents an options entity.
type OptionsBody struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
	Autoload bool   `json:"autoload,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
}

// CreateOptionsBody represents the body for creating options.
type CreateOptionsBody struct {
	OptionsBody
}

// UpdateOptionsBody represents the body for updating options.
type UpdateOptionsBody struct {
	ID string `json:"id,omitempty"`
	OptionsBody
}

// ReadOptions represents the output schema for retrieving options.
type ReadOptions struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Value     string  `json:"value"`
	Autoload  bool    `json:"autoload"`
	TenantID  string  `json:"tenant_id,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	CreatedAt *int64  `json:"created_at,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
	UpdatedAt *int64  `json:"updated_at,omitempty"`
}

// GetID returns the ID of the options.
func (r *ReadOptions) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value.
func (r *ReadOptions) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadOptions) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	case SortByName:
		return r.Name
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindOptions represents the parameters for finding options.
type FindOptions struct {
	Option string `form:"option,omitempty" json:"option,omitempty"`
	Tenant string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Type   string `form:"type,omitempty" json:"type,omitempty"`
	SortBy string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListOptionsParams represents the query parameters for listing options.
type ListOptionsParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Autoload  *bool  `form:"autoload,omitempty" json:"autoload,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
