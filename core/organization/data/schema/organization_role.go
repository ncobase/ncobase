package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// OrganizationRole holds the schema definition for the OrganizationRole entity.
type OrganizationRole struct {
	ent.Schema
}

// Annotations of the OrganizationRole.
func (OrganizationRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "organization_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the OrganizationRole.
func (OrganizationRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.OrgID,
		mixin.RoleID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the OrganizationRole.
func (OrganizationRole) Fields() []ent.Field {
	return nil
}

// Edges of the OrganizationRole.
func (OrganizationRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the OrganizationRole.
func (OrganizationRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("org_id", "role_id"),
	}
}
