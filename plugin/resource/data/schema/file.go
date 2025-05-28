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

// File holds the schema definition for the File entity.
type File struct {
	ent.Schema
}

// Annotations of the File.
func (File) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "res", "file"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the File.
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
		mixin.ObjectID,
		mixin.TenantID,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the File.
func (File) Fields() []ent.Field {
	return []ent.Field{
		field.String("folder_path").
			Optional().
			Comment("Virtual folder path for organization"),

		field.String("access_level").
			Default("private").
			Comment("Access level: public, private, shared"),

		field.Int64("expires_at").
			Optional().
			Nillable().
			Comment("Expiration timestamp"),

		field.JSON("metadata", map[string]any{}).
			Optional().
			Comment("Extended file metadata"),

		field.Strings("tags").
			Optional().
			Comment("Tags for categorization"),

		field.Bool("is_public").
			Default(false).
			Comment("Publicly accessible flag"),

		field.JSON("versions", []string{}).
			Optional().
			Comment("IDs of previous versions"),

		field.String("thumbnail_path").
			Optional().
			Comment("Path to thumbnail if available"),

		field.Int("width").
			Optional().
			Nillable().
			Comment("Width for image files"),

		field.Int("height").
			Optional().
			Nillable().
			Comment("Height for image files"),

		field.Float("duration").
			Optional().
			Nillable().
			Comment("Duration for audio/video files in seconds"),

		field.String("category").
			Default("other").
			Comment("File category: image, document, video, audio, archive, other"),

		field.JSON("processing_result", types.JSON{}).
			Optional().
			Comment("Results from any processing operations"),
	}
}

// Edges of the File.
func (File) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the File.
func (File) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("folder_path"),
		index.Fields("is_public"),
		index.Fields("category"),
		index.Fields("tags"),
	}
}
