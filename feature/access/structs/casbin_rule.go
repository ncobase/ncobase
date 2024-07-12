package structs

import (
	"time"
)

// CasbinRuleBody defines the structure for request body used to create or update Casbin rules.
type CasbinRuleBody struct {
	PType     string  `json:"p_type" validate:"required"`
	V0        string  `json:"v0" validate:"required"`
	V1        string  `json:"v1" validate:"required"`
	V2        string  `json:"v2" validate:"required"`
	V3        *string `json:"v3"`
	V4        *string `json:"v4"`
	V5        *string `json:"v5"`
	CreatedBy *string `json:"created_by,omitempty"`
	UpdatedBy *string `json:"updated_by,omitempty"`
}

// ReadCasbinRule represents a single Casbin rule.
type ReadCasbinRule struct {
	PType     string     `json:"p_type"`
	V0        string     `json:"v0"`
	V1        string     `json:"v1"`
	V2        string     `json:"v2"`
	V3        *string    `json:"v3"`
	V4        *string    `json:"v4"`
	V5        *string    `json:"v5"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// ListCasbinRuleParams defines the parameters for listing Casbin rules.
type ListCasbinRuleParams struct {
	PType  *string `form:"p_type" json:"p_type,omitempty"`
	V0     *string `form:"v0" json:"v0,omitempty"`
	V1     *string `form:"v1" json:"v1,omitempty"`
	V2     *string `form:"v2" json:"v2,omitempty"`
	V3     *string `form:"v3" json:"v3,omitempty"`
	V4     *string `form:"v4" json:"v4,omitempty"`
	V5     *string `form:"v5" json:"v5,omitempty"`
	Cursor string  `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit  int     `form:"limit,omitempty" json:"limit,omitempty"`
}
