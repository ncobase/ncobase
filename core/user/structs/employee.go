package structs

import (
	"fmt"
	"time"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// EmployeeBody represents employee information
type EmployeeBody struct {
	UserID          string      `json:"user_id,omitempty"`
	TenantID        string      `json:"tenant_id,omitempty"`
	EmployeeID      string      `json:"employee_id,omitempty"`
	Department      string      `json:"department,omitempty"`
	Position        string      `json:"position,omitempty"`
	ManagerID       string      `json:"manager_id,omitempty"`
	HireDate        *time.Time  `json:"hire_date,omitempty"`
	TerminationDate *time.Time  `json:"termination_date,omitempty"`
	EmploymentType  string      `json:"employment_type,omitempty"`
	Status          string      `json:"status,omitempty"`
	Salary          float64     `json:"salary,omitempty"`
	WorkLocation    string      `json:"work_location,omitempty"`
	ContactInfo     *types.JSON `json:"contact_info,omitempty"`
	Skills          *[]string   `json:"skills,omitempty"`
	Certifications  *[]string   `json:"certifications,omitempty"`
	Extras          *types.JSON `json:"extras,omitempty"`
}

// CreateEmployeeBody represents the body for creating employee record
type CreateEmployeeBody struct {
	EmployeeBody
}

// UpdateEmployeeBody represents the body for updating employee record
type UpdateEmployeeBody struct {
	EmployeeBody
}

// ReadEmployee represents the output schema for retrieving employee
type ReadEmployee struct {
	UserID          string      `json:"user_id"`
	TenantID        string      `json:"tenant_id"`
	EmployeeID      string      `json:"employee_id,omitempty"`
	Department      string      `json:"department,omitempty"`
	Position        string      `json:"position,omitempty"`
	ManagerID       string      `json:"manager_id,omitempty"`
	HireDate        *time.Time  `json:"hire_date,omitempty"`
	TerminationDate *time.Time  `json:"termination_date,omitempty"`
	EmploymentType  string      `json:"employment_type,omitempty"`
	Status          string      `json:"status,omitempty"`
	Salary          float64     `json:"salary,omitempty"`
	WorkLocation    string      `json:"work_location,omitempty"`
	ContactInfo     *types.JSON `json:"contact_info,omitempty"`
	Skills          *[]string   `json:"skills,omitempty"`
	Certifications  *[]string   `json:"certifications,omitempty"`
	Extras          *types.JSON `json:"extras,omitempty"`
	CreatedAt       *int64      `json:"created_at,omitempty"`
	UpdatedAt       *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadEmployee) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.UserID, convert.ToValue(r.CreatedAt))
}

// ListEmployeeParams represents query parameters for listing employees
type ListEmployeeParams struct {
	Cursor         string `json:"cursor,omitempty" query:"cursor"`
	Limit          int    `json:"limit,omitempty" query:"limit"`
	Direction      string `json:"direction,omitempty" query:"direction"`
	TenantID       string `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	Department     string `form:"department,omitempty" json:"department,omitempty"`
	Status         string `form:"status,omitempty" json:"status,omitempty"`
	EmploymentType string `form:"employment_type,omitempty" json:"employment_type,omitempty"`
	ManagerID      string `form:"manager_id,omitempty" json:"manager_id,omitempty"`
}

// FindEmployee represents parameters for finding an employee
type FindEmployee struct {
	UserID     string `form:"user_id,omitempty" json:"user_id,omitempty"`
	EmployeeID string `form:"employee_id,omitempty" json:"employee_id,omitempty"`
	Department string `form:"department,omitempty" json:"department,omitempty"`
	ManagerID  string `form:"manager_id,omitempty" json:"manager_id,omitempty"`
}
