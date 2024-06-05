package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// TaxonomyRelations holds the schema definition for the TaxonomyRelations entity.
type TaxonomyRelations struct {
	ent.Schema
}

// Annotations of the TaxonomyRelations.
func (TaxonomyRelations) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "taxonomy", "relations"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the TaxonomyRelations.
func (TaxonomyRelations) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("object", "object_id"),
		mixin.TaxonomyID,
		mixin.Type{}, // type, topic, comment, other, ...
		mixin.Order{},
		mixin.CreatedBy{},
		mixin.CreatedAt{},
	}
}

// Fields of the TaxonomyRelations.
func (TaxonomyRelations) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the TaxonomyRelations.
func (TaxonomyRelations) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the TaxonomyRelations
func (TaxonomyRelations) Indexes() []ent.Index {
	return []ent.Index{}
}
