package schema

import (
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/entgo/mixin"
)

// SpaceQuota holds the schema definition for the SpaceQuota entity
type SpaceQuota struct {
	ent.Schema
}

// Annotations of the SpaceQuota
func (SpaceQuota) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "space", "quota"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceQuota
func (SpaceQuota) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceQuota
func (SpaceQuota) Fields() []ent.Field {
	return []ent.Field{
		field.String("quota_type").
			NotEmpty().
			Comment("Type of quota (users, storage, api_calls, etc.)"),
		field.String("quota_name").
			NotEmpty().
			Comment("Human readable name of the quota"),
		field.Int64("max_value").
			NonNegative().
			Comment("Maximum allowed value for this quota"),
		field.Int64("current_used").
			Default(0).
			NonNegative().
			Comment("Current usage of this quota"),
		field.String("unit").
			Default("count").
			Comment("Unit of measurement (count, bytes, mb, gb, tb)"),
		field.Bool("enabled").
			Default(true).
			Comment("Whether this quota is actively enforced"),
	}
}

// Edges of the SpaceQuota
func (SpaceQuota) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceQuota
func (SpaceQuota) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "quota_type").Unique(),
		index.Fields("space_id", "enabled"),
	}
}
