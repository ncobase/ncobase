package structs

import (
	"time"

	"ncobase/common/types"
)

// ModuleBody represents a module entity.
type ModuleBody struct {
	Name      string     `json:"name,omitempty"`
	Title     string     `json:"title,omitempty"`
	Slug      string     `json:"slug,omitempty"`
	Content   string     `json:"content,omitempty"`
	Thumbnail string     `json:"thumbnail,omitempty"`
	Temp      *bool      `json:"temp,omitempty"`
	Markdown  *bool      `json:"markdown,omitempty"`
	Private   *bool      `json:"private,omitempty"`
	Status    *int       `json:"status,omitempty"`
	Released  *time.Time `json:"released,omitempty"`
	BaseEntity
}

// CreateModuleBody represents the body for creating or updating a module.
type CreateModuleBody struct {
	ModuleBody
}

// UpdateModuleBody represents the body for updating a module.
type UpdateModuleBody struct {
	ID string `json:"id,omitempty"`
	ModuleBody
}

// ReadModule represents the output schema for retrieving a module.
type ReadModule struct {
	ID          string      `json:"id"`
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Thumbnail   string      `json:"thumbnail"`
	Temp        bool        `json:"temp"`
	Markdown    bool        `json:"markdown"`
	Private     bool        `json:"private"`
	Status      int         `json:"status"`
	Released    time.Time   `json:"released"`
	Keywords    []string    `json:"keywords"`
	Description string      `json:"description"`
	Extras      *types.JSON `json:"extras,omitempty"`
	BaseEntity
}

// FindModule represents the parameters for finding a module.
type FindModule struct {
	ID   string `form:"id,omitempty" json:"id,omitempty"`
	Slug string `form:"slug,omitempty" json:"slug,omitempty"`
}

// ListModuleParams represents the query parameters for listing modules.
type ListModuleParams struct {
	Cursor string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
}
