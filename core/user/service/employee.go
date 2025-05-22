package service

import (
	"context"
	"ncobase/user/data"
	"ncobase/user/data/ent"
	"ncobase/user/data/repository"
	"ncobase/user/structs"
)

// EmployeeServiceInterface defines the employee service interface
type EmployeeServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateEmployeeBody) (*structs.ReadEmployee, error)
	Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*structs.ReadEmployee, error)
	Get(ctx context.Context, userID string) (*structs.ReadEmployee, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*structs.ReadEmployee, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, params *structs.ListEmployeeParams) ([]*structs.ReadEmployee, error)
	GetByDepartment(ctx context.Context, department string) ([]*structs.ReadEmployee, error)
	GetByManager(ctx context.Context, managerID string) ([]*structs.ReadEmployee, error)
	Serialize(employee *ent.Employee) *structs.ReadEmployee
	Serializes(employees []*ent.Employee) []*structs.ReadEmployee
}

// employeeService implements the EmployeeServiceInterface
type employeeService struct {
	r repository.EmployeeRepositoryInterface
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(d *data.Data) EmployeeServiceInterface {
	return &employeeService{
		r: repository.NewEmployeeRepository(d),
	}
}

// Create creates a new employee record
func (s *employeeService) Create(ctx context.Context, body *structs.CreateEmployeeBody) (*structs.ReadEmployee, error) {
	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Update updates an employee record
func (s *employeeService) Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*structs.ReadEmployee, error) {
	row, err := s.r.Update(ctx, userID, body)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Get retrieves an employee by user ID
func (s *employeeService) Get(ctx context.Context, userID string) (*structs.ReadEmployee, error) {
	row, err := s.r.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByEmployeeID retrieves an employee by employee ID
func (s *employeeService) GetByEmployeeID(ctx context.Context, employeeID string) (*structs.ReadEmployee, error) {
	row, err := s.r.GetByEmployeeID(ctx, employeeID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Delete deletes an employee record
func (s *employeeService) Delete(ctx context.Context, userID string) error {
	err := s.r.Delete(ctx, userID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return err
	}
	return nil
}

// List lists employees with filtering
func (s *employeeService) List(ctx context.Context, params *structs.ListEmployeeParams) ([]*structs.ReadEmployee, error) {
	rows, err := s.r.List(ctx, params)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// GetByDepartment retrieves employees by department
func (s *employeeService) GetByDepartment(ctx context.Context, department string) ([]*structs.ReadEmployee, error) {
	rows, err := s.r.GetByDepartment(ctx, department)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// GetByManager retrieves employees by manager
func (s *employeeService) GetByManager(ctx context.Context, managerID string) ([]*structs.ReadEmployee, error) {
	rows, err := s.r.GetByManager(ctx, managerID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// Serialize converts ent.Employee to structs.ReadEmployee
func (s *employeeService) Serialize(row *ent.Employee) *structs.ReadEmployee {
	return &structs.ReadEmployee{
		UserID:          row.ID,
		TenantID:        row.TenantID,
		EmployeeID:      row.EmployeeID,
		Department:      row.Department,
		Position:        row.Position,
		ManagerID:       row.ManagerID,
		HireDate:        &row.HireDate,
		TerminationDate: row.TerminationDate,
		EmploymentType:  string(row.EmploymentType),
		Status:          string(row.Status),
		Salary:          row.Salary,
		WorkLocation:    row.WorkLocation,
		ContactInfo:     &row.ContactInfo,
		Skills:          &row.Skills,
		Certifications:  &row.Certifications,
		Extras:          &row.Extras,
		CreatedAt:       &row.CreatedAt,
		UpdatedAt:       &row.UpdatedAt,
	}
}

// Serializes converts multiple ent.Employee to structs.ReadEmployee
func (s *employeeService) Serializes(rows []*ent.Employee) []*structs.ReadEmployee {
	var employees []*structs.ReadEmployee
	for _, row := range rows {
		employees = append(employees, s.Serialize(row))
	}
	return employees
}
