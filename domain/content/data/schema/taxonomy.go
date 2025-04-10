package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"

	"entgo.io/ent"
)

// Taxonomy holds the schema definition for the Taxonomy entity.
type Taxonomy struct {
	ent.Schema
}

// Annotations of the Taxonomy.
func (Taxonomy) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "cms", "taxonomy"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Taxonomy.
func (Taxonomy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Type, // type, default 'node', options: 'node', 'plane', 'event', 'page', 'tag', 'link'
		mixin.SlugUnique,
		mixin.Cover,
		mixin.Thumbnail,
		mixin.Color,
		mixin.Icon,
		mixin.URL,
		mixin.Keywords,
		mixin.Description,
		mixin.Status, // status, 0: enabled, 1: disabled, ...
		mixin.ExtraProps,
		mixin.ParentID,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Taxonomy.
func (Taxonomy) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Taxonomy.
func (Taxonomy) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Taxonomy.
func (Taxonomy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
