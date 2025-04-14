package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"

	"entgo.io/ent"
)

// Dictionary holds the schema definition for the Dictionary entity.
type Dictionary struct {
	ent.Schema
}

// Annotations of the Dictionary.
func (Dictionary) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "dictionary"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Dictionary.
func (Dictionary) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.SlugUnique,
		mixin.Type,  // type, object, string, number, ...
		mixin.Value, // type value
		mixin.TenantID,
		mixin.Description,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Dictionary.
func (Dictionary) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Dictionary.
func (Dictionary) Edges() []ent.Edge {
	return nil
}

// Indexes of the Dictionary.
func (Dictionary) Indexes() []ent.Index {
	return []ent.Index{}
}
