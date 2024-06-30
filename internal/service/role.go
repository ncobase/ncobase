package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/helper"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
)

// CreateRoleService creates a new role.
func (svc *Service) CreateRoleService(ctx context.Context, body *structs.CreateRoleBody) (*resp.Exception, error) {
	if body.Name == "" {
		return resp.BadRequest("Role name is required"), nil
	}

	role, err := svc.role.Create(ctx, body)
	if exception, err := helper.HandleError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeRole(role),
	}, nil
}

// UpdateRoleService updates an existing role.
func (svc *Service) UpdateRoleService(ctx context.Context, roleID string, updates types.JSON) (*resp.Exception, error) {
	role, err := svc.role.Update(ctx, roleID, updates)
	if exception, err := helper.HandleError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeRole(role),
	}, nil
}

// GetRoleByIDService retrieves a role by its ID.
func (svc *Service) GetRoleByIDService(ctx context.Context, roleID string) (*resp.Exception, error) {
	role, err := svc.role.GetByID(ctx, roleID)
	if exception, err := helper.HandleError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeRole(role),
	}, nil
}

// DeleteRoleService deletes a role by its ID.
func (svc *Service) DeleteRoleService(ctx context.Context, roleID string) (*resp.Exception, error) {
	err := svc.role.Delete(ctx, roleID)
	if exception, err := helper.HandleError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role deleted successfully",
	}, nil
}

// AddPermissionToRoleService adds a permission to a role.
func (svc *Service) AddPermissionToRoleService(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error) {
	_, err := svc.rolePermission.Create(ctx, &structs.RolePermission{RoleID: roleID, PermissionID: permissionID})
	if exception, err := helper.HandleError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission added to role successfully",
	}, nil
}

// RemovePermissionFromRoleService removes a permission from a role.
func (svc *Service) RemovePermissionFromRoleService(ctx context.Context, roleID string, permissionID string) (*resp.Exception, error) {
	err := svc.rolePermission.Delete(ctx, roleID, permissionID)
	if exception, err := helper.HandleError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission removed from role successfully",
	}, nil
}

// GetRolePermissionsService retrieves permissions associated with a role.
func (svc *Service) GetRolePermissionsService(ctx context.Context, r string) (*resp.Exception, error) {
	permissions, err := svc.rolePermission.GetPermissionsByRoleID(ctx, r)
	if exception, err := helper.HandleError("RolePermission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializePermissions(permissions),
	}, nil
}

// ListRolesService lists all roles.
func (svc *Service) ListRolesService(ctx context.Context, params *structs.ListRoleParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	roles, err := svc.role.List(ctx, params)
	if exception, err := helper.HandleError("Role", err); exception != nil {
		return exception, err
	}

	total := svc.role.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": roles,
			"total":   total,
		},
	}, nil
}

// ****** Internal methods of service

// seializeRoles serializes a list of role entities to a response format.
func (svc *Service) serializeRoles(rows []*ent.Role) []*structs.ReadRole {
	roles := make([]*structs.ReadRole, len(rows))
	for i, row := range rows {
		roles[i] = svc.serializeRole(row)
	}
	return roles
}

// serializeRole serializes a role entity to a response format.
func (svc *Service) serializeRole(row *ent.Role) *structs.ReadRole {
	return &structs.ReadRole{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Disabled:    row.Disabled,
		Description: row.Description,
		Extras:      &row.Extras,
		BaseEntity: structs.BaseEntity{
			CreatedBy: &row.CreatedBy,
			CreatedAt: &row.CreatedAt,
			UpdatedBy: &row.UpdatedBy,
			UpdatedAt: &row.UpdatedAt,
		},
	}
}
