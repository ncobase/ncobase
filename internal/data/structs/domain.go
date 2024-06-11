package structs

import (
	"stocms/pkg/types"
)

// Domain represents common fields for a domain.
type Domain struct {
	Name        string      `json:"name,omitempty"`
	Title       string      `json:"title,omitempty"`
	URL         string      `json:"url,omitempty"`
	Logo        string      `json:"logo,omitempty"`
	LogoAlt     string      `json:"logo_alt,omitempty"`
	Keywords    []string    `json:"keywords,omitempty"`
	Copyright   string      `json:"copyright,omitempty"`
	Description string      `json:"description,omitempty"`
	Order       *int32      `json:"order,omitempty"` // Use pointer for nullable int
	Disabled    bool        `json:"disabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   string      `json:"created_by,omitempty"`
}

// CreateDomainBody represents the body for creating a domain.
type CreateDomainBody struct {
	Domain
}

// UpdateDomainBody represents the body for updating a domain.
type UpdateDomainBody struct {
	Domain
	ID string `json:"id"`
}

// GetDomain represents the output schema for retrieving a domain.
type GetDomain struct {
	ID string `json:"id"`
	Domain
	User *User `json:"user,omitempty"`
}

// ListDomainParams represents the query parameters for listing domains.
type ListDomainParams struct {
	Cursor string `form:"cursor" json:"cursor,omitempty"`
	Limit  int32  `form:"limit" json:"limit,omitempty"`
	User   string `form:"user,omitempty" json:"user,omitempty"`
}

// FindDomain represents the parameters for finding a domain.
type FindDomain struct {
	ID   string `json:"id,omitempty"`
	User string `json:"user,omitempty"`
}
