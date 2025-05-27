package schema

import (
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
)

// TenantQuota holds the schema definition for the TenantQuota entity
type TenantQuota struct {
	ent.Schema
}

// Annotations of the TenantQuota
func (TenantQuota) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "tenant_quota"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantQuota
func (TenantQuota) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantQuota
func (TenantQuota) Fields() []ent.Field {
	return []ent.Field{
		field.String("quota_type").
			NotEmpty().
			Comment("Type of quota (users, storage, api_calls, etc.)"),
		field.String("quota_name").
			NotEmpty().
			Comment("Human readable name of the quota"),
		field.Int64("max_value").
			NonNegative().
			Comment("Maximum allowed value for this quota"),
		field.Int64("current_used").
			Default(0).
			NonNegative().
			Comment("Current usage of this quota"),
		field.String("unit").
			Default("count").
			Comment("Unit of measurement (count, bytes, mb, gb, tb)"),
		field.Bool("enabled").
			Default(true).
			Comment("Whether this quota is actively enforced"),
	}
}

// Edges of the TenantQuota
func (TenantQuota) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantQuota
func (TenantQuota) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "quota_type").Unique(),
		index.Fields("tenant_id", "enabled"),
	}
}
