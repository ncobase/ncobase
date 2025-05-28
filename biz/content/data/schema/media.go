package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Media holds the schema definition for the Media entity.
type Media struct {
	ent.Schema
}

// Annotations of the Media.
func (Media) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "media"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Media.
func (Media) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Title,
		mixin.Type,
		mixin.URL,
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Media.
func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("path").
			Optional().
			Comment("File path"),
		field.String("mime_type").
			Optional().
			Comment("MIME type"),
		field.Int64("size").
			Default(0).
			Comment("File size in bytes"),
		field.Int("width").
			Default(0).
			Comment("Image/video width"),
		field.Int("height").
			Default(0).
			Comment("Image/video height"),
		field.Float("duration").
			Default(0).
			Comment("Audio/video duration in seconds"),
		field.String("description").
			Optional().
			Comment("Media description"),
		field.String("alt").
			Optional().
			Comment("Alternative text for accessibility"),
	}
}

// Edges of the Media.
func (Media) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Media.
func (Media) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
