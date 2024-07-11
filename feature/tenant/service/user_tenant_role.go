package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/feature/tenant/data"
	"ncobase/feature/tenant/data/repository"
	"ncobase/feature/tenant/structs"
)

// UserTenantRoleServiceInterface is the interface for the service.
type UserTenantRoleServiceInterface interface {
	AddRoleToUserInTenantService(ctx context.Context, u string, r string, t string) (*resp.Exception, error)
	GetUserRolesInTenantService(ctx context.Context, u string, t string) ([]string, error)
	RemoveRoleFromUserInTenantService(ctx context.Context, u string, t string, r string) (*resp.Exception, error)
}

// userTenantRoleService is the struct for the service.
type userTenantRoleService struct {
	ts             TenantServiceInterface
	uts            UserTenantServiceInterface
	userTenantRole repository.UserTenantRoleRepositoryInterface
}

// NewUserTenantRoleService creates a new service.
func NewUserTenantRoleService(d *data.Data, ts TenantServiceInterface, uts UserTenantServiceInterface) UserTenantRoleServiceInterface {
	return &userTenantRoleService{
		ts:             ts,
		uts:            uts,
		userTenantRole: repository.NewUserTenantRoleRepository(d),
	}
}

// AddRoleToUserInTenantService adds a role to a user in a tenant.
func (s *userTenantRoleService) AddRoleToUserInTenantService(ctx context.Context, u string, t string, r string) (*resp.Exception, error) {
	_, err := s.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: u, TenantID: t, RoleID: r})
	if exception, err := handleEntError("UserTenantRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role added to user in tenant successfully",
	}, nil
}

// GetUserRolesInTenantService retrieves all roles associated with a user in a tenant.
func (s *userTenantRoleService) GetUserRolesInTenantService(ctx context.Context, u string, t string) ([]string, error) {
	roles, err := s.userTenantRole.GetRolesByUserAndTenant(ctx, u, t)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// RemoveRoleFromUserInTenantService removes a role from a user in a tenant.
func (s *userTenantRoleService) RemoveRoleFromUserInTenantService(ctx context.Context, u string, t string, r string) (*resp.Exception, error) {
	err := s.userTenantRole.DeleteByUserIDAndTenantIDAndRoleID(ctx, u, t, r)
	if exception, err := handleEntError("UserTenantRole", err); exception != nil {
		return exception, err
	}
	return &resp.Exception{
		Data: "Role removed from user in tenant successfully",
	}, nil
}
