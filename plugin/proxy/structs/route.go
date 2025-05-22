package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// RouteBody defines the structure for recording proxy routes.
type RouteBody struct {
	Name                string      `json:"name" validate:"required"`
	Description         string      `json:"description"`
	EndpointID          string      `json:"endpoint_id" validate:"required"`
	PathPattern         string      `json:"path_pattern" validate:"required"`
	TargetPath          string      `json:"target_path" validate:"required"`
	Method              string      `json:"method" validate:"required"`
	InputTransformerID  *string     `json:"input_transformer_id,omitempty"`
	OutputTransformerID *string     `json:"output_transformer_id,omitempty"`
	CacheEnabled        bool        `json:"cache_enabled"`
	CacheTTL            int         `json:"cache_ttl"`
	RateLimit           *string     `json:"rate_limit,omitempty"`
	StripAuthHeader     bool        `json:"strip_auth_header"`
	Disabled            bool        `json:"disabled"`
	Extras              *types.JSON `json:"extras,omitempty"`
	CreatedBy           *string     `json:"created_by,omitempty"`
	UpdatedBy           *string     `json:"updated_by,omitempty"`
}

// CreateRouteBody represents the body for creating a proxy route.
type CreateRouteBody struct {
	RouteBody
}

// UpdateRouteBody represents the body for updating a proxy route.
type UpdateRouteBody struct {
	ID string `json:"id,omitempty"`
	RouteBody
}

// ReadRoute represents the output schema for retrieving a proxy route.
type ReadRoute struct {
	ID                  string      `json:"id"`
	Name                string      `json:"name"`
	Description         string      `json:"description"`
	EndpointID          string      `json:"endpoint_id"`
	PathPattern         string      `json:"path_pattern"`
	TargetPath          string      `json:"target_path"`
	Method              string      `json:"method"`
	InputTransformerID  *string     `json:"input_transformer_id,omitempty"`
	OutputTransformerID *string     `json:"output_transformer_id,omitempty"`
	CacheEnabled        bool        `json:"cache_enabled"`
	CacheTTL            int         `json:"cache_ttl"`
	RateLimit           *string     `json:"rate_limit,omitempty"`
	StripAuthHeader     bool        `json:"strip_auth_header"`
	Disabled            bool        `json:"disabled"`
	Extras              *types.JSON `json:"extras,omitempty"`
	CreatedBy           *string     `json:"created_by,omitempty"`
	CreatedAt           *int64      `json:"created_at,omitempty"`
	UpdatedBy           *string     `json:"updated_by,omitempty"`
	UpdatedAt           *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadRoute) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListRouteParams represents the query parameters for listing proxy routes.
type ListRouteParams struct {
	Name       string `form:"name,omitempty" json:"name,omitempty"`
	EndpointID string `form:"endpoint_id,omitempty" json:"endpoint_id,omitempty"`
	Method     string `form:"method,omitempty" json:"method,omitempty"`
	Disabled   *bool  `form:"disabled,omitempty" json:"disabled,omitempty"`
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string `form:"direction,omitempty" json:"direction,omitempty"`
}
