package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	"ncobase/feature/access/data/repository"
	"ncobase/feature/access/structs"
)

// RoleServiceInterface is the interface for the service.
type RoleServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateRoleBody) (*structs.ReadRole, error)
	Update(ctx context.Context, roleID string, updates types.JSON) (*structs.ReadRole, error)
	Delete(ctx context.Context, roleID string) error
	GetByID(ctx context.Context, roleID string) (*structs.ReadRole, error)
	GetBySlug(ctx context.Context, roleSlug string) (*structs.ReadRole, error)
	GetByIDs(ctx context.Context, roleIDs []string) ([]*structs.ReadRole, error)
	Find(ctx context.Context, r string) (*structs.ReadRole, error)
	CreateSuperAdminRole(ctx context.Context) (*structs.ReadRole, error)
	List(ctx context.Context, params *structs.ListRoleParams) (types.JSON, error)
	CountX(ctx context.Context, params *structs.ListRoleParams) int
	Serialize(row *ent.Role) *structs.ReadRole
	Serializes(rows []*ent.Role) []*structs.ReadRole
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
func (s *roleService) Create(ctx context.Context, body *structs.CreateRoleBody) (*structs.ReadRole, error) {
	if body.Name == "" {
		return nil, errors.New("role name is required")
	}

	role, err := s.role.Create(ctx, body)
	if err := handleEntError("Role", err); err != nil {
		return nil, err
	}

	return s.Serialize(role), nil
}

// Update updates an existing role.
func (s *roleService) Update(ctx context.Context, roleID string, updates types.JSON) (*structs.ReadRole, error) {
	role, err := s.role.Update(ctx, roleID, updates)
	if err := handleEntError("Role", err); err != nil {
		return nil, err
	}
	return s.Serialize(role), nil
}

// GetByID retrieves a role by its ID.
func (s *roleService) GetByID(ctx context.Context, roleID string) (*structs.ReadRole, error) {
	row, err := s.role.GetByID(ctx, roleID)
	if err := handleEntError("Role", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetBySlug retrieves a role by its slug.
func (s *roleService) GetBySlug(ctx context.Context, roleSlug string) (*structs.ReadRole, error) {
	row, err := s.role.GetBySlug(ctx, roleSlug)
	if err := handleEntError("Role", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByIDs retrieves roles by their IDs.
func (s *roleService) GetByIDs(ctx context.Context, roleIDs []string) ([]*structs.ReadRole, error) {
	rows, err := s.role.GetByIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	return s.Serializes(rows), nil
}

// Find finds a role by id or slug.
func (s *roleService) Find(ctx context.Context, r string) (*structs.ReadRole, error) {
	row, err := s.role.FindRole(ctx, &structs.FindRole{Slug: r})
	if err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// Delete deletes a role by its ID.
func (s *roleService) Delete(ctx context.Context, roleID string) error {
	err := s.role.Delete(ctx, roleID)
	if err := handleEntError("Role", err); err != nil {
		return err
	}
	return nil
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
	return s.Serialize(row), nil
}

// List lists all roles.
func (s *roleService) List(ctx context.Context, params *structs.ListRoleParams) (types.JSON, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must be less than 100
	if params.Limit > 100 {
		return nil, errors.New(ecode.FieldIsInvalid("limit"))
	}

	rs, err := s.role.List(ctx, params)
	if err := handleEntError("Role", err); err != nil {
		return nil, err
	}

	total := s.role.CountX(ctx, params)

	return types.JSON{
		"content": s.Serializes(rs),
		"total":   total,
	}, nil
}

// CountX gets a count of roles.
func (s *roleService) CountX(ctx context.Context, params *structs.ListRoleParams) int {
	return s.role.CountX(ctx, params)
}

// Serializes serializes a list of role entities to a response format.
func (s *roleService) Serializes(rows []*ent.Role) []*structs.ReadRole {
	rs := make([]*structs.ReadRole, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes a role entity to a response format.
func (s *roleService) Serialize(row *ent.Role) *structs.ReadRole {
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
