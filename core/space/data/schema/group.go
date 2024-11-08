package schema

import (
	"ncobase/common/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

// Annotations of the Group.
func (Group) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "org", "group"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Group.
func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.SlugUnique,
		mixin.Disabled,
		mixin.Description,
		mixin.Leader,
		mixin.ExtraProps,
		mixin.ParentID,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Group.
func (Group) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Group.
func (Group) Edges() []ent.Edge {
	return nil
}

// Indexes of the Group.
func (Group) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
