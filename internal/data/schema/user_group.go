package schema

import (
	"ncobase/internal/data/schema/mixin"
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
	table := strings.Join([]string{"sc", "user_group"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserGroup.
func (UserGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.GroupID,
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
		index.Fields("id", "group_id"),
	}
}
