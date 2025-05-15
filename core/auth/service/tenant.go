package service

import (
	"context"
	"errors"
	accessService "ncobase/access/service"
	"ncobase/auth/data"
	tenantService "ncobase/tenant/service"
	tenantStructs "ncobase/tenant/structs"
	userService "ncobase/user/service"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/utils/slug"
)

// AuthTenantServiceInterface is the interface for the service.
type AuthTenantServiceInterface interface {
	CreateInitialTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error)
	IsCreateTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error)
}

// authTenantService is the struct for the service.
type authTenantService struct {
	d  *data.Data
	as *accessService.Service
	ts *tenantService.Service
	us *userService.Service
}

// NewAuthTenantService creates a new service.
func NewAuthTenantService(d *data.Data, us *userService.Service, as *accessService.Service, ts *tenantService.Service) AuthTenantServiceInterface {
	return &authTenantService{
		d:  d,
		as: as,
		ts: ts,
		us: us,
	}
}

// CreateInitialTenant creates the initial tenant, initializes roles, and user relationships.
func (s *authTenantService) CreateInitialTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error) {
	// Create the default tenant
	tenant, err := s.ts.Tenant.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	// Get or create the super admin role
	superAdminRole, err := s.as.Role.Find(ctx, "super-admin")
	if superAdminRole == nil {
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
		_, err = s.ts.UserTenant.AddUserToTenant(ctx, *body.CreatedBy, tenant.ID)
		if err != nil {
			return nil, err
		}

		// Assign the tenant to the super admin role
		_, err = s.as.UserTenantRole.AddRoleToUserInTenant(ctx, *body.CreatedBy, tenant.ID, superAdminRole.ID)
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

// IsCreateTenant checks if a tenant needs to be created and initializes tenant, roles, and user relationships if necessary.
func (s *authTenantService) IsCreateTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error) {
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
		return s.CreateInitialTenant(ctx, body)
	}

	// If there are existing users, check if the user already has a tenant
	existingTenant, err := s.ts.Tenant.GetByUser(ctx, *body.CreatedBy)
	if existingTenant == nil {
		// No existing tenant found for the user, proceed with tenant creation
	} else if err != nil {
		return nil, err
	} else {
		// If the user already has a tenant, return the existing tenant
		return existingTenant, nil
	}

	// If there are no existing tenants and body.Tenant is not empty, create the initial tenant
	if body.TenantBody.Name != "" {
		return s.CreateInitialTenant(ctx, body)
	}

	return nil, nil

}
