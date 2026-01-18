package schema

import (
	"strings"

	"entgo.io/ent/schema/field"
	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// SpaceOrganization holds the schema definition for the SpaceOrganization entity.
type SpaceOrganization struct {
	ent.Schema
}

// Annotations of the SpaceOrganization.
func (SpaceOrganization) Annotations() []schema.Annotation {
	// Keep the original table name to avoid migration issues
	table := strings.Join([]string{"ncse", "space", "organization"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceOrganization.
func (SpaceOrganization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.OrgID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceOrganization.
func (SpaceOrganization) Fields() []ent.Field {
	return []ent.Field{
		field.String("relation_type").
			Default("member").
			Comment("Type of relationship between space and group").
			NotEmpty(),
	}
}

// Edges of the SpaceOrganization.
func (SpaceOrganization) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceOrganization.
func (SpaceOrganization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "org_id").Unique(),
		index.Fields("space_id", "relation_type"),
		index.Fields("org_id", "relation_type"),
	}
}
