package schema

import (
	"strings"

	"entgo.io/ent/schema/field"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// TenantGroup holds the schema definition for the TenantGroup entity.
type TenantGroup struct {
	ent.Schema
}

// Annotations of the TenantGroup.
func (TenantGroup) Annotations() []schema.Annotation {
	// Keep the original table name to avoid migration issues
	table := strings.Join([]string{"ncse", "org", "tenant_group"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantGroup.
func (TenantGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.GroupID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantGroup.
func (TenantGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("relation_type").
			Default("member").
			Comment("Type of relationship between tenant and group").
			NotEmpty(),
	}
}

// Edges of the TenantGroup.
func (TenantGroup) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantGroup.
func (TenantGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "group_id").Unique(),
		index.Fields("tenant_id", "relation_type"),
		index.Fields("group_id", "relation_type"),
	}
}
