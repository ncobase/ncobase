package service

import (
	"context"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"

	"ncobase/common/resp"
	"ncobase/common/types"
)

// CreatePermissionService creates a new permission.
func (svc *Service) CreatePermissionService(ctx context.Context, permissionData *structs.CreatePermissionBody) (*resp.Exception, error) {
	if permissionData.Name == "" {
		return resp.BadRequest("Permission name is required"), nil
	}

	permission, err := svc.permission.Create(ctx, permissionData)
	if exception, err := handleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// UpdatePermissionService updates an existing permission.
func (svc *Service) UpdatePermissionService(ctx context.Context, permissionID string, updates types.JSON) (*resp.Exception, error) {
	permission, err := svc.permission.Update(ctx, permissionID, updates)
	if exception, err := handleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// GetPermissionByIDService retrieves a permission by its ID.
func (svc *Service) GetPermissionByIDService(ctx context.Context, permissionID string) (*resp.Exception, error) {
	permission, err := svc.permission.GetByID(ctx, permissionID)
	if exception, err := handleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// DeletePermissionService deletes a permission by its ID.
func (svc *Service) DeletePermissionService(ctx context.Context, permissionID string) (*resp.Exception, error) {
	err := svc.permission.Delete(ctx, permissionID)
	if exception, err := handleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission deleted successfully",
	}, nil
}

// GetPermissionsByRoleIDService retrieves all permissions associated with a role.
func (svc *Service) GetPermissionsByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	permissions, err := svc.rolePermission.GetPermissionsByRoleID(ctx, roleID)
	if exception, err := handleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: permissions,
	}, nil
}

// ****** Internal methods of service

// serializePermissions serializes a list of permission entities to a response format.
func (svc *Service) serializePermissions(permissions []*ent.Permission) []*structs.ReadPermission {
	serializedPermissions := make([]*structs.ReadPermission, len(permissions))
	for i, permission := range permissions {
		serializedPermissions[i] = svc.serializePermission(permission)
	}
	return serializedPermissions
}

// serializePermission serializes a permission entity to a response format.
func (svc *Service) serializePermission(row *ent.Permission) *structs.ReadPermission {
	return &structs.ReadPermission{
		ID:          row.ID,
		Name:        row.Name,
		Action:      row.Action,
		Subject:     row.Subject,
		Description: row.Description,
		Default:     &row.Default,
		Disabled:    &row.Disabled,
		Extras:      &row.Extras,
		BaseEntity: structs.BaseEntity{
			CreatedBy: &row.CreatedBy,
			CreatedAt: &row.CreatedAt,
			UpdatedBy: &row.UpdatedBy,
			UpdatedAt: &row.UpdatedAt},
	}
}
