package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/repository"
	"ncobase/feature/tenant/structs"
)

// UserTenantServiceInterface is the interface for the service.
type UserTenantServiceInterface interface {
	UserBelongTenantService(ctx context.Context, uid string) (*resp.Exception, error)
	UserBelongTenantsService(ctx context.Context, uid string) (*resp.Exception, error)
	AddUserToTenantService(ctx context.Context, u string, t string) (*resp.Exception, error)
	RemoveUserFromTenantService(ctx context.Context, u string, t string) (*resp.Exception, error)
}

// userTenantService is the struct for the service.
type userTenantService struct {
	ts             TenantServiceInterface
	userTenant     repository.UserTenantRepositoryInterface
	userTenantRole repository.UserTenantRoleRepositoryInterface
}

// NewUserTenantService creates a new service.
func NewUserTenantService(d *data.Data, ts TenantServiceInterface) UserTenantServiceInterface {
	return &userTenantService{
		ts:             ts,
		userTenant:     repository.NewUserTenantRepository(d),
		userTenantRole: repository.NewUserTenantRoleRepository(d),
	}
}

// UserBelongTenantService user belong tenant service
func (s *userTenantService) UserBelongTenantService(ctx context.Context, uid string) (*resp.Exception, error) {
	if uid == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("User ID")), nil
	}

	userTenant, err := s.userTenant.GetByUserID(ctx, uid)
	if exception, err := handleEntError("UserTenant", err); exception != nil {
		return exception, err
	}

	tenant, err := s.ts.FindTenantService(ctx, userTenant.TenantID)
	if exception, err := handleEntError("Tenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: tenant,
	}, nil
}

// UserBelongTenantsService user belong tenants service
func (s *userTenantService) UserBelongTenantsService(ctx context.Context, uid string) (*resp.Exception, error) {
	if uid == "" {
		return resp.BadRequest(ecode.FieldIsInvalid("User ID")), nil
	}

	userTenants, err := s.userTenant.GetTenantsByUserID(ctx, uid)
	if exception, err := handleEntError("UserTenants", err); exception != nil {
		return exception, err
	}

	var tenants []*structs.ReadTenant
	for _, userTenant := range userTenants {
		tenant, err := s.ts.FindTenantService(ctx, userTenant.ID)
		if exception, err := handleEntError("Tenant", err); exception != nil {
			return exception, err
		}
		tenants = append(tenants, tenant)
	}

	return &resp.Exception{
		Data: tenants,
	}, nil
}

// AddUserToTenantService adds a user to a tenant.
func (s *userTenantService) AddUserToTenantService(ctx context.Context, u string, t string) (*resp.Exception, error) {
	_, err := s.userTenant.Create(ctx, &structs.UserTenant{UserID: u, TenantID: t})
	if exception, err := handleEntError("UserTenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User added to tenant successfully",
	}, nil
}

// RemoveUserFromTenantService removes a user from a tenant.
func (s *userTenantService) RemoveUserFromTenantService(ctx context.Context, u string, t string) (*resp.Exception, error) {
	err := s.userTenant.Delete(ctx, u, t)
	if exception, err := handleEntError("UserTenant", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "User removed from tenant successfully",
	}, nil
}
