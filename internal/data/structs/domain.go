package structs

import (
	"stocms/pkg/types"
	"time"
)

// DomainBody represents common fields for a domain.
type DomainBody struct {
	Name        string      `json:"name,omitempty"`
	Title       string      `json:"title,omitempty"`
	URL         string      `json:"url,omitempty"`
	Logo        string      `json:"logo,omitempty"`
	LogoAlt     string      `json:"logo_alt,omitempty"`
	Keywords    []string    `json:"keywords,omitempty"`
	Copyright   string      `json:"copyright,omitempty"`
	Description string      `json:"description,omitempty"`
	Order       *int32      `json:"order,omitempty"`
	Disabled    bool        `json:"disabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   string      `json:"created_by,omitempty"`
}

// CreateDomainBody represents the body for creating a domain.
type CreateDomainBody struct {
	DomainBody
}

// UpdateDomainBody represents the body for updating a domain.
type UpdateDomainBody struct {
	ID string `json:"id"`
	DomainBody
}

// ReadDomain represents the output schema for retrieving a domain.
type ReadDomain struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	Logo        string      `json:"logo"`
	LogoAlt     string      `json:"logo_alt"`
	Keywords    []string    `json:"keywords"`
	Copyright   string      `json:"copyright"`
	Description string      `json:"description"`
	Order       *int32      `json:"order"`
	Disabled    bool        `json:"disabled"`
	Extras      *types.JSON `json:"extras,omitempty"`
	User        *User       `json:"user,omitempty"`
	CreatedBy   string      `json:"created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// FindDomain represents the parameters for finding a domain.
type FindDomain struct {
	ID   string `json:"id,omitempty"`
	User string `json:"user,omitempty"`
}

// ListDomainParams represents the query parameters for listing domains.
type ListDomainParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int32  `form:"limit,omitempty" json:"limit,omitempty"`
	User   string `form:"user,omitempty" json:"user,omitempty"`
}

// Validate validates ListDomainParams
func (p *ListDomainParams) Validate() error {
	return validate.Struct(p)
}
