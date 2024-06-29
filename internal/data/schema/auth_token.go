package schema

import (
	"ncobase/common/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// AuthToken holds the schema definition for the AuthToken entity.
type AuthToken struct {
	ent.Schema
}

// Annotations of the AuthToken.
func (AuthToken) Annotations() []schema.Annotation {
	table := strings.Join([]string{"nb", "auth_token"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the AuthToken.
func (AuthToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Disabled,
		mixin.TimeAt{},
		mixin.UserID,
	}
}

// Fields of the AuthToken.
func (AuthToken) Fields() []ent.Field {
	return nil
}

// Edges of the AuthToken.
func (AuthToken) Edges() []ent.Edge {
	return nil
}
