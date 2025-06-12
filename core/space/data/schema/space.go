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

// Space holds the schema definition for the Space entity.
type Space struct {
	ent.Schema
}

// Annotations of the Space.
func (Space) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "space"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Space.
func (Space) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.SlugUnique,
		mixin.Type,
		mixin.Title,
		mixin.URL,
		mixin.Logo,
		mixin.LogoAlt,
		mixin.Keywords,
		mixin.Copyright,
		mixin.Description,
		mixin.Order,
		mixin.Disabled,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.ExpiredAt,
		mixin.TimeAt{},
	}
}

// Fields of the Space.
func (Space) Fields() []ent.Field {
	return nil
}

// Edges of the Space.
func (Space) Edges() []ent.Edge {
	return nil
}

// Indexes of the Space.
func (Space) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
