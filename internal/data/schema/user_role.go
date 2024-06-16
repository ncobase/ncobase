package schema

import (
	"ncobase/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserRole holds the schema definition for the UserRole entity.
type UserRole struct {
	ent.Schema
}

// Annotations of the UserRole.
func (UserRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "user_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the UserRole.
func (UserRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.RoleID,
	}
}

// Fields of the UserRole.
func (UserRole) Fields() []ent.Field {
	return nil
}

// Edges of the UserRole.
func (UserRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserRole.
func (UserRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "role_id"),
	}
}
