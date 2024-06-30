package service

import (
	"context"
	"ncobase/app/data/ent"
	"ncobase/app/data/structs"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/helper"
)

// CreatePermissionService creates a new permission.
func (svc *Service) CreatePermissionService(ctx context.Context, permissionData *structs.CreatePermissionBody) (*resp.Exception, error) {
	if permissionData.Name == "" {
		return resp.BadRequest("Permission name is required"), nil
	}

	permission, err := svc.permission.Create(ctx, permissionData)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// UpdatePermissionService updates an existing permission.
func (svc *Service) UpdatePermissionService(ctx context.Context, permissionID string, updates types.JSON) (*resp.Exception, error) {
	permission, err := svc.permission.Update(ctx, permissionID, updates)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// GetPermissionByIDService retrieves a permission by its ID.
func (svc *Service) GetPermissionByIDService(ctx context.Context, permissionID string) (*resp.Exception, error) {
	permission, err := svc.permission.GetByID(ctx, permissionID)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermission(permission),
	}, nil
}

// DeletePermissionService deletes a permission by its ID.
func (svc *Service) DeletePermissionService(ctx context.Context, permissionID string) (*resp.Exception, error) {
	err := svc.permission.Delete(ctx, permissionID)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission deleted successfully",
	}, nil
}

// GetPermissionsByRoleIDService retrieves all permissions associated with a role.
func (svc *Service) GetPermissionsByRoleIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	permissions, err := svc.rolePermission.GetPermissionsByRoleID(ctx, roleID)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: permissions,
	}, nil
}

// ListPermissionsService lists all permissions.
func (svc *Service) ListPermissionsService(ctx context.Context, params *structs.ListPermissionParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	permissions, err := svc.permission.List(ctx, params)
	if exception, err := helper.HandleError("Permission", err); exception != nil {
		return exception, err
	}

	total := svc.permission.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": permissions,
			"total":   total,
		},
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
