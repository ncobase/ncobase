package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserTenant holds the schema definition for the UserTenant entity.
type UserTenant struct {
	ent.Schema
}

// Annotations of the UserTenant.
func (UserTenant) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "user_tenant"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserTenant.
func (UserTenant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserTenant.
func (UserTenant) Fields() []ent.Field {
	return nil
}

// Edges of the UserTenant.
func (UserTenant) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserTenant.
func (UserTenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "tenant_id"),
	}
}
