package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	"ncobase/feature/access/data/repository"
	"ncobase/feature/access/structs"
)

// PermissionServiceInterface is the interface for the service.
type PermissionServiceInterface interface {
	Create(ctx context.Context, permissionData *structs.CreatePermissionBody) (*structs.ReadPermission, error)
	Update(ctx context.Context, permissionID string, updates types.JSON) (*structs.ReadPermission, error)
	Delete(ctx context.Context, permissionID string) error
	GetByID(ctx context.Context, permissionID string) (*structs.ReadPermission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*structs.ReadPermission, error)
	List(ctx context.Context, params *structs.ListPermissionParams) (paging.Result[*structs.ReadPermission], error)
	CountX(ctx context.Context, params *structs.ListPermissionParams) int
	Serialize(row *ent.Permission) *structs.ReadPermission
	Serializes(rows []*ent.Permission) []*structs.ReadPermission
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
func (s *permissionService) Create(ctx context.Context, body *structs.CreatePermissionBody) (*structs.ReadPermission, error) {
	if body.Name == "" {
		return nil, errors.New("permission name is required")
	}

	row, err := s.permission.Create(ctx, body)
	if err := handleEntError("Permission", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing permission.
func (s *permissionService) Update(ctx context.Context, permissionID string, updates types.JSON) (*structs.ReadPermission, error) {
	row, err := s.permission.Update(ctx, permissionID, updates)
	if err := handleEntError("Permission", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByID retrieves a permission by its ID.
func (s *permissionService) GetByID(ctx context.Context, permissionID string) (*structs.ReadPermission, error) {
	row, err := s.permission.GetByID(ctx, permissionID)
	if err := handleEntError("Permission", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a permission by its ID.
func (s *permissionService) Delete(ctx context.Context, permissionID string) error {
	err := s.permission.Delete(ctx, permissionID)
	if err := handleEntError("Permission", err); err != nil {
		return err
	}
	return nil
}

// GetPermissionsByRoleID retrieves all permissions associated with a role.
func (s *permissionService) GetPermissionsByRoleID(ctx context.Context, roleID string) ([]*structs.ReadPermission, error) {
	rows, err := s.rolePermission.GetPermissionsByRoleID(ctx, roleID)
	if err := handleEntError("Permission", err); err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
}

// List lists all permissions.
func (s *permissionService) List(ctx context.Context, params *structs.ListPermissionParams) (paging.Result[*structs.ReadPermission], error) {
	pp := paging.Params{
		Cursor: params.Cursor,
		Limit:  params.Limit,
	}

	rs, err := paging.Paginate(pp, func(cursor string, limit int) ([]*structs.ReadPermission, int, string, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit

		rows, err := s.permission.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, "", errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if validator.IsNotNil(err) {
			log.Errorf(ctx, "Error listing permissions: %v\n", err)
			return nil, 0, "", err
		}
		if err != nil {
			return nil, 0, "", err
		}

		total := s.permission.CountX(ctx, params)

		var nextCursor string
		if len(rows) > 0 {
			nextCursor = paging.EncodeCursor(rows[len(rows)-1].CreatedAt)
		}

		return s.Serializes(rows), total, nextCursor, nil
	})

	return *rs, err
}

// CountX gets a count of permissions.
func (s *permissionService) CountX(ctx context.Context, params *structs.ListPermissionParams) int {
	return s.permission.CountX(ctx, params)
}

// Serializes serializes a list of permission entities to a response format.
func (s *permissionService) Serializes(rows []*ent.Permission) []*structs.ReadPermission {
	rs := make([]*structs.ReadPermission, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes a permission entity to a response format.
func (s *permissionService) Serialize(row *ent.Permission) *structs.ReadPermission {
	return &structs.ReadPermission{
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
