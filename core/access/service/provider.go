package service

import (
	"github.com/ncobase/ncore/pkg/config"
	"ncobase/core/access/data"
)

// Service represents the auth service.
type Service struct {
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
		Role:           rs,
		Permission:     ps,
		RolePermission: NewRolePermissionService(d, ps),
		UserRole:       NewUserRoleService(d, rs),
		UserTenantRole: NewUserTenantRoleService(d),
		Casbin:         NewCasbinService(d),
		CasbinAdapter:  NewCasbinAdapterService(conf, d),
	}
}
