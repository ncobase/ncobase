package service

import "ncobase/domain/content/data"

// Service represents the content service.
type Service struct {
	Taxonomy         TaxonomyServiceInterface
	TaxonomyRelation TaxonomyRelationServiceInterface
	Topic            TopicServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{
		Taxonomy:         NewTaxonomyService(d),
		TaxonomyRelation: NewTaxonomyRelationService(d),
		Topic:            NewTopicService(d),
	}
}
