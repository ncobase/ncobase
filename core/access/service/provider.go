package service

import (
	"ncobase/access/data"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/logging/logger"
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
	casbinAdapter := NewCasbinAdapterService(conf, d)

	// Initialize Casbin
	_, err := casbinAdapter.InitEnforcer()
	if err != nil {
		logger.Errorf(nil, "casbin enforcer initialization failed: %v", err)
	}

	return &Service{
		Activity:       NewActivityService(d),
		Role:           rs,
		Permission:     ps,
		RolePermission: NewRolePermissionService(d, ps),
		UserRole:       NewUserRoleService(d, rs),
		UserTenantRole: NewUserTenantRoleService(d),
		Casbin:         NewCasbinService(d),
		CasbinAdapter:  casbinAdapter,
	}
}
