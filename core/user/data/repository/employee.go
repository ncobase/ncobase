package repository

import (
	"context"
	"fmt"
	"ncobase/user/data"
	"ncobase/user/data/ent"
	employeeEnt "ncobase/user/data/ent/employee"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// EmployeeRepositoryInterface represents the employee repository interface
type EmployeeRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateEmployeeBody) (*ent.Employee, error)
	Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*ent.Employee, error)
	GetByUserID(ctx context.Context, userID string) (*ent.Employee, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*ent.Employee, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, params *structs.ListEmployeeParams) ([]*ent.Employee, error)
	GetByDepartment(ctx context.Context, department string) ([]*ent.Employee, error)
	GetByManager(ctx context.Context, managerID string) ([]*ent.Employee, error)
	CountX(ctx context.Context, params *structs.ListEmployeeParams) int
}

// employeeRepository implements the EmployeeRepositoryInterface
type employeeRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Employee]
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(d *data.Data) EmployeeRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &employeeRepository{ec, rc, cache.NewCache[ent.Employee](rc, "ncse_employee")}
}

// Create creates a new employee record
func (r *employeeRepository) Create(ctx context.Context, body *structs.CreateEmployeeBody) (*ent.Employee, error) {
	builder := r.ec.Employee.Create()

	// Set required fields
	builder.SetID(body.UserID)
	builder.SetTenantID(body.TenantID)

	// Set optional fields
	if body.EmployeeID != "" {
		builder.SetEmployeeID(body.EmployeeID)
	}
	if body.Department != "" {
		builder.SetDepartment(body.Department)
	}
	if body.Position != "" {
		builder.SetPosition(body.Position)
	}
	if body.ManagerID != "" {
		builder.SetManagerID(body.ManagerID)
	}
	if body.HireDate != nil {
		builder.SetHireDate(*body.HireDate)
	}
	if body.TerminationDate != nil {
		builder.SetTerminationDate(*body.TerminationDate)
	}
	if body.EmploymentType != "" {
		builder.SetEmploymentType(employeeEnt.EmploymentType(body.EmploymentType))
	}
	if body.Status != "" {
		builder.SetStatus(employeeEnt.Status(body.Status))
	}
	if body.Salary > 0 {
		builder.SetSalary(body.Salary)
	}
	if body.WorkLocation != "" {
		builder.SetWorkLocation(body.WorkLocation)
	}
	if body.ContactInfo != nil {
		builder.SetContactInfo(*body.ContactInfo)
	}
	if body.Skills != nil {
		builder.SetSkills(*body.Skills)
	}
	if body.Certifications != nil {
		builder.SetCertifications(*body.Certifications)
	}
	if body.Extras != nil {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the result
	cacheKey := fmt.Sprintf("user:%s", body.UserID)
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "employeeRepo.Create cache error: %v", err)
	}

	return row, nil
}

// Update updates an employee record
func (r *employeeRepository) Update(ctx context.Context, userID string, body *structs.UpdateEmployeeBody) (*ent.Employee, error) {
	employee, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	builder := employee.Update()

	// Update fields if provided
	if body.EmployeeID != "" {
		builder.SetEmployeeID(body.EmployeeID)
	}
	if body.Department != "" {
		builder.SetDepartment(body.Department)
	}
	if body.Position != "" {
		builder.SetPosition(body.Position)
	}
	if body.ManagerID != "" {
		builder.SetManagerID(body.ManagerID)
	}
	if body.HireDate != nil {
		builder.SetHireDate(*body.HireDate)
	}
	if body.TerminationDate != nil {
		builder.SetTerminationDate(*body.TerminationDate)
	}
	if body.EmploymentType != "" {
		builder.SetEmploymentType(employeeEnt.EmploymentType(body.EmploymentType))
	}
	if body.Status != "" {
		builder.SetStatus(employeeEnt.Status(body.Status))
	}
	if body.Salary > 0 {
		builder.SetSalary(body.Salary)
	}
	if body.WorkLocation != "" {
		builder.SetWorkLocation(body.WorkLocation)
	}
	if body.ContactInfo != nil {
		builder.SetContactInfo(*body.ContactInfo)
	}
	if body.Skills != nil {
		builder.SetSkills(*body.Skills)
	}
	if body.Certifications != nil {
		builder.SetCertifications(*body.Certifications)
	}
	if body.Extras != nil {
		builder.SetExtras(*body.Extras)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.Update error: %v", err)
		return nil, err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "employeeRepo.Update cache error: %v", err)
	}

	return row, nil
}

// GetByUserID retrieves an employee by user ID
func (r *employeeRepository) GetByUserID(ctx context.Context, userID string) (*ent.Employee, error) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.Employee.Query().Where(employeeEnt.IDEQ(userID)).Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByUserID error: %v", err)
		return nil, err
	}

	// Cache the result
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByUserID cache error: %v", err)
	}

	return row, nil
}

// GetByEmployeeID retrieves an employee by employee ID
func (r *employeeRepository) GetByEmployeeID(ctx context.Context, employeeID string) (*ent.Employee, error) {
	cacheKey := fmt.Sprintf("emp:%s", employeeID)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.Employee.Query().Where(employeeEnt.EmployeeIDEQ(employeeID)).Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByEmployeeID error: %v", err)
		return nil, err
	}

	// Cache the result
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByEmployeeID cache error: %v", err)
	}

	return row, nil
}

// Delete deletes an employee record
func (r *employeeRepository) Delete(ctx context.Context, userID string) error {
	if _, err := r.ec.Employee.Delete().Where(employeeEnt.IDEQ(userID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "employeeRepo.Delete error: %v", err)
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("user:%s", userID)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "employeeRepo.Delete cache error: %v", err)
	}

	return nil
}

// List retrieves employees with filtering
func (r *employeeRepository) List(ctx context.Context, params *structs.ListEmployeeParams) ([]*ent.Employee, error) {
	builder := r.ec.Employee.Query()

	// Apply filters
	if params.TenantID != "" {
		builder.Where(employeeEnt.TenantIDEQ(params.TenantID))
	}
	if params.Department != "" {
		builder.Where(employeeEnt.DepartmentEQ(params.Department))
	}
	if params.Status != "" {
		builder.Where(employeeEnt.StatusEQ(employeeEnt.Status(params.Status)))
	}
	if params.EmploymentType != "" {
		builder.Where(employeeEnt.EmploymentTypeEQ(employeeEnt.EmploymentType(params.EmploymentType)))
	}
	if params.ManagerID != "" {
		builder.Where(employeeEnt.ManagerIDEQ(params.ManagerID))
	}

	// Apply pagination
	if params.Limit > 0 {
		builder.Limit(params.Limit)
	}

	// Apply ordering
	builder.Order(ent.Desc(employeeEnt.FieldCreatedAt))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// GetByDepartment retrieves employees by department
func (r *employeeRepository) GetByDepartment(ctx context.Context, department string) ([]*ent.Employee, error) {
	rows, err := r.ec.Employee.Query().Where(employeeEnt.DepartmentEQ(department)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByDepartment error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetByManager retrieves employees by manager
func (r *employeeRepository) GetByManager(ctx context.Context, managerID string) ([]*ent.Employee, error) {
	rows, err := r.ec.Employee.Query().Where(employeeEnt.ManagerIDEQ(managerID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "employeeRepo.GetByManager error: %v", err)
		return nil, err
	}
	return rows, nil
}

// CountX gets a count of employees
func (r *employeeRepository) CountX(ctx context.Context, params *structs.ListEmployeeParams) int {
	builder := r.ec.Employee.Query()

	// Apply same filters as List
	if params.TenantID != "" {
		builder.Where(employeeEnt.TenantIDEQ(params.TenantID))
	}
	if params.Department != "" {
		builder.Where(employeeEnt.DepartmentEQ(params.Department))
	}
	if params.Status != "" {
		builder.Where(employeeEnt.StatusEQ(employeeEnt.Status(params.Status)))
	}
	if params.EmploymentType != "" {
		builder.Where(employeeEnt.EmploymentTypeEQ(employeeEnt.EmploymentType(params.EmploymentType)))
	}
	if params.ManagerID != "" {
		builder.Where(employeeEnt.ManagerIDEQ(params.ManagerID))
	}

	return builder.CountX(ctx)
}
