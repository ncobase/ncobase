package schema

import (
	"ncobase/common/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Notification holds the schema definition for the Notification entity.
type Notification struct {
	ent.Schema
}

// Annotations of the Notification.
func (Notification) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "rt", "notification"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Notification.
func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Title,     // notification title
		mixin.Content,   // notification content
		mixin.Type,      // info, warning, error
		mixin.UserID,    // user id
		mixin.Status,    // 0: unread, 1: read, 2: deleted
		mixin.Links,     // related links
		mixin.ChannelID, // notification channel
		mixin.TimeAt{},
	}
}

// Fields of the Notification.
func (Notification) Fields() []ent.Field {
	return nil
}

// Edges of the Notification.
func (Notification) Edges() []ent.Edge {
	return nil
}

// Indexes of the Notification.
func (Notification) Indexes() []ent.Index {
	return []ent.Index{}
}
