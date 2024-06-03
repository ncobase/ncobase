package service

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	repo "stocms/internal/data/repository"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/validator"
)

// Service represents a service definition.
type Service struct {
	d                 *data.Data
	domain            repo.Domain
	user              repo.User
	taxonomy          repo.Taxonomy
	taxonomyRelations repo.TaxonomyRelations
	topic             repo.Topic
}

// New creates a Service instance and returns it.
func New(d *data.Data) *Service {
	return &Service{
		d:                 d,
		domain:            repo.NewDomain(d),
		user:              repo.NewUser(d),
		taxonomy:          repo.NewTaxonomy(d),
		taxonomyRelations: repo.NewTaxonomyRelations(d),
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
		return resp.NotFound(ecode.NotExist(k)), nil
	}
	if ent.IsConstraintError(err) {
		return resp.Conflict(ecode.AlreadyExist(k)), nil
	}
	if validator.IsNotNil(err) {
		return resp.InternalServer(err.Error()), nil
	}
	return nil, err
}
