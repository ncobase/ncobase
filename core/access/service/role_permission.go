package service

import (
	"context"
	"ncobase/core/access/data"
	"ncobase/core/access/data/repository"
	"ncobase/core/access/structs"

	"github.com/ncobase/ncore/logging/logger"
)

// RolePermissionServiceInterface is the interface for the service.
type RolePermissionServiceInterface interface {
	AddPermissionToRole(ctx context.Context, roleID string, permissionID string) (*structs.RolePermission, error)
	RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error
	GetRolePermissions(ctx context.Context, r string) ([]*structs.ReadPermission, error)
}

// rolePermissionService is the struct for the service.
type rolePermissionService struct {
	rolePermission repository.RolePermissionRepositoryInterface
}

// NewRolePermissionService creates a new service.
func NewRolePermissionService(d *data.Data) RolePermissionServiceInterface {
	return &rolePermissionService{
		rolePermission: repository.NewRolePermissionRepository(d),
	}
}

// AddPermissionToRole adds a permission to a role.
func (s *rolePermissionService) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) (*structs.RolePermission, error) {
	row, err := s.rolePermission.Create(ctx, &structs.RolePermission{RoleID: roleID, PermissionID: permissionID})
	if err := handleEntError(ctx, "RolePermission", err); err != nil {
		return nil, err
	}

	return repository.SerializeRolePermission(row), nil
}

// RemovePermissionFromRole removes a permission from a role.
func (s *rolePermissionService) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) error {
	err := s.rolePermission.Delete(ctx, roleID, permissionID)
	if err := handleEntError(ctx, "RolePermission", err); err != nil {
		return err
	}

	return nil
}

// GetRolePermissions retrieves permissions associated with a role.
func (s *rolePermissionService) GetRolePermissions(ctx context.Context, r string) ([]*structs.ReadPermission, error) {
	permissions, err := s.rolePermission.GetPermissionsByRoleID(ctx, r)
	if err != nil {
		logger.Errorf(ctx, "rolePermissionRepo.GetRolePermissions error: %v", err)
		return nil, err
	}

	return repository.SerializePermissions(permissions), nil
}
