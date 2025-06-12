package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"
	"github.com/ncobase/ncore/types"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Delegation represents workflow delegation rules
type Delegation struct {
	ent.Schema
}

func (Delegation) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "delegation"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Delegation) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TextStatus,
		mixin.ExtraProps,
		mixin.SpaceID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Delegation) Fields() []ent.Field {
	return []ent.Field{
		field.String("delegator_id").Comment("User ID who delegates"),
		field.String("delegatee_id").Comment("User ID to delegate to"),
		field.String("template_id").Optional().Comment("Template ID if specific"),
		field.String("node_type").Optional().Comment("Node type if specific"),
		field.JSON("conditions", types.StringArray{}).Optional().Comment("Delegation conditions"),
		field.Int64("start_time").Comment("Delegation start time"),
		field.Int64("end_time").Comment("Delegation end time"),
		field.Bool("is_enabled").Default(true).Comment("Whether delegation is enabled"),
	}
}

func (Delegation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("delegator_id"),
		index.Fields("template_id", "node_type"),
	}
}
