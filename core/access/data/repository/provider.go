package repository

import "ncobase/core/access/data"

// Repository represents the access repository.
type Repository struct {
	casbin         CasbinRuleRepositoryInterface
	role           RoleRepositoryInterface
	permission     PermissionRepositoryInterface
	rolePermission RolePermissionRepositoryInterface
	userRole       UserRoleRepositoryInterface
	userTenantRole UserTenantRoleRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		casbin:         NewCasbinRule(d),
		role:           NewRoleRepository(d),
		permission:     NewPermissionRepository(d),
		rolePermission: NewRolePermissionRepository(d),
		userRole:       NewUserRoleRepository(d),
		userTenantRole: NewUserTenantRoleRepository(d),
	}
}
