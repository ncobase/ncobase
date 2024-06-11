package structs

import "time"

// Module represents a module entity.
type Module struct {
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Title     string     `json:"title,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	Content   string     `json:"content,omitempty"`
	Thumbnail string     `json:"thumbnail,omitempty"`
	Temp      *bool      `json:"temp,omitempty"`
	Markdown  *bool      `json:"markdown,omitempty"`
	Private   *bool      `json:"private,omitempty"`
	Status    *int32     `json:"status,omitempty"`   // Use pointer for nullable field
	Released  *time.Time `json:"released,omitempty"` // Use pointer for nullable field
	BaseEntity
}

// CreateModuleBody represents the body for creating or updating a module.
type CreateModuleBody struct {
	Module
}

// UpdateModuleBody represents the body for updating a module.
type UpdateModuleBody struct {
	Module
	ID string `json:"id,omitempty"`
}

// FindModule represents the parameters for finding a module.
type FindModule struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListModuleParams represents the query parameters for listing modules.
type ListModuleParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int64  `form:"limit,omitempty" json:"limit,omitempty"`
}
