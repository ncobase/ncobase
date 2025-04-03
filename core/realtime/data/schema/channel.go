package schema

import (
	"ncobase/ncore/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Channel holds the schema definition for the Channel entity.
type Channel struct {
	ent.Schema
}

// Annotations of the Channel.
func (Channel) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "rt", "channel"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Channel.
func (Channel) Mixin() []ent.Mixin {
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

// Fields of the Channel.
func (Channel) Fields() []ent.Field {
	return nil
}

// Edges of the Channel.
func (Channel) Edges() []ent.Edge {
	return nil
}

// Indexes of the Channel.
func (Channel) Indexes() []ent.Index {
	return []ent.Index{}
}
