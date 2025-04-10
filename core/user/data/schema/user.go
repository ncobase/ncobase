package schema

import (
	"strings"

	"github.com/ncobase/ncore/pkg/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "user"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.UsernameUnique,
		mixin.Password,
		mixin.Email,
		mixin.Phone,
		mixin.IsCertified,
		mixin.IsAdmin,
		mixin.Status, // status, 0: activated, 1: unactivated, 2: disabled
		mixin.ExtraProps,
		mixin.CreatedAt,
		mixin.UpdatedAt,
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return nil
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
