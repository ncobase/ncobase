package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"github.com/ncobase/ncore/pkg/types"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type History struct {
	ent.Schema
}

func (History) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "history"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate()),
		entsql.WithComments(true),
	}
}

func (History) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Type,
		mixin.ProcessRefMixin{},
		mixin.NodeBaseMixin{},
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (History) Fields() []ent.Field {
	return []ent.Field{
		field.String("node_name").Comment("Node name"),
		field.String("operator").Comment("Operation user"),
		field.String("operator_dept").Optional().Comment("Operator's department"),
		field.String("task_id").Optional().Comment("Task ID"),
		field.JSON("variables", types.JSON{}).Comment("Task variables"),
		field.JSON("form_data", types.JSON{}).Optional().Comment("Form data"),
		field.String("action").Comment("Operation action"),
		field.String("comment").Optional().Comment("Operation comment"),
		field.JSON("details", types.JSON{}).Optional().Comment("Detailed information"),
	}
}

func (History) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("process_id"),
		index.Fields("operator"),
		index.Fields("action"),
	}
}
