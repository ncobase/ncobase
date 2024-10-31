package schema

import (
	"ncobase/common/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Attachment holds the schema definition for the Attachment entity.
type Attachment struct {
	ent.Schema
}

// Annotations of the Attachment.
func (Attachment) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "attachment"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Attachment.
func (Attachment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.NameUnique,
		mixin.Path,
		mixin.Type,
		mixin.Size,
		mixin.Storage,
		mixin.Bucket,
		mixin.Endpoint,
		mixin.ObjectID,
		mixin.TenantID,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Attachment.
func (Attachment) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Attachment.
func (Attachment) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Attachment.
func (Attachment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
