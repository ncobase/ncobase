package structs

import (
	"io"
	"stocms/pkg/types"
	"time"
)

// FindResource represents the parameters for finding an resource.
type FindResource struct {
	ID     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
	User   string `json:"user,omitempty"`
}

// ResourceBody represents the common fields for creating and updating an resource.
type ResourceBody struct {
	File       io.Reader  `json:"-"`
	Name       string     `json:"name,omitempty"`
	Path       string     `json:"path,omitempty"`
	Type       string     `json:"type,omitempty"`
	Size       int64      `json:"size,omitempty"`
	Storage    string     `json:"storage,omitempty"`
	ObjectID   string     `json:"object_id,omitempty"`
	DomainID   string     `json:"domain_id,omitempty"`
	ExtraProps types.JSON `json:"extras,omitempty"`
	CreatedBy  string     `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at,omitempty"`
	UpdatedBy  string     `json:"updated_by,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at,omitempty"`
}

// CreateResourceBody represents the body for creating an resource.
type CreateResourceBody struct {
	ResourceBody
}

// UpdateResourceBody represents the body for updating an resource.
type UpdateResourceBody struct {
	ID string `json:"id"`
	ResourceBody
}

// ListResourceParams represents the parameters for listing resources.
type ListResourceParams struct {
	Cursor  string `form:"cursor" json:"cursor"`
	Limit   int64  `form:"limit" json:"limit"`
	Domain  string `form:"domain,omitempty" json:"domain,omitempty"`
	User    string `form:"user,omitempty" json:"user,omitempty"`
	Type    string `form:"type,omitempty" json:"type,omitempty"`
	Storage string `form:"storage,omitempty" json:"storage,omitempty"`
}
