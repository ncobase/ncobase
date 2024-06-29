package service

import (
	"context"
	"ncobase/common/log"
	"ncobase/internal/data"
	repo "ncobase/internal/data/repository"
)

// Service represents a service definition.
type Service struct {
	d              *data.Data
	captcha        repo.Captcha
	tenant         repo.Tenant
	menu           repo.Menu
	user           repo.User
	userProfile    repo.UserProfile
	userRole       repo.UserRole
	userTenant     repo.UserTenant
	userTenantRole repo.UserTenantRole
	userGroup      repo.UserGroup
	group          repo.Group
	groupRole      repo.GroupRole
	role           repo.Role
	permission     repo.Permission
	rolePermission repo.RolePermission
	asset          repo.Asset
	module         repo.Module
	casbinRule     repo.CasbinRule
}

// New creates a Service instance and returns it.
func New(d *data.Data) *Service {
	svc := &Service{
		d:              d,
		captcha:        repo.NewCaptcha(d),
		tenant:         repo.NewTenant(d),
		menu:           repo.NewMenu(d),
		user:           repo.NewUser(d),
		userProfile:    repo.NewUserProfile(d),
		userRole:       repo.NewUserRole(d),
		userTenant:     repo.NewUserTenant(d),
		userTenantRole: repo.NewUserTenantRole(d),
		userGroup:      repo.NewUserGroup(d),
		group:          repo.NewGroup(d),
		groupRole:      repo.NewGroupRole(d),
		role:           repo.NewRole(d),
		permission:     repo.NewPermission(d),
		rolePermission: repo.NewRolePermission(d),
		asset:          repo.NewAsset(d),
		module:         repo.NewModule(d),
		casbinRule:     repo.NewCasbinRule(d),
	}

	if err := svc.initData(); err != nil {
		log.Fatalf(context.Background(), "❌ Failed initializing data: %+v", err)
	}

	return svc
}

// GetData returns the data.
func (svc *Service) GetData() *data.Data {
	return svc.d
}

// GetCasbinRuleRepo returns the casbin rule repository.
func (svc *Service) GetCasbinRuleRepo() repo.CasbinRule {
	return svc.casbinRule
}

// Ping check server
func (svc *Service) Ping(ctx context.Context) error {
	return svc.d.Ping(ctx)
}
