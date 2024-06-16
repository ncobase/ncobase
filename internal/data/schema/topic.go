package schema

import (
	"ncobase/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"

	"entgo.io/ent"
)

// Topic holds the schema definition for the Topic entity.
type Topic struct {
	ent.Schema
}

// Annotations of the Topic.
func (Topic) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "topic"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Topic.
func (Topic) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Title,
		mixin.SlugUnique,
		mixin.Content,
		mixin.Thumbnail,
		mixin.Temp,
		mixin.Markdown,
		mixin.Private,
		mixin.Status, // status, 0: draft, 1: published, 2: trashed, 3: temp, ...
		mixin.Released,
		mixin.TaxonomyID,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Topic.
func (Topic) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Topic.
func (Topic) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Topic.
func (Topic) Indexes() []ent.Index {
	return []ent.Index{}
}
