// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"ncobase/user/data/ent/employee"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// Employee is the model entity for the Employee schema.
type Employee struct {
	config `json:"-"`
	// ID of the ent.
	// user primary key alias
	ID string `json:"id,omitempty"`
	// space id, e.g. space id, organization id, store id
	SpaceID string `json:"space_id,omitempty"`
	// created at
	CreatedAt int64 `json:"created_at,omitempty"`
	// updated at
	UpdatedAt int64 `json:"updated_at,omitempty"`
	// Employee ID/Number
	EmployeeID string `json:"employee_id,omitempty"`
	// Primary department
	Department string `json:"department,omitempty"`
	// Job position/title
	Position string `json:"position,omitempty"`
	// Direct manager user ID
	ManagerID string `json:"manager_id,omitempty"`
	// Hire date
	HireDate time.Time `json:"hire_date,omitempty"`
	// Termination date
	TerminationDate *time.Time `json:"termination_date,omitempty"`
	// Employment type
	EmploymentType employee.EmploymentType `json:"employment_type,omitempty"`
	// Employee status
	Status employee.Status `json:"status,omitempty"`
	// Base salary
	Salary float64 `json:"salary,omitempty"`
	// Primary work location
	WorkLocation string `json:"work_location,omitempty"`
	// Emergency contact info
	ContactInfo map[string]interface{} `json:"contact_info,omitempty"`
	// Employee skills
	Skills []string `json:"skills,omitempty"`
	// Professional certifications
	Certifications []string `json:"certifications,omitempty"`
	// Additional employee data
	Extras       map[string]interface{} `json:"extras,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Employee) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case employee.FieldContactInfo, employee.FieldSkills, employee.FieldCertifications, employee.FieldExtras:
			values[i] = new([]byte)
		case employee.FieldSalary:
			values[i] = new(sql.NullFloat64)
		case employee.FieldCreatedAt, employee.FieldUpdatedAt:
			values[i] = new(sql.NullInt64)
		case employee.FieldID, employee.FieldSpaceID, employee.FieldEmployeeID, employee.FieldDepartment, employee.FieldPosition, employee.FieldManagerID, employee.FieldEmploymentType, employee.FieldStatus, employee.FieldWorkLocation:
			values[i] = new(sql.NullString)
		case employee.FieldHireDate, employee.FieldTerminationDate:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Employee fields.
func (e *Employee) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case employee.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				e.ID = value.String
			}
		case employee.FieldSpaceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field space_id", values[i])
			} else if value.Valid {
				e.SpaceID = value.String
			}
		case employee.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				e.CreatedAt = value.Int64
			}
		case employee.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				e.UpdatedAt = value.Int64
			}
		case employee.FieldEmployeeID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field employee_id", values[i])
			} else if value.Valid {
				e.EmployeeID = value.String
			}
		case employee.FieldDepartment:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field department", values[i])
			} else if value.Valid {
				e.Department = value.String
			}
		case employee.FieldPosition:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field position", values[i])
			} else if value.Valid {
				e.Position = value.String
			}
		case employee.FieldManagerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field manager_id", values[i])
			} else if value.Valid {
				e.ManagerID = value.String
			}
		case employee.FieldHireDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field hire_date", values[i])
			} else if value.Valid {
				e.HireDate = value.Time
			}
		case employee.FieldTerminationDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field termination_date", values[i])
			} else if value.Valid {
				e.TerminationDate = new(time.Time)
				*e.TerminationDate = value.Time
			}
		case employee.FieldEmploymentType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field employment_type", values[i])
			} else if value.Valid {
				e.EmploymentType = employee.EmploymentType(value.String)
			}
		case employee.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				e.Status = employee.Status(value.String)
			}
		case employee.FieldSalary:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field salary", values[i])
			} else if value.Valid {
				e.Salary = value.Float64
			}
		case employee.FieldWorkLocation:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field work_location", values[i])
			} else if value.Valid {
				e.WorkLocation = value.String
			}
		case employee.FieldContactInfo:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field contact_info", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.ContactInfo); err != nil {
					return fmt.Errorf("unmarshal field contact_info: %w", err)
				}
			}
		case employee.FieldSkills:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field skills", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Skills); err != nil {
					return fmt.Errorf("unmarshal field skills: %w", err)
				}
			}
		case employee.FieldCertifications:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field certifications", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Certifications); err != nil {
					return fmt.Errorf("unmarshal field certifications: %w", err)
				}
			}
		case employee.FieldExtras:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field extras", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &e.Extras); err != nil {
					return fmt.Errorf("unmarshal field extras: %w", err)
				}
			}
		default:
			e.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Employee.
// This includes values selected through modifiers, order, etc.
func (e *Employee) Value(name string) (ent.Value, error) {
	return e.selectValues.Get(name)
}

// Update returns a builder for updating this Employee.
// Note that you need to call Employee.Unwrap() before calling this method if this Employee
// was returned from a transaction, and the transaction was committed or rolled back.
func (e *Employee) Update() *EmployeeUpdateOne {
	return NewEmployeeClient(e.config).UpdateOne(e)
}

// Unwrap unwraps the Employee entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (e *Employee) Unwrap() *Employee {
	_tx, ok := e.config.driver.(*txDriver)
	if !ok {
		panic("ent: Employee is not a transactional entity")
	}
	e.config.driver = _tx.drv
	return e
}

// String implements the fmt.Stringer.
func (e *Employee) String() string {
	var builder strings.Builder
	builder.WriteString("Employee(")
	builder.WriteString(fmt.Sprintf("id=%v, ", e.ID))
	builder.WriteString("space_id=")
	builder.WriteString(e.SpaceID)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(fmt.Sprintf("%v", e.CreatedAt))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(fmt.Sprintf("%v", e.UpdatedAt))
	builder.WriteString(", ")
	builder.WriteString("employee_id=")
	builder.WriteString(e.EmployeeID)
	builder.WriteString(", ")
	builder.WriteString("department=")
	builder.WriteString(e.Department)
	builder.WriteString(", ")
	builder.WriteString("position=")
	builder.WriteString(e.Position)
	builder.WriteString(", ")
	builder.WriteString("manager_id=")
	builder.WriteString(e.ManagerID)
	builder.WriteString(", ")
	builder.WriteString("hire_date=")
	builder.WriteString(e.HireDate.Format(time.ANSIC))
	builder.WriteString(", ")
	if v := e.TerminationDate; v != nil {
		builder.WriteString("termination_date=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("employment_type=")
	builder.WriteString(fmt.Sprintf("%v", e.EmploymentType))
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", e.Status))
	builder.WriteString(", ")
	builder.WriteString("salary=")
	builder.WriteString(fmt.Sprintf("%v", e.Salary))
	builder.WriteString(", ")
	builder.WriteString("work_location=")
	builder.WriteString(e.WorkLocation)
	builder.WriteString(", ")
	builder.WriteString("contact_info=")
	builder.WriteString(fmt.Sprintf("%v", e.ContactInfo))
	builder.WriteString(", ")
	builder.WriteString("skills=")
	builder.WriteString(fmt.Sprintf("%v", e.Skills))
	builder.WriteString(", ")
	builder.WriteString("certifications=")
	builder.WriteString(fmt.Sprintf("%v", e.Certifications))
	builder.WriteString(", ")
	builder.WriteString("extras=")
	builder.WriteString(fmt.Sprintf("%v", e.Extras))
	builder.WriteByte(')')
	return builder.String()
}

// Employees is a parsable slice of Employee.
type Employees []*Employee
