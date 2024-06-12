package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Asset holds the schema definition for the Asset entity.
type Asset struct {
	ent.Schema
}

// Annotations of the Asset.
func (Asset) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "asset"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Asset.
func (Asset) Mixin() []ent.Mixin {
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

// Fields of the Asset.
func (Asset) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Asset.
func (Asset) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Asset.
func (Asset) Indexes() []ent.Index {
	return []ent.Index{}
}
