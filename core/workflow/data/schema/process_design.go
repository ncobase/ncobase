package schema

import (
	"ncobase/common/data/entgo/mixin"
	"ncobase/common/types"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProcessDesign represents process design data
type ProcessDesign struct {
	ent.Schema
}

func (ProcessDesign) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "process_design"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (ProcessDesign) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Version,
		mixin.Disabled,
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (ProcessDesign) Fields() []ent.Field {
	return []ent.Field{
		field.String("template_id").Comment("Template ID"),
		field.JSON("graph_data", types.JSON{}).Optional().Comment("Process graph data"),
		field.JSON("node_layouts", types.JSON{}).Optional().Comment("Node layout positions"),
		field.JSON("properties", types.JSON{}).Optional().Comment("Process design properties"),
		field.JSON("validation_rules", types.JSON{}).Optional().Comment("Process validation rules"),
		field.Bool("is_draft").Default(false).Comment("Whether is draft"),
		field.String("source_version").Optional().Comment("Source version"),
	}
}

func (ProcessDesign) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("template_id"),
	}
}
