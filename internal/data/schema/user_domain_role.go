package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserDomainRole holds the schema definition for the UserDomainRole entity.
type UserDomainRole struct {
	ent.Schema
}

// Annotations of the UserDomainRole.
func (UserDomainRole) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "user_domain_role"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserDomainRole.
func (UserDomainRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.DomainID,
		mixin.RoleID,
	}
}

// Fields of the UserDomainRole.
func (UserDomainRole) Fields() []ent.Field {
	return nil
}

// Edges of the UserDomainRole.
func (UserDomainRole) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserDomainRole.
func (UserDomainRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "domain_id", "role_id"),
	}
}
