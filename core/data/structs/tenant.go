package structs

import (
	"ncobase/common/types"
	"time"
)

// TenantBody represents common fields for a tenant.
type TenantBody struct {
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
	ExpiredAt   *time.Time  `json:"expired_at,omitempty"`
	OperatorBy
}

// CreateTenantBody represents the body for creating a tenant.
type CreateTenantBody struct {
	TenantBody
}

// UpdateTenantBody represents the body for updating a tenant.
type UpdateTenantBody struct {
	ID string `json:"id"`
	TenantBody
}

// ReadTenant represents the output schema for retrieving a tenant.
type ReadTenant struct {
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
	ExpiredAt   *time.Time  `json:"expired_at"`
	BaseEntity
}

// FindTenant represents the parameters for finding a tenant.
type FindTenant struct {
	Slug string `json:"slug,omitempty"`
	User string `json:"user,omitempty"`
}

// ListTenantParams represents the query parameters for listing tenants.
type ListTenantParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
	User   string `form:"user,omitempty" json:"user,omitempty"`
}
