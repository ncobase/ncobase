package schema

import (
	"ncobase/common/data/entgo/mixin"
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
	table := strings.Join([]string{"ncse", "iam", "user_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the UserRole.
func (UserRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.RoleID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
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
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "role_id"),
	}
}
