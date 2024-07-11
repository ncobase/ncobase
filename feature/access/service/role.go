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
	"ncobase/feature/access/structs"
)

// RoleServiceInterface is the interface for the service.
type RoleServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateRoleBody) (*resp.Exception, error)
	Update(ctx context.Context, roleID string, updates types.JSON) (*resp.Exception, error)
	Delete(ctx context.Context, roleID string) (*resp.Exception, error)
	GetByID(ctx context.Context, roleID string) (*resp.Exception, error)
	Find(ctx context.Context, r string) (*structs.ReadRole, error)
	CreateSuperAdminRole(ctx context.Context) (*structs.ReadRole, error)
	ListRoles(ctx context.Context, params *structs.ListRoleParams) (*resp.Exception, error)
	SerializeRole(row *ent.Role) *structs.ReadRole
	SerializeRoles(rows []*ent.Role) []*structs.ReadRole
}

// roleService is the struct for the service.
type roleService struct {
	ps   PermissionServiceInterface
	role repository.RoleRepositoryInterface
}

// NewRoleService creates a new service.
func NewRoleService(d *data.Data, ps PermissionServiceInterface) RoleServiceInterface {
	return &roleService{
		ps:   ps,
		role: repository.NewRoleRepository(d),
	}
}

// Create creates a new role.
func (s *roleService) Create(ctx context.Context, body *structs.CreateRoleBody) (*resp.Exception, error) {
	if body.Name == "" {
		return resp.BadRequest("Role name is required"), nil
	}

	role, err := s.role.Create(ctx, body)
	if exception, err := handleEntError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeRole(role),
	}, nil
}

// Update updates an existing role.
func (s *roleService) Update(ctx context.Context, roleID string, updates types.JSON) (*resp.Exception, error) {
	role, err := s.role.Update(ctx, roleID, updates)
	if exception, err := handleEntError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeRole(role),
	}, nil
}

// GetByID retrieves a role by its ID.
func (s *roleService) GetByID(ctx context.Context, roleID string) (*resp.Exception, error) {
	role, err := s.role.GetByID(ctx, roleID)
	if exception, err := handleEntError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeRole(role),
	}, nil
}

// Find finds a role by id or slug.
func (s *roleService) Find(ctx context.Context, r string) (*structs.ReadRole, error) {
	row, err := s.role.FindRole(ctx, &structs.FindRole{Slug: r})
	if err != nil {
		return nil, err
	}
	return s.SerializeRole(row), nil
}

// Delete deletes a role by its ID.
func (s *roleService) Delete(ctx context.Context, roleID string) (*resp.Exception, error) {
	err := s.role.Delete(ctx, roleID)
	if exception, err := handleEntError("Role", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role deleted successfully",
	}, nil
}

// CreateSuperAdminRole creates a new super admin role.
func (s *roleService) CreateSuperAdminRole(ctx context.Context) (*structs.ReadRole, error) {
	row, err := s.role.Create(ctx, &structs.CreateRoleBody{
		RoleBody: structs.RoleBody{
			Name:        "Super Admin",
			Slug:        "super-admin",
			Disabled:    false,
			Description: "Super Administrator role with all permissions",
			Extras:      nil,
		},
	})
	if err != nil {
		return nil, err
	}
	return s.SerializeRole(row), nil
}

// ListRoles lists all roles.
func (s *roleService) ListRoles(ctx context.Context, params *structs.ListRoleParams) (*resp.Exception, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return resp.BadRequest(ecode.FieldIsInvalid("limit")), nil
	}

	roles, err := s.role.List(ctx, params)
	if exception, err := handleEntError("Role", err); exception != nil {
		return exception, err
	}

	total := s.role.CountX(ctx, params)

	return &resp.Exception{
		Data: types.JSON{
			"content": roles,
			"total":   total,
		},
	}, nil
}

// SerializeRoles serializes a list of role entities to a response format.
func (s *roleService) SerializeRoles(rows []*ent.Role) []*structs.ReadRole {
	roles := make([]*structs.ReadRole, len(rows))
	for i, row := range rows {
		roles[i] = s.SerializeRole(row)
	}
	return roles
}

// SerializeRole serializes a role entity to a response format.
func (s *roleService) SerializeRole(row *ent.Role) *structs.ReadRole {
	return &structs.ReadRole{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Disabled:    row.Disabled,
		Description: row.Description,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
