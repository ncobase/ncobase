package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

// Annotations of the Event.
func (Event) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "rt", "event"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Event.
func (Event) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Type,      // event type, e.g. created, updated, deleted, connected, disconnected
		mixin.ChannelID, // notification channel
		mixin.Payload,   // payload data
		mixin.UserID,    // user id
		mixin.CreatedAt,
	}
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return nil
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return nil
}

// Indexes of the Event.
func (Event) Indexes() []ent.Index {
	return []ent.Index{}
}
