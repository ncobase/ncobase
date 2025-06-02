package repository

import (
	"ncobase/access/data"
)

// Repository represents the access repository.
type Repository struct {
	Activity       ActivityRepositoryInterface
	Casbin         CasbinRuleRepositoryInterface
	Role           RoleRepositoryInterface
	Permission     PermissionRepositoryInterface
	RolePermission RolePermissionRepositoryInterface
	UserRole       UserRoleRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Activity:       NewActivityRepository(d),
		Casbin:         NewCasbinRule(d),
		Role:           NewRoleRepository(d),
		Permission:     NewPermissionRepository(d),
		RolePermission: NewRolePermissionRepository(d),
		UserRole:       NewUserRoleRepository(d),
	}
}
