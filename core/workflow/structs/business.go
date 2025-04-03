package structs

import (
	"fmt"
	"ncobase/ncore/types"
)

// BusinessBody represents a business entity base fields
type BusinessBody struct {
	Code         string            `json:"code,omitempty"`
	Status       string            `json:"status,omitempty"`
	ModuleCode   string            `json:"module_code,omitempty"`
	FormCode     string            `json:"form_code,omitempty"`
	FormVersion  string            `json:"form_version,omitempty"`
	ProcessID    string            `json:"process_id,omitempty"`
	TemplateID   string            `json:"template_id,omitempty"`
	FlowStatus   string            `json:"flow_status,omitempty"`
	OriginData   types.JSON        `json:"origin_data,omitempty"`
	CurrentData  types.JSON        `json:"current_data,omitempty"`
	Variables    types.JSON        `json:"variables,omitempty"`
	IsDraft      bool              `json:"is_draft,omitempty"`
	BusinessTags types.StringArray `json:"business_tags,omitempty"`
	Viewers      types.StringArray `json:"viewers,omitempty"`
	Editors      types.StringArray `json:"editors,omitempty"`
	TenantID     string            `json:"tenant_id,omitempty"`
	Extras       types.JSON        `json:"extras,omitempty"`
}

// CreateBusinessBody represents the body for creating business
type CreateBusinessBody struct {
	BusinessBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateBusinessBody represents the body for updating business
type UpdateBusinessBody struct {
	ID string `json:"id,omitempty"`
	BusinessBody
}

// ReadBusiness represents the output schema for retrieving business
type ReadBusiness struct {
	ID           string            `json:"id"`
	Code         string            `json:"code,omitempty"`
	Status       string            `json:"status,omitempty"`
	ModuleCode   string            `json:"module_code,omitempty"`
	FormCode     string            `json:"form_code,omitempty"`
	FormVersion  string            `json:"form_version,omitempty"`
	ProcessID    string            `json:"process_id,omitempty"`
	TemplateID   string            `json:"template_id,omitempty"`
	FlowStatus   string            `json:"flow_status,omitempty"`
	OriginData   types.JSON        `json:"origin_data,omitempty"`
	CurrentData  types.JSON        `json:"current_data,omitempty"`
	Variables    types.JSON        `json:"variables,omitempty"`
	IsDraft      bool              `json:"is_draft,omitempty"`
	BusinessTags types.StringArray `json:"business_tags,omitempty"`
	Viewers      types.StringArray `json:"viewers,omitempty"`
	Editors      types.StringArray `json:"editors,omitempty"`
	TenantID     string            `json:"tenant_id,omitempty"`
	Extras       types.JSON        `json:"extras,omitempty"`
	CreatedBy    *string           `json:"created_by,omitempty"`
	CreatedAt    *int64            `json:"created_at,omitempty"`
	UpdatedBy    *string           `json:"updated_by,omitempty"`
	UpdatedAt    *int64            `json:"updated_at,omitempty"`
	LastModified *int64            `json:"last_modified,omitempty"`
	LastModifier *string           `json:"last_modifier,omitempty"`
}

// GetID returns the ID of the business
func (r *ReadBusiness) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadBusiness) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadBusiness) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindBusinessParams represents query parameters for finding business records
type FindBusinessParams struct {
	ProcessID  string `form:"process_id,omitempty" json:"process_id,omitempty"`
	ModuleCode string `form:"module_code,omitempty" json:"module_code,omitempty"`
	FormCode   string `form:"form_code,omitempty" json:"form_code,omitempty"`
	Code       string `form:"code,omitempty" json:"code,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	FlowStatus string `form:"flow_status,omitempty" json:"flow_status,omitempty"`
	IsDraft    *bool  `form:"is_draft,omitempty" json:"is_draft,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListBusinessParams represents list parameters for business records
type ListBusinessParams struct {
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string `form:"direction,omitempty" json:"direction,omitempty"`
	ModuleCode string `form:"module_code,omitempty" json:"module_code,omitempty"`
	FormCode   string `form:"form_code,omitempty" json:"form_code,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	IsDraft    *bool  `form:"is_draft,omitempty" json:"is_draft,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
