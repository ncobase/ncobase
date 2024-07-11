package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/resp"
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
	AccountTenantService(ctx context.Context) (*resp.Exception, error)
	AccountTenantsService(ctx context.Context) (*resp.Exception, error)
	UserOwnTenantService(ctx context.Context, uid string) (*resp.Exception, error)
	CreateTenantService(ctx context.Context, body *structs.CreateTenantBody) (*resp.Exception, error)
	UpdateTenantService(ctx context.Context, body *structs.UpdateTenantBody) (*resp.Exception, error)
	GetTenantService(ctx context.Context, id string) (*resp.Exception, error)
	FindTenantService(ctx context.Context, id string) (*structs.ReadTenant, error)
	DeleteTenantService(ctx context.Context, id string) (*resp.Exception, error)
	ListTenantsService(ctx context.Context, params *structs.ListTenantParams) (*resp.Exception, error)
	CreateInitialTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error)
	IsCreateTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error)
	SerializeTenants(rows []*ent.Tenant) []*structs.ReadTenant
	SerializeTenant(tenant *ent.Tenant) *structs.ReadTenant
}

// tenantService is the struct for the service.
type tenantService struct {
	tenant         repository.TenantRepositoryInterface
	userTenant     repository.UserTenantRepositoryInterface
	userTenantRole repository.UserTenantRoleRepositoryInterface
	usi            userService.UserServiceInterface
	arsi           accessService.RoleServiceInterface
	aursi          accessService.UserRoleServiceInterface
}

// NewTenantService creates a new service.
func NewTenantService(d *data.Data, usi userService.UserServiceInterface, arsi accessService.RoleServiceInterface, aursi accessService.UserRoleServiceInterface) TenantServiceInterface {
	return &tenantService{
		tenant:         repository.NewTenantRepository(d),
		userTenant:     repository.NewUserTenantRepository(d),
		userTenantRole: repository.NewUserTenantRoleRepository(d),
		usi:            usi,
		arsi:           arsi,
		aursi:          aursi,
	}
}

// AccountTenantService retrieves the tenant associated with the user's account.
func (svc *tenantService) AccountTenantService(ctx context.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Retrieve the tenant associated with the user
	tenant, err := svc.tenant.GetByUser(ctx, userID)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: svc.SerializeTenant(tenant),
	}, nil
}

// AccountTenantsService retrieves the tenant associated with the user's account.
func (svc *tenantService) AccountTenantsService(ctx context.Context) (*resp.Exception, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	tenants, err := svc.ListTenantsService(ctx, &structs.ListTenantParams{
		User: userID,
	})
	if exception, err := handleEntError("Tenants", err); exception != nil {
		return exception, err
	}

	return tenants, nil
}

// UserOwnTenantService user own tenant service
func (svc *tenantService) UserOwnTenantService(ctx context.Context, uid string) (*resp.Exception, error) {
	if uid == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("User ID")), nil
	}

	tenant, err := svc.tenant.GetByUser(ctx, uid)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.SerializeTenant(tenant),
	}, nil
}

// CreateTenantService creates a tenant service.
func (svc *tenantService) CreateTenantService(ctx context.Context, body *structs.CreateTenantBody) (*resp.Exception, error) {
	if body.CreatedBy == nil {
		body.CreatedBy = types.ToPointer(helper.GetUserID(ctx))
	}

	// Create the tenant
	tenant, err := svc.IsCreateTenant(ctx, body)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.SerializeTenant(tenant),
	}, nil
}

// UpdateTenantService updates tenant service (full and partial).
func (svc *tenantService) UpdateTenantService(ctx context.Context, body *structs.UpdateTenantBody) (*resp.Exception, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the tenant
	if body.CreatedBy != nil {
		_, err := svc.tenant.GetByUser(ctx, *body.CreatedBy)
		if exception, err := handleEntError("Tenant", err); exception != nil {
			return exception, err
		}
	}

	// If ID is not provided, get the tenant ID associated with the user
	if body.ID == "" {
		body.ID, _ = svc.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	tenant, err := svc.FindTenantService(ctx, body.ID)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if types.ToValue(tenant.CreatedBy) != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

	// set updated by
	body.UpdatedBy = &userID

	// Serialize request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	// Unmarshal the JSON data into a types.JSON object
	var d types.JSON
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	// Update the tenant with the provided data
	_, err = svc.tenant.Update(ctx, tenant.ID, d)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	// Return success response
	return &resp.Exception{
		Data: types.JSON{
			"id": body.ID,
		},
	}, nil
}

// GetTenantService reads tenant service.
func (svc *tenantService) GetTenantService(ctx context.Context, id string) (*resp.Exception, error) {
	userID := helper.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the tenant ID associated with the user
	if id == "" {
		id, _ = svc.tenant.GetIDByUser(ctx, userID)
	}

	// Retrieve the tenant by ID
	tenant, err := svc.FindTenantService(ctx, id)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	// Check if the user is the creator of the tenant
	if types.ToValue(tenant.CreatedBy) != userID {
		return resp.Forbidden("This tenant is not yours"), nil
	}

	// Serialize tenant data and return
	return &resp.Exception{
		Data: tenant,
	}, nil
}

// FindTenantService finds tenant service.
func (svc *tenantService) FindTenantService(ctx context.Context, id string) (*structs.ReadTenant, error) {
	tenant, err := svc.tenant.GetBySlug(ctx, id)
	if err != nil {
		return nil, err
	}
	return svc.SerializeTenant(tenant), nil
}

// DeleteTenantService deletes tenant service.
func (svc *tenantService) DeleteTenantService(ctx context.Context, id string) (*resp.Exception, error) {
	err := svc.tenant.Delete(ctx, id)
	if err != nil {
		return resp.BadRequest(err.Error()), nil
	}

	// TODO: Freed all roles / groups / users that are associated with the tenant

	return &resp.Exception{
		Message: "Tenant deleted successfully",
	}, nil
}

// ListTenantsService lists tenant service.
func (svc *tenantService) ListTenantsService(ctx context.Context, params *structs.ListTenantParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	tenants, err := svc.tenant.List(ctx, params)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	total := svc.tenant.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": tenants,
			"total":   total,
		},
	}, nil
}

// CreateInitialTenant creates the initial tenant, initializes roles, and user relationships.
func (svc *tenantService) CreateInitialTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	// Create the default tenant
	defaultTenant, err := svc.tenant.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	// Get or create the super admin role
	superAdminRole, err := svc.arsi.Find(ctx, "super-admin")
	if ent.IsNotFound(err) {
		// Super admin role does not exist, create it
		superAdminRole, err = svc.arsi.CreateSuperAdminRole(ctx)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Assign the user to the default tenant with the super admin role
	if body.CreatedBy != nil {
		_, err = svc.userTenant.Create(ctx, &structs.UserTenant{UserID: *body.CreatedBy, TenantID: defaultTenant.ID})
		if err != nil {
			return nil, err
		}

		// Assign the tenant to the super admin role
		_, err = svc.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: *body.CreatedBy, RoleID: superAdminRole.ID, TenantID: defaultTenant.ID})
		if err != nil {
			return nil, err
		}

		// Assign the super admin role to the user
		_, err = svc.aursi.AddRoleToUserService(ctx, *body.CreatedBy, superAdminRole.ID)
		if err != nil {
			return nil, err
		}
	}

	return defaultTenant, nil
}

// IsCreateTenant checks if a tenant needs to be created and initializes tenant, roles, and user relationships if necessary.
func (svc *tenantService) IsCreateTenant(ctx context.Context, body *structs.CreateTenantBody) (*ent.Tenant, error) {
	if body.CreatedBy == nil {
		return nil, errors.New("invalid user ID")
	}

	// If slug is not provided, generate it
	if body.Slug == "" && body.Name != "" {
		body.Slug = slug.Unicode(body.Name)
	}

	// Check the number of existing users
	countUsers := svc.usi.CountX(ctx, &userStructs.ListUserParams{})

	// If there are no existing users, create the initial tenant
	if countUsers <= 1 {
		return svc.CreateInitialTenant(ctx, body)
	}

	// If there are existing users, check if the user already has a tenant
	existingTenant, err := svc.tenant.GetByUser(ctx, *body.CreatedBy)
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
		return svc.CreateInitialTenant(ctx, body)
	}

	return nil, nil

}

// SerializeTenants serialize tenants
func (svc *tenantService) SerializeTenants(rows []*ent.Tenant) []*structs.ReadTenant {
	tenants := make([]*structs.ReadTenant, 0, len(rows))
	for _, row := range rows {
		tenants = append(tenants, svc.SerializeTenant(row))
	}
	return tenants
}

// SerializeTenant serialize tenant
func (svc *tenantService) SerializeTenant(row *ent.Tenant) *structs.ReadTenant {
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
