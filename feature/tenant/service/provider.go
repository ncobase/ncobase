package service

import (
	accessService "ncobase/feature/access/service"
	"ncobase/feature/tenant/data"
	userService "ncobase/feature/user/service"
)

// Service represents the tenant service.
type Service struct {
	Tenant         TenantServiceInterface
	UserTenant     UserTenantServiceInterface
	UserTenantRole UserTenantRoleServiceInterface
}

// New creates a new service.
func New(d *data.Data, usi userService.UserServiceInterface, arsi accessService.RoleServiceInterface, aursi accessService.UserRoleServiceInterface) *Service {
	ts := NewTenantService(d, usi, arsi, aursi)
	uts := NewUserTenantService(d, ts)

	return &Service{
		Tenant:         ts,
		UserTenant:     uts,
		UserTenantRole: NewUserTenantRoleService(d, ts, uts),
	}
}
