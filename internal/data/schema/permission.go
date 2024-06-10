package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Permission holds the schema definition for the Permission entity.
type Permission struct {
	ent.Schema
}

// Annotations of the Permission.
func (Permission) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "permission"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Permission.
func (Permission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey{},
		mixin.Name{},
		mixin.Action{},
		mixin.Subject{},
		mixin.Description{},
		mixin.Default{},
		mixin.Disabled{},
		mixin.ExtraProps{},
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Permission.
func (Permission) Fields() []ent.Field {
	return nil
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return nil
}

// Indexes of the Permission.
func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action", "subject"),
	}
}
