package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// RTChannel holds the schema definition for the  RTChannel entity.
type RTChannel struct {
	ent.Schema
}

// Annotations of the  RTChannel.
func (RTChannel) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "rt", "channel"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the  RTChannel.
func (RTChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.Type,   // public, private, direct
		mixin.Status, // 0: disabled, 1: enabled
		mixin.ExtraProps,
		mixin.TimeAt{},
	}
}

// Fields of the  RTChannel.
func (RTChannel) Fields() []ent.Field {
	return nil
}

// Edges of the  RTChannel.
func (RTChannel) Edges() []ent.Edge {
	return nil
}

// Indexes of the  RTChannel.
func (RTChannel) Indexes() []ent.Index {
	return []ent.Index{}
}
