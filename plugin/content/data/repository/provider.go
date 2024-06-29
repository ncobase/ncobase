package repository

import (
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/data/repository/taxonomy"
	"ncobase/plugin/content/data/repository/topic"
)

type Repository struct {
	Taxonomy          taxonomy.RepositoryInterface
	TaxonomyRelations taxonomy.RelationRepositoryInterface
	Topic             topic.RepositoryInterface
}

func New(d *data.Data) *Repository {
	return &Repository{
		Taxonomy:          taxonomy.NewTaxonomyRepo(d),
		TaxonomyRelations: taxonomy.NewTaxonomyRelationRepo(d),
		Topic:             topic.NewTopicRepo(d),
	}
}
