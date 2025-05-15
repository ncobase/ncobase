package service

import (
	"context"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	"ncobase/access/data/repository"
	"ncobase/access/structs"

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
	ps             PermissionServiceInterface
	rolePermission repository.RolePermissionRepositoryInterface
}

// NewRolePermissionService creates a new service.
func NewRolePermissionService(d *data.Data, ps PermissionServiceInterface) RolePermissionServiceInterface {
	return &rolePermissionService{
		ps:             ps,
		rolePermission: repository.NewRolePermissionRepository(d),
	}
}

// AddPermissionToRole adds a permission to a role.
func (s *rolePermissionService) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) (*structs.RolePermission, error) {
	row, err := s.rolePermission.Create(ctx, &structs.RolePermission{RoleID: roleID, PermissionID: permissionID})
	if err := handleEntError(ctx, "RolePermission", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
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

	return s.ps.Serializes(permissions), nil
}

// Serializes serializes role permissions.
func (s *rolePermissionService) Serializes(rows []*ent.RolePermission) []*structs.RolePermission {
	var rs []*structs.RolePermission
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a role permission.
func (s *rolePermissionService) Serialize(row *ent.RolePermission) *structs.RolePermission {
	return &structs.RolePermission{
		RoleID:       row.RoleID,
		PermissionID: row.PermissionID,
	}
}
