package service

import (
	"ncobase/feature/content/data/repository"
)

// Service represents the content service.
type Service struct {
	Taxonomy         TaxonomyServiceInterface
	TaxonomyRelation TaxonomyRelationServiceInterface
	Topic            TopicServiceInterface
}

// New creates a new service.
func New(r *repository.Repository) *Service {
	return &Service{
		Taxonomy:         NewTaxonomyService(r),
		TaxonomyRelation: NewTaxonomyRelationService(r),
		Topic:            NewTopicService(r),
	}
}
