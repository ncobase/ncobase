package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Options holds the schema definition for the Options entity.
type Options struct {
	ent.Schema
}

// Annotations of the Options.
func (Options) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "option"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Options.
func (Options) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.NameUnique, // name unique
		mixin.Type,       // type, object, string, number, boolean, ...
		mixin.Value,      // value
		mixin.Autoload,   // autoload
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Options.
func (Options) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Options.
func (Options) Edges() []ent.Edge {
	return nil
}

// Indexes of the Options.
func (Options) Indexes() []ent.Index {
	return []ent.Index{}
}
