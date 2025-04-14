package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
)

// DelegationBody represents a delegation entity base fields
type DelegationBody struct {
	DelegatorID string            `json:"delegator_id,omitempty"`
	DelegateeID string            `json:"delegatee_id,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	NodeType    string            `json:"node_type,omitempty"`
	Conditions  types.StringArray `json:"conditions,omitempty"`
	StartTime   int64             `json:"start_time,omitempty"`
	EndTime     int64             `json:"end_time,omitempty"`
	IsEnabled   bool              `json:"is_enabled,omitempty"`
	Status      string            `json:"status,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	Extras      types.JSON        `json:"extras,omitempty"`
}

// CreateDelegationBody represents body for creating delegation
type CreateDelegationBody struct {
	DelegationBody
	TenantID string `json:"tenant_id,omitempty"`
}

// UpdateDelegationBody represents body for updating delegation
type UpdateDelegationBody struct {
	ID string `json:"id,omitempty"`
	DelegationBody
}

// ReadDelegation represents output schema for retrieving delegation
type ReadDelegation struct {
	ID          string            `json:"id"`
	DelegatorID string            `json:"delegator_id,omitempty"`
	DelegateeID string            `json:"delegatee_id,omitempty"`
	TemplateID  string            `json:"template_id,omitempty"`
	NodeType    string            `json:"node_type,omitempty"`
	Conditions  types.StringArray `json:"conditions,omitempty"`
	StartTime   int64             `json:"start_time,omitempty"`
	EndTime     int64             `json:"end_time,omitempty"`
	IsEnabled   bool              `json:"is_enabled,omitempty"`
	Status      string            `json:"status,omitempty"`
	TenantID    string            `json:"tenant_id,omitempty"`
	Extras      types.JSON        `json:"extras,omitempty"`
	CreatedBy   *string           `json:"created_by,omitempty"`
	CreatedAt   *int64            `json:"created_at,omitempty"`
	UpdatedBy   *string           `json:"updated_by,omitempty"`
	UpdatedAt   *int64            `json:"updated_at,omitempty"`
}

// GetID returns ID of the delegation
func (r *ReadDelegation) GetID() string {
	return r.ID
}

// GetCursorValue returns cursor value
func (r *ReadDelegation) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// FindDelegationParams represents query parameters for finding delegations
type FindDelegationParams struct {
	DelegatorID string `form:"delegator_id,omitempty" json:"delegator_id,omitempty"`
	DelegateeID string `form:"delegatee_id,omitempty" json:"delegatee_id,omitempty"`
	TemplateID  string `form:"template_id,omitempty" json:"template_id,omitempty"`
	NodeType    string `form:"node_type,omitempty" json:"node_type,omitempty"`
	IsEnabled   *bool  `form:"is_enabled,omitempty" json:"is_enabled,omitempty"`
	Status      string `form:"status,omitempty" json:"status,omitempty"`
	StartTime   *int64 `form:"start_time,omitempty" json:"start_time,omitempty"`
	EndTime     *int64 `form:"end_time,omitempty" json:"end_time,omitempty"`
	Tenant      string `form:"tenant,omitempty" json:"tenant,omitempty"`
}

// ListDelegationParams represents list parameters for delegations
type ListDelegationParams struct {
	Cursor      string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit       int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction   string `form:"direction,omitempty" json:"direction,omitempty"`
	DelegatorID string `form:"delegator_id,omitempty" json:"delegator_id,omitempty"`
	TemplateID  string `form:"template_id,omitempty" json:"template_id,omitempty"`
	NodeType    string `form:"node_type,omitempty" json:"node_type,omitempty"`
	IsEnabled   *bool  `form:"is_enabled,omitempty" json:"is_enabled,omitempty"`
	Status      string `form:"status,omitempty" json:"status,omitempty"`
	Tenant      string `form:"tenant,omitempty" json:"tenant,omitempty"`
}
