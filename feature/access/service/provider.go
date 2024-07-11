package service

import "ncobase/feature/access/data"

// Service represents the auth service.
type Service struct {
	Role           RoleServiceInterface
	Permission     PermissionServiceInterface
	RolePermission RolePermissionServiceInterface
	UserRole       UserRoleServiceInterface
	Casbin         CasbinServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	ps := NewPermissionService(d)
	rs := NewRoleService(d, ps)
	return &Service{
		Role:           rs,
		Permission:     ps,
		RolePermission: NewRolePermissionService(d, ps),
		UserRole:       NewUserRoleService(d, rs),
		Casbin:         NewCasbinService(d),
	}
}
