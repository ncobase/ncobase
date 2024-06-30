package schema

import (
	"ncobase/common/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"

	"entgo.io/ent"
)

// Module holds the schema definition for the Module entity.
type Module struct {
	ent.Schema
}

// Annotations of the Module.
func (Module) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "module"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Module.
func (Module) Mixin() []ent.Mixin {
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
		mixin.Status,
		mixin.Released,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Module.
func (Module) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Module.
func (Module) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Module.
func (Module) Indexes() []ent.Index {
	return []ent.Index{}
}
