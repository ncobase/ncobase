package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// ProcessDesignBody represents a process design entity base fields
type ProcessDesignBody struct {
	TemplateID      string     `json:"template_id,omitempty"`
	GraphData       types.JSON `json:"graph_data,omitempty"`
	NodeLayouts     types.JSON `json:"node_layouts,omitempty"`
	Properties      types.JSON `json:"properties,omitempty"`
	ValidationRules types.JSON `json:"validation_rules,omitempty"`
	IsDraft         bool       `json:"is_draft,omitempty"`
	Version         string     `json:"version,omitempty"`
	SourceVersion   string     `json:"source_version,omitempty"`
	SpaceID         string     `json:"space_id,omitempty"`
	Extras          types.JSON `json:"extras,omitempty"`
}

// CreateProcessDesignBody represents body for creating process design
type CreateProcessDesignBody struct {
	ProcessDesignBody
	SpaceID string `json:"space_id,omitempty"`
}

// UpdateProcessDesignBody represents body for updating process design
type UpdateProcessDesignBody struct {
	ID string `json:"id,omitempty"`
	ProcessDesignBody
}

// ReadProcessDesign represents output schema for retrieving process design
type ReadProcessDesign struct {
	ID              string     `json:"id"`
	TemplateID      string     `json:"template_id,omitempty"`
	GraphData       types.JSON `json:"graph_data,omitempty"`
	NodeLayouts     types.JSON `json:"node_layouts,omitempty"`
	Properties      types.JSON `json:"properties,omitempty"`
	ValidationRules types.JSON `json:"validation_rules,omitempty"`
	IsDraft         bool       `json:"is_draft,omitempty"`
	Version         string     `json:"version,omitempty"`
	SourceVersion   string     `json:"source_version,omitempty"`
	SpaceID         string     `json:"space_id,omitempty"`
	Extras          types.JSON `json:"extras,omitempty"`
	CreatedBy       *string    `json:"created_by,omitempty"`
	CreatedAt       *int64     `json:"created_at,omitempty"`
	UpdatedBy       *string    `json:"updated_by,omitempty"`
	UpdatedAt       *int64     `json:"updated_at,omitempty"`
}

// GetID returns ID of the process design
func (r *ReadProcessDesign) GetID() string {
	return r.ID
}

// GetCursorValue returns cursor value
func (r *ReadProcessDesign) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// FindProcessDesignParams represents query parameters for finding process designs
type FindProcessDesignParams struct {
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	Version    string `form:"version,omitempty" json:"version,omitempty"`
	IsDraft    *bool  `form:"is_draft,omitempty" json:"is_draft,omitempty"`
	Space      string `form:"space,omitempty" json:"space,omitempty"`
}

// ListProcessDesignParams represents list parameters for process designs
type ListProcessDesignParams struct {
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string `form:"direction,omitempty" json:"direction,omitempty"`
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	IsDraft    *bool  `form:"is_draft,omitempty" json:"is_draft,omitempty"`
	Space      string `form:"space,omitempty" json:"space,omitempty"`
}
