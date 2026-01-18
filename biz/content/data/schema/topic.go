package schema

import (
	"strings"

	"entgo.io/ent/schema/field"
	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"

	"entgo.io/ent"
)

// Topic holds the schema definition for the Topic entity.
type Topic struct {
	ent.Schema
}

// Annotations of the Topic.
func (Topic) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "topic"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Topic.
func (Topic) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Title,
		mixin.SlugUnique,
		mixin.Content,
		mixin.Thumbnail,
		mixin.Temp,
		mixin.Markdown,
		mixin.Private,
		mixin.Status,
		mixin.Released,
		mixin.TaxonomyID,
		mixin.SpaceID,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Topic.
func (Topic) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").
			Default(1).
			Comment("Content version"),
		field.String("content_type").
			Default("article").
			Comment("Content type: article, video, etc."),
		field.String("seo_title").
			Optional().
			Comment("SEO title"),
		field.String("seo_description").
			Optional().
			Comment("SEO description"),
		field.String("seo_keywords").
			Optional().
			Comment("SEO keywords"),
		field.Bool("excerpt_auto").
			Default(true).
			Comment("Auto generate excerpt"),
		field.String("excerpt").
			Optional().
			Comment("Manual excerpt"),
		field.String("featured_media").
			Optional().
			Comment("Featured media ID"),
		field.Strings("tags").
			Optional().
			Comment("Content tags"),
	}
}

// Edges of the Topic.
func (Topic) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Topic.
func (Topic) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
