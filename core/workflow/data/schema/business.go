package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Business holds the schema for business data
type Business struct {
	ent.Schema
}

func (Business) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "flow", "business"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (Business) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Code,       // Business code
		mixin.TextStatus, // Business status
		mixin.FormBaseMixin{},
		mixin.ProcessRefMixin{},
		mixin.DataTrackingMixin{},
		mixin.BusinessFlowMixin{},
		mixin.BusinessTagMixin{},
		mixin.PermissionMixin{},
		mixin.ExtraProps,
		mixin.TenantID,
		mixin.Operator,
		mixin.TimeAt{},
	}
}

func (Business) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("business_key").Unique(),
		index.Fields("process_id"),
		index.Fields("module_code", "form_code"),
		index.Fields("flow_status"),
	}
}
