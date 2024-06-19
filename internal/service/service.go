package service

import (
	"context"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	repo "ncobase/internal/data/repository"

	"github.com/ncobase/common/ecode"
	"github.com/ncobase/common/log"
	"github.com/ncobase/common/resp"
	"github.com/ncobase/common/validator"
)

// Service represents a service definition.
type Service struct {
	d                 *data.Data
	captcha           repo.Captcha
	tenant            repo.Tenant
	menu              repo.Menu
	user              repo.User
	userProfile       repo.UserProfile
	userRole          repo.UserRole
	userTenant        repo.UserTenant
	userTenantRole    repo.UserTenantRole
	userGroup         repo.UserGroup
	group             repo.Group
	groupRole         repo.GroupRole
	role              repo.Role
	permission        repo.Permission
	rolePermission    repo.RolePermission
	asset             repo.Asset
	module            repo.Module
	casbinRule        repo.CasbinRule
	taxonomy          repo.Taxonomy
	taxonomyRelations repo.TaxonomyRelation
	topic             repo.Topic
}

// New creates a Service instance and returns it.
func New(d *data.Data) *Service {
	return &Service{
		d:                 d,
		captcha:           repo.NewCaptcha(d),
		tenant:            repo.NewTenant(d),
		menu:              repo.NewMenu(d),
		user:              repo.NewUser(d),
		userProfile:       repo.NewUserProfile(d),
		userRole:          repo.NewUserRole(d),
		userTenant:        repo.NewUserTenant(d),
		userTenantRole:    repo.NewUserTenantRole(d),
		userGroup:         repo.NewUserGroup(d),
		group:             repo.NewGroup(d),
		groupRole:         repo.NewGroupRole(d),
		role:              repo.NewRole(d),
		permission:        repo.NewPermission(d),
		rolePermission:    repo.NewRolePermission(d),
		asset:             repo.NewAsset(d),
		module:            repo.NewModule(d),
		casbinRule:        repo.NewCasbinRule(d),
		taxonomy:          repo.NewTaxonomy(d),
		taxonomyRelations: repo.NewTaxonomyRelation(d),
		topic:             repo.NewTopic(d),
	}
}

// Ping check server
func (svc *Service) Ping(ctx context.Context) error {
	return svc.d.Ping(ctx)
}

// handleError is a helper function to handle errors in a consistent manner.
func handleError(k string, err error) (*resp.Exception, error) {
	if ent.IsNotFound(err) {
		log.Errorf(context.Background(), "Error not found in %s: %v\n", k, err)
		return resp.NotFound(ecode.NotExist(k)), nil
	}
	if ent.IsConstraintError(err) {
		log.Errorf(context.Background(), "Error constraint in %s: %v\n", k, err)
		return resp.Conflict(ecode.AlreadyExist(k)), nil
	}
	if ent.IsNotSingular(err) {
		log.Errorf(context.Background(), "Error not singular in %s: %v\n", k, err)
		return resp.BadRequest(ecode.NotSingular(k)), nil
	}
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "Error internal in %s: %v\n", k, err)
		return resp.InternalServer(err.Error()), nil
	}
	return nil, err
}
