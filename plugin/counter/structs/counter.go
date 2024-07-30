package structs

import (
	"fmt"
	"ncobase/common/types"
)

// CounterBody represents a counter entity.
type CounterBody struct {
	Identifier    string  `json:"identifier,omitempty"`
	Name          string  `json:"name,omitempty"`
	Prefix        string  `json:"prefix,omitempty"`
	Suffix        string  `json:"suffix,omitempty"`
	StartValue    int     `json:"start_value,omitempty"`
	IncrementStep int     `json:"increment_step,omitempty"`
	DateFormat    string  `json:"date_format,omitempty"`
	CurrentValue  int     `json:"current_value,omitempty"`
	Disabled      bool    `json:"disabled,omitempty"`
	Description   string  `json:"description,omitempty"`
	TenantID      *string `json:"tenant_id,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
	UpdatedBy     *string `json:"updated_by,omitempty"`
}

// CreateCounterBody represents the body for creating or updating a counter.
type CreateCounterBody struct {
	CounterBody
}

// UpdateCounterBody represents the body for updating a counter.
type UpdateCounterBody struct {
	ID string `json:"id,omitempty"`
	CounterBody
}

// ReadCounter represents the output schema for retrieving a counter.
type ReadCounter struct {
	ID            string  `json:"id"`
	Identifier    string  `json:"identifier"`
	Name          string  `json:"name"`
	Prefix        string  `json:"prefix"`
	Suffix        string  `json:"suffix"`
	StartValue    int     `json:"start_value"`
	IncrementStep int     `json:"increment_step"`
	DateFormat    string  `json:"date_format"`
	CurrentValue  int     `json:"current_value"`
	Disabled      bool    `json:"disabled"`
	Description   string  `json:"description"`
	TenantID      *string `json:"tenant_id,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
	CreatedAt     *int64  `json:"created_at,omitempty"`
	UpdatedBy     *string `json:"updated_by,omitempty"`
	UpdatedAt     *int64  `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadCounter) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, types.ToValue(r.CreatedAt))
}

// FindCounter represents the parameters for finding a counter.
type FindCounter struct {
	Counter  string `form:"counter,omitempty" json:"counter,omitempty"`
	Tenant   string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Disabled bool   `form:"disabled,omitempty" json:"disabled,omitempty"`
}

// ListCounterParams represents the query parameters for listing counters.
type ListCounterParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
	Disabled  bool   `form:"disabled,omitempty" json:"disabled,omitempty"`
}
