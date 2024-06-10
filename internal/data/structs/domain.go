package structs

import "stocms/pkg/types"

// FindDomain represents the parameters for finding a domain.
type FindDomain struct {
	ID   string `json:"id,omitempty"`
	User string `json:"user,omitempty"`
}

// DomainBody - Common fields for creating and updating domains
type DomainBody struct {
	Name        string     `json:"name,omitempty"`
	Title       string     `json:"title,omitempty"`
	URL         string     `json:"url,omitempty"`
	Logo        string     `json:"logo,omitempty"`
	LogoAlt     string     `json:"logo_alt,omitempty"`
	Keywords    []string   `json:"keywords,omitempty"`
	Copyright   string     `json:"copyright,omitempty"`
	Description string     `json:"description,omitempty"`
	Order       int64      `json:"order,omitempty"`
	Disabled    bool       `json:"disabled,omitempty"`
	Extras      types.JSON `json:"extras,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
}

// CreateDomainBody - Create domain body
type CreateDomainBody struct {
	DomainBody
}

// UpdateDomainBody - Update domain body
type UpdateDomainBody struct {
	DomainBody
	ID string `json:"id"`
}

// ReadDomain - Output domain schema
type ReadDomain struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Logo        string     `json:"logo"`
	LogoAlt     string     `json:"logo_alt,omitempty"`
	Keywords    []string   `json:"keywords"`
	Copyright   string     `json:"copyright"`
	Description string     `json:"description"`
	Order       int32      `json:"order"`
	Disabled    bool       `json:"disabled"`
	Extras      types.JSON `json:"extras,omitempty"`
	User        *User      `json:"user,omitempty"`
}

// ListDomainParams - Query domain list params
type ListDomainParams struct {
	Cursor string `form:"cursor" json:"cursor"`
	Limit  int32  `form:"limit" json:"limit"`
	User   string `form:"user,omitempty" json:"user,omitempty"`
}
