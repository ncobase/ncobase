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

type Task struct {
	ent.Schema
}

func (Task) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "task"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Task) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.TextStatus,
		mixin.ProcessRefMixin{},
		mixin.NodeBaseMixin{},
		mixin.TaskAssigneeMixin{},
		mixin.TimeTrackingMixin{},
		mixin.WorkflowControlMixin{},
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.String("task_key").Unique().Comment("Task unique identifier"),
		field.String("parent_id").Optional().Comment("Parent task ID"),
		field.JSON("child_ids", []string{}).Default([]string{}).Comment("Child task IDs"),
		// Task processing
		field.String("action").Optional().Comment("Processing action"),
		field.String("comment").Optional().Comment("Processing comment"),
		field.JSON("attachments", types.JSON{}).Optional().Comment("Attachment information"),
		field.JSON("form_data", types.JSON{}).Optional().Comment("Form data"),
		field.JSON("variables", types.JSON{}).Optional().Comment("Task variables"),
		field.Bool("is_resubmit").Default(false).Comment("Whether is resubmitted"),

		// Task tracking
		field.Int64("claim_time").Optional().Nillable().Comment("Claim time"),
		field.Bool("is_urged").Default(false).Comment("Whether is urged"),
		field.Int("urge_count").Default(0).Comment("Number of urges"),
	}
}

func (Task) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("task_key").Unique(),
		index.Fields("process_id", "node_key"),
		index.Fields("node_type"),
		index.Fields("due_time"),
	}
}
