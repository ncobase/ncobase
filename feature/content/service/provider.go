package service

import (
	"ncobase/feature/content/data"
)

// Service represents the content service.
type Service struct {
	Taxonomy TaxonomyServiceInterface
	Topic    TopicServiceInterface
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{
		Taxonomy: NewTaxonomyService(d),
		Topic:    NewTopicService(d),
	}
}
