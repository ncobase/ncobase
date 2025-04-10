package schema

import (
	"strings"

	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"github.com/ncobase/ncore/pkg/types"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Process struct {
	ent.Schema
}

func (Process) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "process"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Process) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TextStatus,
		mixin.ProcessRefMixin{},
		mixin.FormBaseMixin{},
		mixin.BusinessTagMixin{},
		mixin.BusinessFlowMixin{},
		mixin.TimeTrackingMixin{},
		mixin.WorkflowControlMixin{},
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Process) Fields() []ent.Field {
	return []ent.Field{
		field.String("process_key").Unique().Comment("Process unique identifier"),
		field.String("initiator").Comment("Process initiator"),
		field.String("initiator_dept").Optional().Comment("Initiator's department"),
		field.String("process_code").Comment("Process code"),
		field.JSON("variables", types.JSON{}).Comment("Process variables"),
		field.String("current_node").Optional().Comment("Current node"),
		field.JSON("active_nodes", types.StringArray{}).Optional().Comment("Currently active nodes"),

		// Process snapshots
		field.JSON("process_snapshot", types.JSON{}).Optional().Comment("Process snapshot"),
		field.JSON("form_snapshot", types.JSON{}).Optional().Comment("Form snapshot"),

		// Urge tracking
		field.Int("urge_count").Default(0).Comment("Number of urges"),
	}
}

func (Process) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("process_key").Unique(),
		index.Fields("business_key"),
		index.Fields("module_code", "form_code"),
		index.Fields("initiator"),
	}
}
