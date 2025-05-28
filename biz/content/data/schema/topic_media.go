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

// TopicMedia holds the schema definition for the TopicMedia entity.
type TopicMedia struct {
	ent.Schema
}

// Annotations of the TopicMedia.
func (TopicMedia) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "topic_media"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the TopicMedia.
func (TopicMedia) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Type,
		mixin.Order,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TopicMedia.
func (TopicMedia) Fields() []ent.Field {
	return []ent.Field{
		field.String("topic_id").
			NotEmpty().
			Comment("Topic ID"),
		field.String("media_id").
			NotEmpty().
			Comment("Media ID"),
	}
}

// Edges of the TopicMedia.
func (TopicMedia) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("media", Media.Type).
			Field("media_id").
			Unique().
			Required(),
		edge.To("topic", Topic.Type).
			Field("topic_id").
			Unique().
			Required(),
	}
}

// Indexes of the TopicMedia.
func (TopicMedia) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("topic_id", "media_id").Unique(),
	}
}
