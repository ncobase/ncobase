package schema

import (
	"ncobase/common/data/entgo/mixin"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserGroup holds the schema definition for the UserGroup entity.
type UserGroup struct {
	ent.Schema
}

// Annotations of the UserGroup.
func (UserGroup) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "org", "user_group"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserGroup.
func (UserGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.GroupID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserGroup.
func (UserGroup) Fields() []ent.Field {
	return nil
}

// Edges of the UserGroup.
func (UserGroup) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserGroup.
func (UserGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "group_id"),
	}
}
