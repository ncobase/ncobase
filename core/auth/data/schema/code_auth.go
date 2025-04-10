package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// CodeAuth holds the schema definition for the CodeAuth entity.
type CodeAuth struct {
	ent.Schema
}

// Annotations of the CodeAuth.
func (CodeAuth) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "code_auth"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the CodeAuth.
func (CodeAuth) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Code,
		mixin.Email,
		mixin.Logged,
		mixin.TimeAt{},
	}
}

// Fields of the CodeAuth.
func (CodeAuth) Fields() []ent.Field {
	return nil
}

// Edges of the CodeAuth.
func (CodeAuth) Edges() []ent.Edge {
	return nil
}

// Indexes of the CodeAuth.
func (CodeAuth) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("code"),
	}
}
