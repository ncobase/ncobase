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

// Rule represents workflow rule configurations
type Rule struct {
	ent.Schema
}

func (Rule) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "rule"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Rule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Code,
		mixin.Description,
		mixin.Type,
		mixin.TextStatus,
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Rule) Fields() []ent.Field {
	return []ent.Field{
		field.String("rule_key").Unique().Comment("Rule unique key"),
		field.String("template_id").Optional().Comment("Template ID if template specific"),
		field.String("node_key").Optional().Comment("Node key if node specific"),
		field.JSON("conditions", types.StringArray{}).Comment("Rule conditions"),
		field.JSON("actions", types.JSON{}).Comment("Rule actions"),
		field.Int("priority").Default(0).Comment("Rule priority"),
		field.Bool("is_enabled").Default(true).Comment("Whether rule is enabled"),
		field.Int64("effective_time").Optional().Comment("Rule effective time"),
		field.Int64("expire_time").Optional().Comment("Rule expire time"),
	}
}

func (Rule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("rule_key").Unique(),
		index.Fields("template_id", "node_key"),
	}
}
