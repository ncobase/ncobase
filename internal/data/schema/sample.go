package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Sample holds the schema definition for the Sample entity.
type Sample struct {
	ent.Schema
}

// Annotations of the Sample.
func (Sample) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sample"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Sample.
func (Sample) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey{},
		mixin.Name{},
		mixin.Content{},
		mixin.TimeAt{},
	}
}

// Fields of the Sample.
func (Sample) Fields() []ent.Field {
	return nil
}

// Edges of the Sample.
func (Sample) Edges() []ent.Edge {
	return nil
}
