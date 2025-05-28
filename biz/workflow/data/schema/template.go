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

type Template struct {
	ent.Schema
}

func (Template) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "template"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Template) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Code,
		mixin.Description,
		mixin.Type,
		mixin.Version,
		mixin.TextStatus,
		mixin.Disabled,
		mixin.FormBaseMixin{},
		mixin.NodeBaseMixin{},
		mixin.BusinessTagMixin{},
		mixin.WorkflowControlMixin{},
		mixin.PermissionMixin{},
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Template) Fields() []ent.Field {
	return []ent.Field{
		field.String("template_key").Unique().Comment("Template unique identifier"),

		// Configurations
		field.JSON("process_rules", types.JSON{}).Optional().Comment("Process rules"),
		field.JSON("trigger_conditions", types.JSON{}).Optional().Comment("Trigger conditions"),
		field.JSON("timeout_config", types.JSON{}).Optional().Comment("Timeout configuration"),
		field.JSON("reminder_config", types.JSON{}).Optional().Comment("Reminder configuration"),

		// Version control
		field.String("source_version").Optional().Comment("Source version"),
		field.Bool("is_latest").Default(false).Comment("Whether is latest version"),
		field.Int64("effective_time").Optional().Comment("Effective time"),
		field.Int64("expire_time").Optional().Comment("Expire time"),
	}
}

func (Template) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("template_key").Unique(),
		index.Fields("code").Unique(),
		index.Fields("module_code", "form_code"),
	}
}
