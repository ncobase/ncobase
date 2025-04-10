package structs

import (
	"fmt"

	"github.com/ncobase/ncore/pkg/types"
)

// TemplateBody represents a template entity base fields
type TemplateBody struct {
	DefaultTitle   string     `json:"default_title,omitempty"`
	Name           string     `json:"name,omitempty"`
	Code           string     `json:"code,omitempty"`
	Description    string     `json:"description,omitempty"`
	Type           string     `json:"type,omitempty"`
	Version        string     `json:"version,omitempty"`
	Status         string     `json:"status,omitempty"`
	ModuleCode     string     `json:"module_code,omitempty"`
	FormCode       string     `json:"form_code,omitempty"`
	TemplateKey    string     `json:"template_key,omitempty"`
	Category       string     `json:"category,omitempty"`
	NodeConfig     types.JSON `json:"node_config,omitempty"`
	NodeRules      types.JSON `json:"node_rules,omitempty"`
	NodeEvents     types.JSON `json:"node_events,omitempty"`
	ProcessRules   types.JSON `json:"process_rules,omitempty"`
	FormConfig     types.JSON `json:"form_config,omitempty"`
	FormPerms      types.JSON `json:"form_permissions,omitempty"`
	RoleConfigs    types.JSON `json:"role_configs,omitempty"`
	PermConfigs    types.JSON `json:"permission_configs,omitempty"`
	VisibleRange   types.JSON `json:"visible_range,omitempty"`
	IsDraftEnabled bool       `json:"is_draft_enabled,omitempty"`
	IsAutoStart    bool       `json:"is_auto_start,omitempty"`
	StrictMode     bool       `json:"strict_mode,omitempty"`
	AllowCancel    bool       `json:"allow_cancel,omitempty"`
	AllowUrge      bool       `json:"allow_urge,omitempty"`
	AllowDelegate  bool       `json:"allow_delegate,omitempty"`
	AllowTransfer  bool       `json:"allow_transfer,omitempty"`
	TimeoutConfig  types.JSON `json:"timeout_config,omitempty"`
	ReminderConfig types.JSON `json:"reminder_config,omitempty"`
	SourceVersion  string     `json:"source_version,omitempty"`
	IsLatest       bool       `json:"is_latest,omitempty"`
	Disabled       bool       `json:"disabled,omitempty"`
	TenantID       string     `json:"tenant_id,omitempty"`
	EffectiveTime  *int64     `json:"effective_time,omitempty"`
	ExpireTime     *int64     `json:"expire_time,omitempty"`
	Extras         types.JSON `json:"extras,omitempty"`
}

// CreateTemplateBody represents the body for creating template
type CreateTemplateBody struct {
	TemplateBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateTemplateBody represents the body for updating template
type UpdateTemplateBody struct {
	ID string `json:"id,omitempty"`
	TemplateBody
}

// ReadTemplate represents the output schema for retrieving template
type ReadTemplate struct {
	ID             string     `json:"id"`
	DefaultTitle   string     `json:"default_title,omitempty"`
	Name           string     `json:"name,omitempty"`
	Code           string     `json:"code,omitempty"`
	Description    string     `json:"description,omitempty"`
	Type           string     `json:"type,omitempty"`
	Version        string     `json:"version,omitempty"`
	Status         string     `json:"status,omitempty"`
	ModuleCode     string     `json:"module_code,omitempty"`
	FormCode       string     `json:"form_code,omitempty"`
	TemplateKey    string     `json:"template_key,omitempty"`
	Category       string     `json:"category,omitempty"`
	NodeConfig     types.JSON `json:"node_config,omitempty"`
	NodeRules      types.JSON `json:"node_rules,omitempty"`
	NodeEvents     types.JSON `json:"node_events,omitempty"`
	ProcessRules   types.JSON `json:"process_rules,omitempty"`
	FormConfig     types.JSON `json:"form_config,omitempty"`
	FormPerms      types.JSON `json:"form_permissions,omitempty"`
	RoleConfigs    types.JSON `json:"role_configs,omitempty"`
	PermConfigs    types.JSON `json:"permission_configs,omitempty"`
	VisibleRange   types.JSON `json:"visible_range,omitempty"`
	IsDraftEnabled bool       `json:"is_draft_enabled,omitempty"`
	IsAutoStart    bool       `json:"is_auto_start,omitempty"`
	StrictMode     bool       `json:"strict_mode,omitempty"`
	AllowCancel    bool       `json:"allow_cancel,omitempty"`
	AllowUrge      bool       `json:"allow_urge,omitempty"`
	AllowDelegate  bool       `json:"allow_delegate,omitempty"`
	AllowTransfer  bool       `json:"allow_transfer,omitempty"`
	TimeoutConfig  types.JSON `json:"timeout_config,omitempty"`
	ReminderConfig types.JSON `json:"reminder_config,omitempty"`
	SourceVersion  string     `json:"source_version,omitempty"`
	IsLatest       bool       `json:"is_latest,omitempty"`
	Disabled       bool       `json:"disabled,omitempty"`
	TenantID       string     `json:"tenant_id,omitempty"`
	EffectiveTime  *int64     `json:"effective_time,omitempty"`
	ExpireTime     *int64     `json:"expire_time,omitempty"`
	Extras         types.JSON `json:"extras,omitempty"`
	CreatedBy      *string    `json:"created_by,omitempty"`
	CreatedAt      *int64     `json:"created_at,omitempty"`
	UpdatedBy      *string    `json:"updated_by,omitempty"`
	UpdatedAt      *int64     `json:"updated_at,omitempty"`
}

// GetID returns the ID of the template
func (r *ReadTemplate) GetID() string {
	return r.ID
}

// GetCursorValue returns the cursor value
func (r *ReadTemplate) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// GetSortValue get sort value
func (r *ReadTemplate) GetSortValue(field string) any {
	switch field {
	case SortByCreatedAt:
		return types.ToValue(r.CreatedAt)
	case SortByName:
		return r.Name
	default:
		return types.ToValue(r.CreatedAt)
	}
}

// FindTemplateParams represents query parameters for finding templates
type FindTemplateParams struct {
	ID         string `form:"id,omitempty" json:"id,omitempty"`
	Code       string `form:"code,omitempty" json:"code,omitempty"`
	ModuleCode string `form:"module_code,omitempty" json:"module_code,omitempty"`
	FormCode   string `form:"form_code,omitempty" json:"form_code,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Version    string `form:"version,omitempty" json:"version,omitempty"`
	Category   string `form:"category,omitempty" json:"category,omitempty"`
	IsLatest   *bool  `form:"is_latest,omitempty" json:"is_latest,omitempty"`
	Disabled   *bool  `form:"disabled,omitempty" json:"disabled,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}

// ListTemplateParams represents list parameters for templates
type ListTemplateParams struct {
	Code       string `form:"code,omitempty" json:"code,omitempty"`
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string `form:"direction,omitempty" json:"direction,omitempty"`
	ModuleCode string `form:"module_code,omitempty" json:"module_code,omitempty"`
	FormCode   string `form:"form_code,omitempty" json:"form_code,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Category   string `form:"category,omitempty" json:"category,omitempty"`
	IsLatest   *bool  `form:"is_latest,omitempty" json:"is_latest,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
	SortBy     string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
}
