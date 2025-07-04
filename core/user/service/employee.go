package service

import (
	"context"
	"errors"
	"ncobase/user/data/ent"
	"ncobase/user/data/repository"
	"ncobase/user/event"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// EmployeeServiceInterface defines the employee service interface
type EmployeeServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateEmployeeBody) (*structs.ReadEmployee, error)
	Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*structs.ReadEmployee, error)
	Get(ctx context.Context, userID string) (*structs.ReadEmployee, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*structs.ReadEmployee, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, params *structs.ListEmployeeParams) (paging.Result[*structs.ReadEmployee], error)
	GetByDepartment(ctx context.Context, department string) ([]*structs.ReadEmployee, error)
	GetByManager(ctx context.Context, managerID string) ([]*structs.ReadEmployee, error)
	CountX(ctx context.Context, params *structs.ListEmployeeParams) int
	Serialize(employee *ent.Employee) *structs.ReadEmployee
	Serializes(employees []*ent.Employee) []*structs.ReadEmployee
}

// employeeService implements the EmployeeServiceInterface
type employeeService struct {
	employee repository.EmployeeRepositoryInterface
	ep       event.PublisherInterface
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(repo *repository.Repository, ep event.PublisherInterface) EmployeeServiceInterface {
	return &employeeService{
		employee: repo.Employee,
		ep:       ep,
	}
}

// Create creates a new employee record
func (s *employeeService) Create(ctx context.Context, body *structs.CreateEmployeeBody) (*structs.ReadEmployee, error) {
	row, err := s.employee.Create(ctx, body)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Update updates an employee record
func (s *employeeService) Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*structs.ReadEmployee, error) {
	row, err := s.employee.Update(ctx, userID, body)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Get retrieves an employee by user ID
func (s *employeeService) Get(ctx context.Context, userID string) (*structs.ReadEmployee, error) {
	row, err := s.employee.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByEmployeeID retrieves an employee by employee ID
func (s *employeeService) GetByEmployeeID(ctx context.Context, employeeID string) (*structs.ReadEmployee, error) {
	row, err := s.employee.GetByEmployeeID(ctx, employeeID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Delete deletes an employee record
func (s *employeeService) Delete(ctx context.Context, userID string) error {
	err := s.employee.Delete(ctx, userID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return err
	}
	return nil
}

// List lists employees with cursor-based pagination
func (s *employeeService) List(ctx context.Context, params *structs.ListEmployeeParams) (paging.Result[*structs.ReadEmployee], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadEmployee, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.employee.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing employees: %v", err)
			return nil, 0, err
		}

		total := s.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// GetByDepartment retrieves employees by department
func (s *employeeService) GetByDepartment(ctx context.Context, department string) ([]*structs.ReadEmployee, error) {
	rows, err := s.employee.GetByDepartment(ctx, department)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// GetByManager retrieves employees by manager
func (s *employeeService) GetByManager(ctx context.Context, managerID string) ([]*structs.ReadEmployee, error) {
	rows, err := s.employee.GetByManager(ctx, managerID)
	if err := handleEntError(ctx, "Employee", err); err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// CountX gets a count of employees
func (s *employeeService) CountX(ctx context.Context, params *structs.ListEmployeeParams) int {
	return s.employee.CountX(ctx, params)
}

// Serialize converts ent.Employee to structs.ReadEmployee
func (s *employeeService) Serialize(row *ent.Employee) *structs.ReadEmployee {
	return &structs.ReadEmployee{
		UserID:          row.ID,
		SpaceID:         row.SpaceID,
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
	rs := make([]*structs.ReadEmployee, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}
