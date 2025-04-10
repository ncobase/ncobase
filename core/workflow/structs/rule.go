package structs

import (
	"fmt"
	"github.com/ncobase/ncore/pkg/types"
)

// RuleBody represents a rule entity base fields
type RuleBody struct {
	Name          string            `json:"name,omitempty"`
	Code          string            `json:"code,omitempty"`
	Description   string            `json:"description,omitempty"`
	Type          string            `json:"type,omitempty"`
	Status        string            `json:"status,omitempty"`
	RuleKey       string            `json:"rule_key,omitempty"`
	TemplateID    string            `json:"template_id,omitempty"`
	NodeKey       string            `json:"node_key,omitempty"`
	Conditions    types.StringArray `json:"conditions,omitempty"`
	Actions       types.JSON        `json:"actions,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	IsEnabled     bool              `json:"is_enabled,omitempty"`
	EffectiveTime *int64            `json:"effective_time,omitempty"`
	ExpireTime    *int64            `json:"expire_time,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	Extras        types.JSON        `json:"extras,omitempty"`
}

// CreateRuleBody represents body for creating rule
type CreateRuleBody struct {
	RuleBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateRuleBody represents body for updating rule
type UpdateRuleBody struct {
	ID string `json:"id,omitempty"`
	RuleBody
}

// ReadRule represents output schema for retrieving rule
type ReadRule struct {
	ID            string            `json:"id"`
	Name          string            `json:"name,omitempty"`
	Code          string            `json:"code,omitempty"`
	Description   string            `json:"description,omitempty"`
	Type          string            `json:"type,omitempty"`
	Status        string            `json:"status,omitempty"`
	RuleKey       string            `json:"rule_key,omitempty"`
	TemplateID    string            `json:"template_id,omitempty"`
	NodeKey       string            `json:"node_key,omitempty"`
	Conditions    types.StringArray `json:"conditions,omitempty"`
	Actions       types.JSON        `json:"actions,omitempty"`
	Priority      int               `json:"priority,omitempty"`
	IsEnabled     bool              `json:"is_enabled,omitempty"`
	EffectiveTime *int64            `json:"effective_time,omitempty"`
	ExpireTime    *int64            `json:"expire_time,omitempty"`
	TenantID      string            `json:"tenant_id,omitempty"`
	Extras        types.JSON        `json:"extras,omitempty"`
	CreatedBy     *string           `json:"created_by,omitempty"`
	CreatedAt     *int64            `json:"created_at,omitempty"`
	UpdatedBy     *string           `json:"updated_by,omitempty"`
	UpdatedAt     *int64            `json:"updated_at,omitempty"`
}

// GetID returns ID of the rule
func (r *ReadRule) GetID() string {
	return r.ID
}

// GetCursorValue returns cursor value
func (r *ReadRule) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// FindRuleParams represents query parameters for finding rules
type FindRuleParams struct {
	RuleKey    string `form:"rule_key,omitempty" json:"rule_key,omitempty"`
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	NodeKey    string `form:"node_key,omitempty" json:"node_key,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	IsEnabled  *bool  `form:"is_enabled,omitempty" json:"is_enabled,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
}

// ListRuleParams represents list parameters for rules
type ListRuleParams struct {
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction  string `form:"direction,omitempty" json:"direction,omitempty"`
	Type       string `form:"type,omitempty" json:"type,omitempty"`
	Status     string `form:"status,omitempty" json:"status,omitempty"`
	TemplateID string `form:"template_id,omitempty" json:"template_id,omitempty"`
	NodeKey    string `form:"node_key,omitempty" json:"node_key,omitempty"`
	IsEnabled  *bool  `form:"is_enabled,omitempty" json:"is_enabled,omitempty"`
	Tenant     string `form:"tenant,omitempty" json:"tenant,omitempty"`
}
