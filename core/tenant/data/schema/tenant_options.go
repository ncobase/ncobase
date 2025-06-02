package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// TenantOptions holds the schema definition for the TenantOptions entity.
type TenantOptions struct {
	ent.Schema
}

// Annotations of the TenantOptions.
func (TenantOptions) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "tenant_options"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantOptions.
func (TenantOptions) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.OptionsID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantOptions.
func (TenantOptions) Fields() []ent.Field {
	return nil
}

// Edges of the TenantOptions.
func (TenantOptions) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantOptions.
func (TenantOptions) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "options_id"),
	}
}
