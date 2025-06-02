package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// TenantOption holds the schema definition for the TenantOption entity.
type TenantOption struct {
	ent.Schema
}

// Annotations of the TenantOption.
func (TenantOption) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "tenant_option"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantOption.
func (TenantOption) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.OptionID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantOption.
func (TenantOption) Fields() []ent.Field {
	return nil
}

// Edges of the TenantOption.
func (TenantOption) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantOption.
func (TenantOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "option_id"),
	}
}
