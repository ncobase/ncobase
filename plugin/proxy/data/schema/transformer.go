package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/entgo/mixin"
)

// Transformer holds the schema definition for the Transformer entity.
type Transformer struct {
	ent.Schema
}

// Annotations of the Transformer.
func (Transformer) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "tbp", "transformer"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Transformer.
func (Transformer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.Disabled,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Transformer.
func (Transformer) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			Comment("Transformer type (template, script, mapping)").
			NotEmpty(),
		field.Text("content").
			Comment("Transformer content (template, script or mapping definition)").
			NotEmpty(),
		field.String("content_type").
			Comment("Content type (text/javascript, application/json, text/template)").
			Default("application/json"),
	}
}

// Edges of the Transformer.
func (Transformer) Edges() []ent.Edge {
	return nil
}

// Indexes of the Transformer.
func (Transformer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}
