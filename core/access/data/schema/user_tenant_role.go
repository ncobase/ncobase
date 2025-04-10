package schema

import (
	"strings"

	"github.com/ncobase/ncore/pkg/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserTenantRole holds the schema definition for the UserTenantRole entity.
type UserTenantRole struct {
	ent.Schema
}

// Annotations of the UserTenantRole.
func (UserTenantRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "user_tenant_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserTenantRole.
func (UserTenantRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.TenantID,
		mixin.RoleID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserTenantRole.
func (UserTenantRole) Fields() []ent.Field {
	return nil
}

// Edges of the UserTenantRole.
func (UserTenantRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserTenantRole.
func (UserTenantRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "tenant_id", "role_id"),
	}
}
