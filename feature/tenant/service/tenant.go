package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	accessService "ncobase/feature/access/service"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/ent"
	"ncobase/feature/tenant/data/repository"
	"ncobase/feature/tenant/structs"
	userService "ncobase/feature/user/service"
	userStructs "ncobase/feature/user/structs"
	"ncobase/helper"
)

// TenantServiceInterface is the interface for the service.
type TenantServiceInterface interface {
	Account(ctx context.Context) (*structs.ReadTenant, error)
	AccountTenants(ctx context.Context) (*types.JSON, error)
	UserOwn(ctx context.Context, uid string) (*structs.ReadTenant, error)
	Create(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error)
	Update(ctx context.Context, body *structs.UpdateTenantBody) (*structs.ReadTenant, error)
	Get(ctx context.Context, id string) (*structs.ReadTenant, error)
	GetByUser(ctx context.Context, uid string) (*structs.ReadTenant, error)
	Find(ctx context.Context, id string) (*structs.ReadTenant, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantParams) (*types.JSON, error)
	CreateInitial(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error)
	IsCreate(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error)
	Serializes(rows []*ent.Tenant) []*structs.ReadTenant
	Serialize(tenant *ent.Tenant) *structs.ReadTenant
}

// tenantService is the struct for the service.
type tenantService struct {
	tenant     repository.TenantRepositoryInterface
	userTenant repository.UserTenantRepositoryInterface
	us         *userService.Service
	as         *accessService.Service
}

// NewTenantService creates a new service.
func NewTenantService(d *data.Data, us *userService.Service, as *accessService.Service) TenantServiceInterface {
	return &tenantService{
		tenant:     repository.NewTenantRepository(d),
		userTenant: repository.NewUserTenantRepository(d),
		us:         us,
		as:         as,
	}
}

// Account retrieves the tenant associated with the user's account.
func (s *tenantService) Account(ctx context.Context) (*structs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	row, err := s.tenant.GetByUser(ctx, userID)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	// Serialize tenant data and return
	return s.Serialize(row), nil
}

// AccountTenants retrieves the tenant associated with the user's account.
func (s *tenantService) AccountTenants(ctx context.Context) (*types.JSON, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	rows, err := s.List(ctx, &structs.ListTenantParams{
		User: userID,
	})
	if err := handleEntError("Tenants", err); err != nil {
		return nil, err
	}

	return rows, nil
}

// UserOwn user own tenant service
func (s *tenantService) UserOwn(ctx context.Context, uid string) (*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	row, err := s.tenant.GetByUser(ctx, uid)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Create creates a tenant service.
func (s *tenantService) Create(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error) {
	if body.CreatedBy == nil {
		body.CreatedBy = types.ToPointer(helper.GetUserID(ctx))
	}

	// Create the tenant
	tenant, err := s.tenant.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	return s.Serialize(tenant), nil
}

// Update updates tenant service (full and partial).
func (s *tenantService) Update(ctx context.Context, body *structs.UpdateTenantBody) (*structs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the tenant
	if body.CreatedBy != nil {
		_, err := s.tenant.GetByUser(ctx, *body.CreatedBy)
		if err := handleEntError("Tenant", err); err != nil {
			return nil, err
		}
	}

	// If ID is not provided, get the tenant ID associated with the user
	if body.ID == "" {
		body.ID, _ = s.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	row, err := s.Find(ctx, body.ID)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator of the tenant
	if types.ToValue(row.CreatedBy) != userID {
		return nil, errors.New("this tenant is not yours")
	}

	// set updated by
	body.UpdatedBy = &userID

	// Serialize request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a types.JSON object
	var d types.JSON
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, err
	}

	// Update the tenant with the provided data
	_, err = s.tenant.Update(ctx, row.ID, d)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Get reads tenant service.
func (s *tenantService) Get(ctx context.Context, id string) (*structs.ReadTenant, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the tenant ID associated with the user
	if id == "" {
		id, _ = s.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	row, err := s.Find(ctx, id)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator of the tenant
	if types.ToValue(row.CreatedBy) != userID {
		return nil, errors.New("this tenant is not yours")
	}

	return row, nil
}

// GetByUser returns the tenant for the created by user
func (s *tenantService) GetByUser(ctx context.Context, uid string) (*structs.ReadTenant, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}
	tenant, err := s.tenant.GetByUser(ctx, uid)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}
	return s.Serialize(tenant), nil
}

// Find finds tenant service.
func (s *tenantService) Find(ctx context.Context, id string) (*structs.ReadTenant, error) {
	tenant, err := s.tenant.GetBySlug(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.Serialize(tenant), nil
}

// Delete deletes tenant service.
func (s *tenantService) Delete(ctx context.Context, id string) error {
	err := s.tenant.Delete(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Freed all roles / groups / users that are associated with the tenant

	return nil
}

// List lists tenant service.
func (s *tenantService) List(ctx context.Context, params *structs.ListTenantParams) (*types.JSON, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return nil, errors.New(ecode.FieldIsInvalid("limit"))
	}

	rows, err := s.tenant.List(ctx, params)
	if err := handleEntError("Tenant", err); err != nil {
		return nil, err
	}

	total := s.tenant.CountX(ctx, params)

	return &types.JSON{
		"content": s.Serializes(rows),
		"total":   total,
	}, nil
}

// CreateInitial creates the initial tenant, initializes roles, and user relationships.
func (s *tenantService) CreateInitial(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error) {
	// Create the default tenant
	tenant, err := s.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	// Get or create the super admin role
	superAdminRole, err := s.as.Role.Find(ctx, "super-admin")
	if ent.IsNotFound(err) {
		// Super admin role does not exist, create it
		superAdminRole, err = s.as.Role.CreateSuperAdminRole(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Assign the user to the default tenant with the super admin role
	if body.CreatedBy != nil {
		_, err = s.userTenant.Create(ctx, &structs.UserTenant{UserID: *body.CreatedBy, TenantID: tenant.ID})
		if err != nil {
			return nil, err
		}

		// Assign the tenant to the super admin role
		_, err = s.as.UserTenantRole.AddRoleToUserInTenant(ctx, *body.CreatedBy, superAdminRole.ID, tenant.ID)
		if err != nil {
			return nil, err
		}

		// Assign the super admin role to the user
		if err = s.as.UserRole.AddRoleToUser(ctx, *body.CreatedBy, superAdminRole.ID); err != nil {
			return nil, err
		}
	}

	return tenant, nil
}

// IsCreate checks if a tenant needs to be created and initializes tenant, roles, and user relationships if necessary.
func (s *tenantService) IsCreate(ctx context.Context, body *structs.CreateTenantBody) (*structs.ReadTenant, error) {
	if body.CreatedBy == nil {
		return nil, errors.New("invalid user ID")
	}

	// If slug is not provided, generate it
	if body.Slug == "" && body.Name != "" {
		body.Slug = slug.Unicode(body.Name)
	}

	// Check the number of existing users
	countUsers := s.us.User.CountX(ctx, &userStructs.ListUserParams{})

	// If there are no existing users, create the initial tenant
	if countUsers <= 1 {
		return s.CreateInitial(ctx, body)
	}

	// If there are existing users, check if the user already has a tenant
	existingTenant, err := s.GetByUser(ctx, *body.CreatedBy)
	if ent.IsNotFound(err) {
		// No existing tenant found for the user, proceed with tenant creation
	} else if err != nil {
		return nil, err
	} else {
		// If the user already has a tenant, return the existing tenant
		return existingTenant, nil
	}

	// If there are no existing tenants and body.Tenant is not empty, create the initial tenant
	if body.TenantBody.Name != "" {
		return s.CreateInitial(ctx, body)
	}

	return nil, nil

}

// Serializes serialize tenants
func (s *tenantService) Serializes(rows []*ent.Tenant) []*structs.ReadTenant {
	tenants := make([]*structs.ReadTenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, s.Serialize(row))
	}
	return tenants
}

// Serialize serialize a tenant
func (s *tenantService) Serialize(row *ent.Tenant) *structs.ReadTenant {
	return &structs.ReadTenant{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Type:        row.Type,
		Title:       row.Title,
		URL:         row.URL,
		Logo:        row.Logo,
		LogoAlt:     row.LogoAlt,
		Keywords:    row.Keywords,
		Copyright:   row.Copyright,
		Description: row.Description,
		Order:       &row.Order,
		Disabled:    row.Disabled,
		Extras:      &row.Extras,
		ExpiredAt:   &row.ExpiredAt,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
