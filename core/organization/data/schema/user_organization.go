package schema

import (
	"strings"

	"entgo.io/ent/schema/field"
	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserOrganization holds the schema definition for the UserOrganization entity.
type UserOrganization struct {
	ent.Schema
}

// Annotations of the UserOrganization.
func (UserOrganization) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "user_organization"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserOrganization.
func (UserOrganization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UserID,
		mixin.OrgID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the UserOrganization.
func (UserOrganization) Fields() []ent.Field {
	return []ent.Field{
		field.String("role").
			Default("member").
			Comment("Role of the user in the organization").
			NotEmpty(),
	}
}

// Edges of the UserOrganization.
func (UserOrganization) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserOrganization.
func (UserOrganization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "org_id").Unique(),
		index.Fields("user_id", "role"),
		index.Fields("org_id", "role"),
	}
}
