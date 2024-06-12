package service

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	repo "stocms/internal/data/repository"
	"stocms/pkg/ecode"
	"stocms/pkg/log"
	"stocms/pkg/resp"
	"stocms/pkg/validator"
)

// Service represents a service definition.
type Service struct {
	d                 *data.Data
	domain            repo.Domain
	user              repo.User
	userProfile       repo.UserProfile
	userRole          repo.UserRole
	userDomain        repo.UserDomain
	userDomainRole    repo.UserDomainRole
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
		domain:            repo.NewDomain(d),
		user:              repo.NewUser(d),
		userProfile:       repo.NewUserProfile(d),
		userRole:          repo.NewUserRole(d),
		userDomain:        repo.NewUserDomain(d),
		userDomainRole:    repo.NewUserDomainRole(d),
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
		log.Errorf(nil, "Error not found in %s: %v\n", k, err)
		return resp.NotFound(ecode.NotExist(k)), nil
	}
	if ent.IsConstraintError(err) {
		log.Errorf(nil, "Error constraint in %s: %v\n", k, err)
		return resp.Conflict(ecode.AlreadyExist(k)), nil
	}
	if validator.IsNotNil(err) {
		log.Errorf(nil, "Error internal in %s: %v\n", k, err)
		return resp.InternalServer(err.Error()), nil
	}
	return nil, err
}
