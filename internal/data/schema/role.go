package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Annotations of the Role.
func (Role) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Role.
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey{},
		mixin.Name{},
		mixin.SlugUnique{},
		mixin.Disabled{},
		mixin.Description{},
		mixin.ExtraProps{},
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return nil
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return nil
}

// Indexes of the Role.
func (Role) Indexes() []ent.Index {
	return []ent.Index{}
}
