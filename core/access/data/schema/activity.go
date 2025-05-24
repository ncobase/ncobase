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

// Activity schema for activity
type Activity struct {
	ent.Schema
}

// Annotations of the Activity
func (Activity) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "activity"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Activity
func (Activity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Type,
		mixin.UserID,
		mixin.Details,
		mixin.Metadata,
		mixin.TimeAt{},
	}
}

// Fields of the Activity
func (Activity) Fields() []ent.Field {
	return []ent.Field{}
}

// Indexes of the Activity
func (Activity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "type"),
		index.Fields("user_id", "created_at"),
	}
}
