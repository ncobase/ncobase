package schema

import (
	"ncore/pkg/data/entgo/mixin"
	"ncore/pkg/types"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Node struct {
	ent.Schema
}

func (Node) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "node"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Node) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.Type,
		mixin.TextStatus,
		mixin.NodeBaseMixin{},
		mixin.FormBaseMixin{},
		mixin.TaskAssigneeMixin{},
		mixin.WorkflowControlMixin{},
		mixin.TimeTrackingMixin{},
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Node) Fields() []ent.Field {
	return []ent.Field{
		field.String("process_id").Comment("Process ID"),
		field.JSON("permissions", types.JSON{}).Comment("Permission configs"),
		// Node relationships
		field.JSON("prev_nodes", types.StringArray{}).Optional().Comment("Previous nodes"),
		field.JSON("next_nodes", types.StringArray{}).Optional().Comment("Next nodes"),
		field.JSON("parallel_nodes", types.StringArray{}).Optional().Comment("Parallel nodes"),
		field.JSON("branch_nodes", types.StringArray{}).Optional().Comment("Branch nodes"),

		// Node specific configs
		field.JSON("conditions", types.StringArray{}).Optional().Comment("Transition conditions"),
		field.JSON("properties", types.JSON{}).Optional().Comment("Node properties"),
		field.Bool("is_countersign").Default(false).Comment("Whether requires countersign"),
		field.String("countersign_rule").Optional().Comment("Countersign rules"),

		// Handlers
		field.JSON("handlers", types.JSON{}).Optional().Comment("Handler configurations"),
		field.JSON("listeners", types.JSON{}).Optional().Comment("Listener configurations"),
		field.JSON("hooks", types.JSON{}).Optional().Comment("Hook configurations"),

		field.JSON("variables", types.JSON{}).Optional().Comment("Node variables"),

		// Retry settings
		field.Int("retry_times").Optional().Default(0).Comment("Number of retries"),
		field.Int("retry_interval").Optional().Default(0).Comment("Retry interval in seconds"),
		field.Bool("is_working_day").Default(true).Comment("Whether to count working days only"),
	}
}

func (Node) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("node_key").Unique(),
		index.Fields("type"),
	}
}
