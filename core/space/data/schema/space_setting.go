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

// SpaceSetting holds the schema definition for the SpaceSetting entity
type SpaceSetting struct {
	ent.Schema
}

// Annotations of the SpaceSetting
func (SpaceSetting) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "space_setting"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceSetting
func (SpaceSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceSetting
func (SpaceSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("setting_key").
			NotEmpty().
			Comment("Unique key for the setting"),
		field.String("setting_name").
			NotEmpty().
			Comment("Human readable name of the setting"),
		field.Text("setting_value").
			Optional().
			Comment("Current value of the setting"),
		field.Text("default_value").
			Optional().
			Comment("Default value of the setting"),
		field.String("setting_type").
			Default("string").
			Comment("Data type of the setting value"),
		field.String("scope").
			Default("space").
			Comment("Scope of the setting (system, space, user, feature)"),
		field.String("category").
			Default("general").
			Comment("Category grouping for settings"),
		field.Bool("is_public").
			Default(false).
			Comment("Whether setting is publicly readable"),
		field.Bool("is_required").
			Default(false).
			Comment("Whether setting is required"),
		field.Bool("is_readonly").
			Default(false).
			Comment("Whether setting is read-only"),
		field.JSON("validation", map[string]any{}).
			Optional().
			Comment("Validation rules for the setting value"),
	}
}

// Edges of the SpaceSetting
func (SpaceSetting) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceSetting
func (SpaceSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "setting_key").Unique(),
		index.Fields("space_id", "category"),
		index.Fields("space_id", "scope"),
		index.Fields("is_public"),
	}
}
