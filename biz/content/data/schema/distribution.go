package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Distribution holds the schema definition for the Distribution entity.
type Distribution struct {
	ent.Schema
}

// Annotations of the Distribution.
func (Distribution) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "distribution"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Distribution.
func (Distribution) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Distribution.
func (Distribution) Fields() []ent.Field {
	return []ent.Field{
		field.String("topic_id").
			NotEmpty().
			Comment("Topic ID"),
		field.String("channel_id").
			NotEmpty().
			Comment("Channel ID"),
		field.Int("status").
			Default(0).
			Comment("Distribution status: 0:draft, 1:scheduled, 2:published, 3:failed, 4:cancelled"),
		field.Int64("scheduled_at").
			Optional().
			Nillable().
			Comment("Scheduled publish time"),
		field.Int64("published_at").
			Optional().
			Nillable().
			Comment("Actual publish time"),
		field.String("external_id").
			Optional().
			Comment("External ID on the platform"),
		field.String("external_url").
			Optional().
			Comment("URL on the external platform"),
		field.String("error_details").
			Optional().
			Comment("Error details if distribution failed"),
	}
}

// Edges of the Distribution.
func (Distribution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("topic", Topic.Type).
			Field("topic_id").
			Unique().
			Required(),
		edge.To("channel", CMSChannel.Type).
			Field("channel_id").
			Unique().
			Required(),
	}
}

// Indexes of the Distribution.
func (Distribution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("topic_id", "channel_id").Unique(),
	}
}
