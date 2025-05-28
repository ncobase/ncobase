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

// CMSChannel holds the schema definition for the CMSChannel entity.
type CMSChannel struct {
	ent.Schema
}

// Annotations of the CMSChannel.
func (CMSChannel) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "channel"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the CMSChannel.
func (CMSChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Type,
		mixin.SlugUnique,
		mixin.Icon,
		mixin.Status,
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the CMSChannel.
func (CMSChannel) Fields() []ent.Field {
	return []ent.Field{
		field.Strings("allowed_types").
			Optional().
			Comment("Allowed content types for this channel"),
		field.String("description").
			Optional().
			Comment("Channel description"),
		field.String("logo").
			Optional().
			Comment("Channel logo URL"),
		field.String("webhook_url").
			Optional().
			Comment("Webhook URL for notifications"),
		field.Bool("auto_publish").
			Default(false).
			Comment("Auto publish content to this channel"),
		field.Bool("require_review").
			Default(false).
			Comment("Require review before publishing"),
	}
}

// Edges of the CMSChannel.
func (CMSChannel) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the CMSChannel.
func (CMSChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
