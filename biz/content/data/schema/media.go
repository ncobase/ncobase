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

// Media schema definition
type Media struct {
	ent.Schema
}

// Annotations for Media
func (Media) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "media"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin for Media
func (Media) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Title,
		mixin.Type,
		mixin.URL,
		mixin.ExtraProps,
		mixin.SpaceID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields for Media
func (Media) Fields() []ent.Field {
	return []ent.Field{
		field.String("owner_id").
			NotEmpty().
			Comment("Media owner identifier"),

		field.String("resource_id").
			Optional().
			Comment("Reference to resource plugin file ID"),

		field.String("description").
			Optional().
			Comment("Media description"),

		field.String("alt").
			Optional().
			Comment("Alternative text for accessibility"),
	}
}

// Edges for Media
func (Media) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes for Media
func (Media) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "owner_id"),
		index.Fields("resource_id"),
		index.Fields("type"),
	}
}
