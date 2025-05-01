package repository

import (
	"ncobase/domain/content/data"
)

// Repository represents the content repository.
type Repository struct {
	Taxonomy          TaxonomyRepositoryInterface
	TaxonomyRelations TaxonomyRelationsRepositoryInterface
	Topic             TopicRepositoryInterface
	Channel           ChannelRepositoryInterface
	Distribution      DistributionRepositoryInterface
	Media             MediaRepositoryInterface
	TopicMedia        TopicMediaRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Taxonomy:          NewTaxonomyRepository(d),
		TaxonomyRelations: NewTaxonomyRelationsRepository(d),
		Topic:             NewTopicRepository(d),
		Channel:           NewChannelRepository(d),
		Distribution:      NewDistributionRepository(d),
		Media:             NewMediaRepository(d),
		TopicMedia:        NewTopicMediaRepository(d),
	}
}
