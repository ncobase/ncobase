package schema

import (
	"ncobase/internal/data/schema/mixin"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// UserDomain holds the schema definition for the UserDomain entity.
type UserDomain struct {
	ent.Schema
}

// Annotations of the UserDomain.
func (UserDomain) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "user_domain"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the UserDomain.
func (UserDomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.DomainID,
	}
}

// Fields of the UserDomain.
func (UserDomain) Fields() []ent.Field {
	return nil
}

// Edges of the UserDomain.
func (UserDomain) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserDomain.
func (UserDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "domain_id"),
	}
}
