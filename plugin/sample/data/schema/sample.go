package schema

import (
	"ncobase/common/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"

	"entgo.io/ent"
)

// Sample holds the schema definition for the Sample entity.
type Sample struct {
	ent.Schema
}

// Annotations of the Sample.
func (Sample) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sample"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Sample.
func (Sample) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Sample.
func (Sample) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Sample.
func (Sample) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Sample.
func (Sample) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
