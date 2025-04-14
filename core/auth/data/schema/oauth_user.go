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

// OAuthUser holds the schema definition for the OAuthUser entity.
type OAuthUser struct {
	ent.Schema
}

// Annotations of the OAuthUser.
func (OAuthUser) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "oauth_user"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the OAuthUser.
func (OAuthUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.OAuthID,
		mixin.AccessToken,
		mixin.Provider,
		mixin.UserID,
		mixin.CreatedAt,
		mixin.UpdatedAt,
	}
}

// Fields of the OAuthUser.
func (OAuthUser) Fields() []ent.Field {
	return nil
}

// Edges of the OAuthUser.
func (OAuthUser) Edges() []ent.Edge {
	return nil
}

// Indexes of the OAuthUser.
func (OAuthUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
