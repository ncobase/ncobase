package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// QuotaType represents different types of quotas
type QuotaType string

const (
	QuotaTypeUser    QuotaType = "users"
	QuotaTypeStorage QuotaType = "storage"
	QuotaTypeAPI     QuotaType = "api_calls"
	QuotaTypeProject QuotaType = "projects"
	QuotaTypeCustom  QuotaType = "custom"
)

// QuotaUnit represents the unit of measurement
type QuotaUnit string

const (
	UnitCount QuotaUnit = "count"
	UnitBytes QuotaUnit = "bytes"
	UnitMB    QuotaUnit = "mb"
	UnitGB    QuotaUnit = "gb"
	UnitTB    QuotaUnit = "tb"
)

// TenantQuotaBody represents quota configuration for a tenant
type TenantQuotaBody struct {
	TenantID    string      `json:"tenant_id,omitempty"`
	QuotaType   QuotaType   `json:"quota_type,omitempty"`
	QuotaName   string      `json:"quota_name,omitempty"`
	MaxValue    int64       `json:"max_value,omitempty"`
	CurrentUsed int64       `json:"current_used,omitempty"`
	Unit        QuotaUnit   `json:"unit,omitempty"`
	Description string      `json:"description,omitempty"`
	Enabled     bool        `json:"enabled,omitempty"`
	Extras      *types.JSON `json:"extras,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateTenantQuotaBody represents the body for creating tenant quota
type CreateTenantQuotaBody struct {
	TenantQuotaBody
}

// UpdateTenantQuotaBody represents the body for updating tenant quota
type UpdateTenantQuotaBody struct {
	ID string `json:"id,omitempty"`
	TenantQuotaBody
}

// ReadTenantQuota represents the output schema for retrieving tenant quota
type ReadTenantQuota struct {
	ID                 string      `json:"id"`
	TenantID           string      `json:"tenant_id"`
	QuotaType          QuotaType   `json:"quota_type"`
	QuotaName          string      `json:"quota_name"`
	MaxValue           int64       `json:"max_value"`
	CurrentUsed        int64       `json:"current_used"`
	Unit               QuotaUnit   `json:"unit"`
	Description        string      `json:"description"`
	Enabled            bool        `json:"enabled"`
	UtilizationPercent float64     `json:"utilization_percent"`
	IsExceeded         bool        `json:"is_exceeded"`
	Extras             *types.JSON `json:"extras,omitempty"`
	CreatedBy          *string     `json:"created_by,omitempty"`
	CreatedAt          *int64      `json:"created_at,omitempty"`
	UpdatedBy          *string     `json:"updated_by,omitempty"`
	UpdatedAt          *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value
func (r *ReadTenantQuota) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// CalculateUtilization calculates utilization percentage
func (r *ReadTenantQuota) CalculateUtilization() {
	if r.MaxValue > 0 {
		r.UtilizationPercent = float64(r.CurrentUsed) / float64(r.MaxValue) * 100
		r.IsExceeded = r.CurrentUsed > r.MaxValue
	}
}

// QuotaUsageRequest represents request to update quota usage
type QuotaUsageRequest struct {
	TenantID  string    `json:"tenant_id" validate:"required"`
	QuotaType QuotaType `json:"quota_type" validate:"required"`
	Delta     int64     `json:"delta" validate:"required"`
}

// ListTenantQuotaParams represents the query parameters for listing tenant quotas
type ListTenantQuotaParams struct {
	TenantID  string    `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	QuotaType QuotaType `form:"quota_type,omitempty" json:"quota_type,omitempty"`
	Enabled   *bool     `form:"enabled,omitempty" json:"enabled,omitempty"`
	Cursor    string    `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int       `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string    `form:"direction,omitempty" json:"direction,omitempty"`
}
