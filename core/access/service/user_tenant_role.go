package service

import (
	"context"
	"ncobase/core/access/data"
	"ncobase/core/access/data/ent"
	"ncobase/core/access/data/repository"
	"ncobase/core/access/structs"
)

// UserTenantRoleServiceInterface is the interface for the service.
type UserTenantRoleServiceInterface interface {
	AddRoleToUserInTenant(ctx context.Context, u, t, r string) (*structs.UserTenantRole, error)
	GetUserRolesInTenant(ctx context.Context, u, t string) ([]string, error)
	RemoveRoleFromUserInTenant(ctx context.Context, u, t, r string) error
}

// userTenantRoleService is the struct for the service.
type userTenantRoleService struct {
	userTenantRole repository.UserTenantRoleRepositoryInterface
}

// NewUserTenantRoleService creates a new service.
func NewUserTenantRoleService(d *data.Data) UserTenantRoleServiceInterface {
	return &userTenantRoleService{
		userTenantRole: repository.NewUserTenantRoleRepository(d),
	}
}

// AddRoleToUserInTenant adds a role to a user in a tenant.
func (s *userTenantRoleService) AddRoleToUserInTenant(ctx context.Context, u, t, r string) (*structs.UserTenantRole, error) {
	row, err := s.userTenantRole.Create(ctx, &structs.UserTenantRole{UserID: u, TenantID: t, RoleID: r})
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return nil, err
	}
	return s.SerializeUserTenantRole(row), nil
}

// GetUserRolesInTenant retrieves all roles associated with a user in a tenant.
func (s *userTenantRoleService) GetUserRolesInTenant(ctx context.Context, u string, t string) ([]string, error) {
	roleIDs, err := s.userTenantRole.GetRolesByUserAndTenant(ctx, u, t)
	if err != nil {
		return nil, err
	}
	return roleIDs, nil
}

// RemoveRoleFromUserInTenant removes a role from a user in a tenant.
func (s *userTenantRoleService) RemoveRoleFromUserInTenant(ctx context.Context, u, t, r string) error {
	err := s.userTenantRole.DeleteByUserIDAndTenantIDAndRoleID(ctx, u, t, r)
	if err := handleEntError(ctx, "UserTenantRole", err); err != nil {
		return err
	}
	return nil
}

// SerializeUserTenantRole serializes a user tenant role.
func (s *userTenantRoleService) SerializeUserTenantRole(row *ent.UserTenantRole) *structs.UserTenantRole {
	return &structs.UserTenantRole{
		UserID:   row.UserID,
		TenantID: row.TenantID,
		RoleID:   row.RoleID,
	}
}
