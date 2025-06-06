package service

import (
	"context"
	"errors"
	accessStructs "ncobase/access/structs"
	"ncobase/auth/data"
	"ncobase/auth/wrapper"
	tenantStructs "ncobase/tenant/structs"
	userStructs "ncobase/user/structs"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/slug"
)

// AuthTenantServiceInterface is the interface for the service.
type AuthTenantServiceInterface interface {
	CreateInitialTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error)
	IsCreateTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error)
}

// authTenantService is the struct for the service.
type authTenantService struct {
	d *data.Data

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.TenantServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewAuthTenantService creates a new service.
func NewAuthTenantService(d *data.Data, usw *wrapper.UserServiceWrapper, tsw *wrapper.TenantServiceWrapper, asw *wrapper.AccessServiceWrapper) AuthTenantServiceInterface {
	return &authTenantService{
		d:   d,
		usw: usw,
		tsw: tsw,
		asw: asw,
	}
}

// CreateInitialTenant creates the initial tenant and sets up roles and user relationships
func (s *authTenantService) CreateInitialTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error) {
	// Create the tenant
	tenant, err := s.tsw.CreateTenant(ctx, body)
	if err != nil {
		logger.Errorf(ctx, "Failed to create tenant: %v", err)
		return nil, err
	}

	if body.CreatedBy == nil {
		logger.Infof(ctx, "No user specified for tenant creation, skipping role assignment")
		return tenant, nil
	}

	// Get or create the super admin role
	superAdminRole, err := s.getSuperAdminRole(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to get super admin role: %v", err)
		return nil, err
	}

	// Add user to tenant
	if _, err := s.tsw.AddUserToTenant(ctx, *body.CreatedBy, tenant.ID); err != nil {
		logger.Errorf(ctx, "Failed to add user to tenant: %v", err)
		return nil, err
	}

	// Assign global role to user
	if err := s.asw.AddRoleToUser(ctx, *body.CreatedBy, superAdminRole.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign global role to user: %v", err)
	}

	// Assign tenant-specific role
	if _, err := s.tsw.AddRoleToUserInTenant(ctx, *body.CreatedBy, tenant.ID, superAdminRole.ID); err != nil {
		logger.Warnf(ctx, "Failed to assign tenant role to user: %v", err)
	}

	logger.Infof(ctx, "Successfully created initial tenant '%s' with user assignments", tenant.Name)
	return tenant, nil
}

// IsCreateTenant checks conditions and creates tenant if needed
func (s *authTenantService) IsCreateTenant(ctx context.Context, body *tenantStructs.CreateTenantBody) (*tenantStructs.ReadTenant, error) {
	if body.CreatedBy == nil {
		return nil, errors.New("user ID is required for tenant creation")
	}

	// Generate slug if not provided
	if body.Slug == "" && body.Name != "" {
		body.Slug = slug.Unicode(body.Name)
	}

	// Check if user already has a tenant
	existingTenant, err := s.tsw.GetTenantByUser(ctx, *body.CreatedBy)
	if err == nil && existingTenant != nil {
		logger.Infof(ctx, "User already has tenant '%s'", existingTenant.Name)
		return existingTenant, nil
	}

	// Check if this is system initialization (first user or explicitly requested)
	userCount := s.usw.CountX(ctx, &userStructs.ListUserParams{})
	shouldCreateInitial := userCount <= 1 || body.Name != ""

	if shouldCreateInitial {
		logger.Infof(ctx, "Creating initial tenant for user (user count: %d)", userCount)
		return s.CreateInitialTenant(ctx, body)
	}

	logger.Infof(ctx, "Tenant creation conditions not met")
	return nil, nil
}

// getSuperAdminRole gets or creates super admin role
func (s *authTenantService) getSuperAdminRole(ctx context.Context) (*accessStructs.ReadRole, error) {
	// Try to find existing super admin role
	role, err := s.asw.FindRole(ctx, &accessStructs.FindRole{
		Slug: "super-admin",
	})
	if err == nil && role != nil {
		return role, nil
	}

	// Try system admin as fallback
	role, err = s.asw.FindRole(ctx, &accessStructs.FindRole{
		Slug: "system-admin",
	})
	if err == nil && role != nil {
		return role, nil
	}

	// Create new super admin role if none exists
	logger.Infof(ctx, "Creating new super admin role")
	return s.asw.CreateSuperAdminRole(ctx)
}
