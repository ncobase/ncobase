package schema

import (
	"ncobase/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// RolePermission holds the schema definition for the RolePermission entity.
type RolePermission struct {
	ent.Schema
}

// Annotations of the RolePermission.
func (RolePermission) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "role_permission"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the RolePermission.
func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("role", "role_id"),
		mixin.PermissionID,
	}
}

// Fields of the RolePermission.
func (RolePermission) Fields() []ent.Field {
	return nil
}

// Edges of the RolePermission.
func (RolePermission) Edges() []ent.Edge {
	return nil
}

// Indexes of the RolePermission.
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "permission_id"),
	}
}
