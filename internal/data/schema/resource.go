package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Resource holds the schema definition for the Resource entity.
type Resource struct {
	ent.Schema
}

// Annotations of the Resource.
func (Resource) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "resource"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Resource.
func (Resource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey{},
		mixin.NameUnique{},
		mixin.Path{},
		mixin.Type{},
		mixin.Size{},
		mixin.Storage{},
		mixin.URL{},
		mixin.ObjectID,
		mixin.DomainID,
		mixin.ExtraProps{},
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Resource.
func (Resource) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Resource.
func (Resource) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Resource.
func (Resource) Indexes() []ent.Index {
	return []ent.Index{}
}
