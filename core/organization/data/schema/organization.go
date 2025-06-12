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

// Organization holds the schema definition for the Organization entity.
type Organization struct {
	ent.Schema
}

// Annotations of the Organization.
func (Organization) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "organization"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Organization.
func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.SlugUnique,
		mixin.Type,
		mixin.Disabled,
		mixin.Description,
		mixin.Leader,
		mixin.ExtraProps,
		mixin.ParentID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Organization.
func (Organization) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Organization.
func (Organization) Edges() []ent.Edge {
	return nil
}

// Indexes of the Organization.
func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
