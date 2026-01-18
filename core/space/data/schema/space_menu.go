package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// SpaceMenu holds the schema definition for the SpaceMenu entity.
type SpaceMenu struct {
	ent.Schema
}

// Annotations of the SpaceMenu.
func (SpaceMenu) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "space_menu"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceMenu.
func (SpaceMenu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.MenuID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceMenu.
func (SpaceMenu) Fields() []ent.Field {
	return nil
}

// Edges of the SpaceMenu.
func (SpaceMenu) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceMenu.
func (SpaceMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "menu_id"),
	}
}
