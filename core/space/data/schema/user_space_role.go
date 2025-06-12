package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserSpaceRole holds the schema definition for the UserSpaceRole entity.
type UserSpaceRole struct {
	ent.Schema
}

// Annotations of the UserSpaceRole.
func (UserSpaceRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "user_space_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserSpaceRole.
func (UserSpaceRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.SpaceID,
		mixin.RoleID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserSpaceRole.
func (UserSpaceRole) Fields() []ent.Field {
	return nil
}

// Edges of the UserSpaceRole.
func (UserSpaceRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserSpaceRole.
func (UserSpaceRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "space_id", "role_id"),
	}
}
