package service

import (
	"ncobase/access/data"

	"github.com/ncobase/ncore/config"
)

// Service represents the auth service.
type Service struct {
	Activity       ActivityServiceInterface
	Role           RoleServiceInterface
	Permission     PermissionServiceInterface
	RolePermission RolePermissionServiceInterface
	UserRole       UserRoleServiceInterface
	UserTenantRole UserTenantRoleServiceInterface
	Casbin         CasbinServiceInterface
	CasbinAdapter  CasbinAdapterServiceInterface
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	ps := NewPermissionService(d)
	rs := NewRoleService(d, ps)
	return &Service{
		Activity:       NewActivityService(d),
		Role:           rs,
		Permission:     ps,
		RolePermission: NewRolePermissionService(d, ps),
		UserRole:       NewUserRoleService(d, rs),
		UserTenantRole: NewUserTenantRoleService(d),
		Casbin:         NewCasbinService(d),
		CasbinAdapter:  NewCasbinAdapterService(conf, d),
	}
}
