package schema

import (
	"ncobase/common/data/entgo/mixin"
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
	table := strings.Join([]string{"ncse", "iam", "role_permission"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the RolePermission.
func (RolePermission) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.RoleID,
		mixin.PermissionID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
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
		index.Fields("id", "created_at").Unique(),
		index.Fields("role_id", "permission_id"),
	}
}
