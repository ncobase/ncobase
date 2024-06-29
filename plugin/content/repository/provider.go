package repository

import (
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/repository/taxonomy"
	"ncobase/plugin/content/repository/topic"
)

type Repository struct {
	Taxonomy          taxonomy.ITaxonomy
	TaxonomyRelations taxonomy.ITaxonomyRelation
	Topic             topic.ITopic
}

func New(d *data.Data) *Repository {
	return &Repository{
		Taxonomy:          taxonomy.NewTaxonomyRepo(d),
		TaxonomyRelations: taxonomy.NewTaxonomyRelationRepo(d),
		Topic:             topic.NewTopicRepo(d),
	}
}
