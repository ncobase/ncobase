package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// SpaceOption holds the schema definition for the SpaceOption entity.
type SpaceOption struct {
	ent.Schema
}

// Annotations of the SpaceOption.
func (SpaceOption) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "space_option"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceOption.
func (SpaceOption) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.OptionID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceOption.
func (SpaceOption) Fields() []ent.Field {
	return nil
}

// Edges of the SpaceOption.
func (SpaceOption) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceOption.
func (SpaceOption) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "option_id"),
	}
}
