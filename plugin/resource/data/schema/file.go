package schema

import (
	"strings"

	"entgo.io/ent/schema/field"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
	"github.com/ncobase/ncore/types"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// File schema definition
type File struct {
	ent.Schema
}

// Annotations for File
func (File) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "res", "file"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin for File
func (File) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.NameUnique,
		mixin.Path,
		mixin.Type,
		mixin.Size,
		mixin.Storage,
		mixin.Bucket,
		mixin.Endpoint,
		mixin.OwnerID,
		mixin.SpaceID,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields for File
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("folder_path").
			Optional().
			Comment("Virtual folder path"),

		field.String("access_level").
			Default("private").
			Comment("Access level: public, private, shared"),

		field.Int64("expires_at").
			Optional().
			Nillable().
			Comment("Expiration timestamp"),

		field.Strings("tags").
			Optional().
			Comment("File tags"),

		field.Bool("is_public").
			Default(false).
			Comment("Public access flag"),

		field.String("category").
			Default("other").
			Comment("File category"),

		field.JSON("processing_result", types.JSON{}).
			Optional().
			Comment("Processing operation results"),
	}
}

// Edges for File
func (File) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes for File
func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "owner_id"),
		index.Fields("folder_path"),
		index.Fields("category"),
		index.Fields("is_public"),
		index.Fields("access_level"),
	}
}
