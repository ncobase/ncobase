package schema

import (
	"ncobase/common/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Subscription holds the schema definition for the Subscription entity.
type Subscription struct {
	ent.Schema
}

// Annotations of the Subscription.
func (Subscription) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "rt", "subscription"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Subscription.
func (Subscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,    // user id
		mixin.ChannelID, // notification channel
		mixin.Status,    // 0: disabled, 1: enabled
		mixin.TimeAt{},
	}
}

// Fields of the Subscription.
func (Subscription) Fields() []ent.Field {
	return nil
}

// Edges of the Subscription.
func (Subscription) Edges() []ent.Edge {
	return nil
}

// Indexes of the Subscription.
func (Subscription) Indexes() []ent.Index {
	return []ent.Index{}
}
