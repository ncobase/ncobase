package structs

import (
	"io"
	"stocms/pkg/types"
)

// FindResource represents the parameters for finding a resource.
type FindResource struct {
	ID     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
	User   string `json:"user,omitempty"`
}

// ResourceBody represents the common fields for creating and updating a resource.
type ResourceBody struct {
	File     io.Reader   `json:"-"`
	Name     string      `json:"name,omitempty"`
	Path     string      `json:"path,omitempty"`
	Type     string      `json:"type,omitempty"`
	Size     *int64      `json:"size,omitempty"`
	Storage  string      `json:"storage,omitempty"`
	URL      string      `json:"url,omitempty"`
	ObjectID string      `json:"object_id,omitempty"`
	DomainID string      `json:"domain_id,omitempty"`
	Extras   *types.JSON `json:"extras,omitempty"`
	BaseEntity
}

// CreateResourceBody represents the body for creating a resource.
type CreateResourceBody struct {
	ResourceBody
}

// UpdateResourceBody represents the body for updating a resource.
type UpdateResourceBody struct {
	ID string `json:"id"`
	ResourceBody
}

// ListResourceParams represents the parameters for listing resources.
type ListResourceParams struct {
	Cursor  string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit   int64  `form:"limit,omitempty" json:"limit,omitempty"`
	Domain  string `form:"domain,omitempty" json:"domain,omitempty"`
	Object  string `form:"object,omitempty" json:"object,omitempty"`
	User    string `form:"user,omitempty" json:"user,omitempty"`
	Type    string `form:"type,omitempty" json:"type,omitempty"`
	Storage string `form:"storage,omitempty" json:"storage,omitempty"`
}
