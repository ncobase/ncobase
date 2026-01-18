package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserSpace holds the schema definition for the UserSpace entity.
type UserSpace struct {
	ent.Schema
}

// Annotations of the UserSpace.
func (UserSpace) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "space", "user"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserSpace.
func (UserSpace) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.SpaceID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserSpace.
func (UserSpace) Fields() []ent.Field {
	return nil
}

// Edges of the UserSpace.
func (UserSpace) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserSpace.
func (UserSpace) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "space_id"),
	}
}
