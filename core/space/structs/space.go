package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// SpaceBody represents common fields for a space.
type SpaceBody struct {
	Name        string      `json:"name,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Type        string      `json:"type,omitempty"`
	Title       string      `json:"title,omitempty"`
	URL         string      `json:"url,omitempty"`
	Logo        string      `json:"logo,omitempty"`
	LogoAlt     string      `json:"logo_alt,omitempty"`
	Keywords    string      `json:"keywords,omitempty"`
	Copyright   string      `json:"copyright,omitempty"`
	Description string      `json:"description,omitempty"`
	Order       *int        `json:"order,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ExpiredAt   *int64      `json:"expired_at,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateSpaceBody represents the body for creating a space.
type CreateSpaceBody struct {
	SpaceBody
}

// UpdateSpaceBody represents the body for updating a space.
type UpdateSpaceBody struct {
	ID string `json:"id"`
	SpaceBody
}

// ReadSpace represents the output schema for retrieving a space.
type ReadSpace struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	Logo        string      `json:"logo"`
	LogoAlt     string      `json:"logo_alt"`
	Keywords    string      `json:"keywords"`
	Copyright   string      `json:"copyright"`
	Description string      `json:"description"`
	Order       *int        `json:"order"`
	Disabled    bool        `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
	ExpiredAt   *int64      `json:"expired_at"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadSpace) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// FindSpace represents the parameters for finding a space.
type FindSpace struct {
	Slug string `json:"slug,omitempty"`
	User string `json:"user,omitempty"`
}

// ListSpaceParams represents the query parameters for listing spaces.
type ListSpaceParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	User      string `form:"user,omitempty" json:"user,omitempty"`
}
