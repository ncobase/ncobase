package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/repository"
	structs2 "ncobase/feature/access/structs"
)

// RolePermissionServiceInterface is the interface for the service.
type RolePermissionServiceInterface interface {
	AddPermissionToRole(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error)
	RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error)
	GetRolePermissions(ctx context.Context, r string) (*resp.Exception, error)
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
func (s *rolePermissionService) AddPermissionToRole(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error) {
	_, err := s.rolePermission.Create(ctx, &structs2.RolePermission{RoleID: roleID, PermissionID: permissionID})
	if exception, err := handleEntError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission added to role successfully",
	}, nil
}

// RemovePermissionFromRole removes a permission from a role.
func (s *rolePermissionService) RemovePermissionFromRole(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error) {
	err := s.rolePermission.Delete(ctx, roleID, permissionID)
	if exception, err := handleEntError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission removed from role successfully",
	}, nil
}

// GetRolePermissions retrieves permissions associated with a role.
func (s *rolePermissionService) GetRolePermissions(ctx context.Context, r string) (*resp.Exception, error) {
	permissions, err := s.rolePermission.GetPermissionsByRoleID(ctx, r)
	if exception, err := handleEntError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.ps.SerializePermissions(permissions),
	}, nil
}
