package schema

import (
	"ncobase/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// TaxonomyRelation holds the schema definition for the TaxonomyRelation entity.
type TaxonomyRelation struct {
	ent.Schema
}

// Annotations of the TaxonomyRelation.
func (TaxonomyRelation) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "taxonomy_relation"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the TaxonomyRelation.
func (TaxonomyRelation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("object", "object_id"),
		mixin.TaxonomyID,
		mixin.Type, // type, topic, comment, other, ...
		mixin.Order,
		mixin.CreatedBy,
		mixin.CreatedAt,
	}
}

// Fields of the TaxonomyRelation.
func (TaxonomyRelation) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the TaxonomyRelation.
func (TaxonomyRelation) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the TaxonomyRelation
func (TaxonomyRelation) Indexes() []ent.Index {
	return []ent.Index{}
}
