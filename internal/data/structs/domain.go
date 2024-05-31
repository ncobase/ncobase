package structs

import "stocms/pkg/types"

// CreateDomainBody - Create domain body
type CreateDomainBody struct {
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
	UserID      string     `json:"user_id"`
}

// UpdateDomainBody - Update domain body
type UpdateDomainBody struct {
	ID          string     `json:"id"`
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
	UserID      string     `json:"user_id,omitempty"`
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
	Order       int64      `json:"order"`
	Disabled    bool       `json:"disabled"`
	Extras      types.JSON `json:"extras,omitempty"`
	User        ReadUser   `json:"user,omitempty"`
}

// ListDomainParams - Query domain list params
type ListDomainParams struct {
	Cursor string `json:"cursor"`
	Limit  int64  `json:"limit"`
	UserID string `json:"uid"`
}
