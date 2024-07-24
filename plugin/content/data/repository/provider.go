package repository

import "ncobase/plugin/content/data"

// Repository represents the content repository.
type Repository struct {
	Taxonomy          TaxonomyRepositoryInterface
	TaxonomyRelations TaxonomyRelationsRepositoryInterface
	Topic             TopicRepositoryInterface
}

// NewRepository creates a new repository.
func NewRepository(d *data.Data) *Repository {
	return &Repository{
		Taxonomy:          NewTaxonomyRepository(d),
		TaxonomyRelations: NewTaxonomyRelationsRepository(d),
		Topic:             NewTopicRepository(d),
	}
}
