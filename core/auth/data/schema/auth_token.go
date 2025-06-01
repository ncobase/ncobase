package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
)

// AuthToken holds the schema definition for the AuthToken entity.
type AuthToken struct {
	ent.Schema
}

// Annotations of the AuthToken.
func (AuthToken) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "auth_token"}, "_")
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

// Indexes of the AuthToken.
func (AuthToken) Indexes() []ent.Index {
	return []ent.Index{
		// Create a composite index on id and created_at
		index.Fields("id", "created_at").Unique(),
	}
}
