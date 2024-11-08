package schema

import (
	"ncobase/common/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// UserProfile holds the schema definition for the UserProfile entity.
type UserProfile struct {
	ent.Schema
}

// Annotations of the UserProfile.
func (UserProfile) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "user_profile"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the UserProfile.
func (UserProfile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.NewPrimaryKeyAlias("user", "user_id"),
		mixin.DisplayName,
		mixin.ShortBio,
		mixin.About,
		mixin.Links,
		mixin.Thumbnail,
		mixin.ExtraProps,
	}
}

// Fields of the UserProfile.
func (UserProfile) Fields() []ent.Field {
	return nil
}

// Edges of the UserProfile.
func (UserProfile) Edges() []ent.Edge {
	return nil
}

// Indexes of the UserProfile.
func (UserProfile) Indexes() []ent.Index {
	return nil
}
