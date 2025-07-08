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
		mixin.PrimaryKey,   // id field
		mixin.NameUnique,   // name field (unique per owner)
		mixin.Path,         // path field (storage path)
		mixin.Type,         // type field (content type)
		mixin.Size,         // size field (file size in bytes)
		mixin.Storage,      // storage field (storage provider)
		mixin.Bucket,       // bucket field (storage bucket)
		mixin.Endpoint,     // endpoint field (storage endpoint)
		mixin.OwnerID,      // owner_id field (file owner)
		mixin.ExtraProps,   // extras field (additional properties)
		mixin.OperatorBy{}, // created_by, updated_by fields
		mixin.TimeAt{},     // created_at, updated_at fields
	}
}

// Fields for File - All additional fields not covered by mixins
func (File) Fields() []ent.Field {
	return []ent.Field{
		// Original filename before processing
		field.String("original_name").
			Optional().
			Comment("Original filename before processing"),

		// Access control
		field.String("access_level").
			Default("private").
			Comment("Access level: public, private, shared"),

		// Expiration
		field.Int64("expires_at").
			Optional().
			Nillable().
			Comment("Expiration timestamp"),

		// Tags for categorization
		field.Strings("tags").
			Optional().
			Comment("File tags for categorization"),

		// Public access flag
		field.Bool("is_public").
			Default(false).
			Comment("Public access flag"),

		// File category
		field.String("category").
			Default("other").
			Comment("File category (image, document, video, etc.)"),

		// File content hash for deduplication
		field.String("hash").
			Optional().
			Comment("File content hash for deduplication"),

		// Processing results
		field.JSON("processing_result", types.JSON{}).
			Optional().
			Comment("Processing operation results"),
	}
}

// Edges for File
func (File) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes for File with comprehensive coverage
func (File) Indexes() []ent.Index {
	return []ent.Index{
		// Primary uniqueness constraints
		index.Fields("owner_id", "name").Unique(),
		index.Fields("id", "created_at").Unique(),

		// Performance indexes for common queries
		index.Fields("owner_id", "created_at"),
		index.Fields("category"),
		index.Fields("is_public"),
		index.Fields("access_level"),
		index.Fields("storage"),
		index.Fields("type"),
		index.Fields("hash"), // For deduplication queries

		// Composite indexes for filtered queries
		index.Fields("owner_id", "category"),
		index.Fields("owner_id", "is_public"),
		index.Fields("owner_id", "access_level"),
		index.Fields("owner_id", "type"),
		index.Fields("created_by", "created_at"),

		// Search and filtering indexes
		index.Fields("name"),
		index.Fields("original_name"),
		index.Fields("path"),

		// Expiration and cleanup indexes
		index.Fields("expires_at"),
		index.Fields("created_at", "expires_at"),

		// Size-based queries
		index.Fields("owner_id", "size"),
		index.Fields("size"),
	}
}
