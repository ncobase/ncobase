package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	"ncobase/feature/access/data/repository"
	structs2 "ncobase/feature/access/structs"
)

// PermissionServiceInterface is the interface for the service.
type PermissionServiceInterface interface {
	Create(ctx context.Context, permissionData *structs2.CreatePermissionBody) (*resp.Exception, error)
	Update(ctx context.Context, permissionID string, updates types.JSON) (*resp.Exception, error)
	Delete(ctx context.Context, permissionID string) (*resp.Exception, error)
	GetByID(ctx context.Context, permissionID string) (*resp.Exception, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) (*resp.Exception, error)
	List(ctx context.Context, params *structs2.ListPermissionParams) (*resp.Exception, error)
	SerializePermission(row *ent.Permission) *structs2.ReadPermission
	SerializePermissions(rows []*ent.Permission) []*structs2.ReadPermission
}

// permissionService is the struct for the service.
type permissionService struct {
	permission     repository.PermissionRepositoryInterface
	rolePermission repository.RolePermissionRepositoryInterface
}

// NewPermissionService creates a new service.
func NewPermissionService(d *data.Data) PermissionServiceInterface {
	return &permissionService{
		permission:     repository.NewPermissionRepository(d),
		rolePermission: repository.NewRolePermissionRepository(d),
	}
}

// Create creates a new permission.
func (s *permissionService) Create(ctx context.Context, permissionData *structs2.CreatePermissionBody) (*resp.Exception, error) {
	if permissionData.Name == "" {
		return resp.BadRequest("Permission name is required"), nil
	}

	permission, err := s.permission.Create(ctx, permissionData)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializePermission(permission),
	}, nil
}

// Update updates an existing permission.
func (s *permissionService) Update(ctx context.Context, permissionID string, updates types.JSON) (*resp.Exception, error) {
	permission, err := s.permission.Update(ctx, permissionID, updates)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializePermission(permission),
	}, nil
}

// GetByID retrieves a permission by its ID.
func (s *permissionService) GetByID(ctx context.Context, permissionID string) (*resp.Exception, error) {
	permission, err := s.permission.GetByID(ctx, permissionID)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializePermission(permission),
	}, nil
}

// Delete deletes a permission by its ID.
func (s *permissionService) Delete(ctx context.Context, permissionID string) (*resp.Exception, error) {
	err := s.permission.Delete(ctx, permissionID)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Permission deleted successfully",
	}, nil
}

// GetPermissionsByRoleID retrieves all permissions associated with a role.
func (s *permissionService) GetPermissionsByRoleID(ctx context.Context, roleID string) (*resp.Exception, error) {
	permissions, err := s.rolePermission.GetPermissionsByRoleID(ctx, roleID)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: permissions,
	}, nil
}

// List lists all permissions.
func (s *permissionService) List(ctx context.Context, params *structs2.ListPermissionParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	permissions, err := s.permission.List(ctx, params)
	if exception, err := handleEntError("Permission", err); exception != nil {
		return exception, err
	}

	total := s.permission.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": permissions,
			"total":   total,
		},
	}, nil
}

// SerializePermissions serializes a list of permission entities to a response format.
func (s *permissionService) SerializePermissions(permissions []*ent.Permission) []*structs2.ReadPermission {
	serializedPermissions := make([]*structs2.ReadPermission, len(permissions))
	for i, permission := range permissions {
		serializedPermissions[i] = s.SerializePermission(permission)
	}
	return serializedPermissions
}

// SerializePermission serializes a permission entity to a response format.
func (s *permissionService) SerializePermission(row *ent.Permission) *structs2.ReadPermission {
	return &structs2.ReadPermission{
		ID:          row.ID,
		Name:        row.Name,
		Action:      row.Action,
		Subject:     row.Subject,
		Description: row.Description,
		Default:     &row.Default,
		Disabled:    &row.Disabled,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
