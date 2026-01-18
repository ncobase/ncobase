package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		mixin.Type,    // event type
		mixin.Payload, // event payload data
		mixin.CreatedAt,
	}
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.String("source").
			Comment("Event source identifier").
			Optional(),

		field.String("status").
			Comment("Processing status: pending, processed, failed, retry").
			Default("pending"),

		field.String("priority").
			Comment("Event priority: low, normal, high, critical").
			Default("normal").
			Optional(),

		field.Int64("processed_at").
			Comment("Event processing timestamp").
			Optional(),

		field.Int("retry_count").
			Comment("Number of retry attempts").
			Default(0),

		field.Text("error_message").
			Comment("Error message if processing failed").
			Optional(),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return nil
}

// Indexes of the Event.
func (Event) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("source"),
		index.Fields("type"),
		index.Fields("priority"),
		index.Fields("created_at", "status"),
		index.Fields("type", "source"),
		index.Fields("status", "retry_count"),
	}
}
