package repository

import (
	"ncobase/domain/content/data"
)

// Repository represents the content repository.
type Repository struct {
	Taxonomy          TaxonomyRepositoryInterface
	TaxonomyRelations TaxonomyRelationsRepositoryInterface
	Topic             TopicRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Taxonomy:          NewTaxonomyRepository(d),
		TaxonomyRelations: NewTaxonomyRelationsRepository(d),
		Topic:             NewTopicRepository(d),
	}
}
