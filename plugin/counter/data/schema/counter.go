package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"

	"entgo.io/ent"
)

// Counter holds the schema definition for the Counter entity.
type Counter struct {
	ent.Schema
}

// Annotations of the Counter.
func (Counter) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "counter"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Counter.
func (Counter) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Identifier,
		mixin.Name,
		mixin.Prefix,
		mixin.Suffix,
		mixin.StartValue,
		mixin.IncrementStep,
		mixin.DateFormat("200601"),
		mixin.CurrentValue,
		mixin.Disabled,
		mixin.Description,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Counter.
func (Counter) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Counter.
func (Counter) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Counter.
func (Counter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
