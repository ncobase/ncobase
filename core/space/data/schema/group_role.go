package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// GroupRole holds the schema definition for the GroupRole entity.
type GroupRole struct {
	ent.Schema
}

// Annotations of the GroupRole.
func (GroupRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "group_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the GroupRole.
func (GroupRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.GroupID,
		mixin.RoleID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the GroupRole.
func (GroupRole) Fields() []ent.Field {
	return nil
}

// Edges of the GroupRole.
func (GroupRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the GroupRole.
func (GroupRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("group_id", "role_id"),
	}
}
