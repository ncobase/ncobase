package structs

import (
	"time"

	"github.com/ncobase/common/validator"
)

// Validate is a wrapper around validator.Validate that returns a map of JSON field names to friendly error messages.
var Validate = validator.ValidateStruct

// BaseEntity contains common fields for entities.
type BaseEntity struct {
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
