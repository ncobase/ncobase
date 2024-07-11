package repository

import "ncobase/feature/access/data"

// Repository represents the access repository.
type Repository struct {
	casbin         CasbinRuleRepositoryInterface
	role           RoleRepositoryInterface
	permission     PermissionRepositoryInterface
	rolePermission RolePermissionRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		casbin:         NewCasbinRule(d),
		role:           NewRoleRepository(d),
		permission:     NewPermissionRepository(d),
		rolePermission: NewRolePermissionRepository(d),
	}
}
