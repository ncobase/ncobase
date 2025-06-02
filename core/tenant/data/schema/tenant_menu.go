package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// TenantMenu holds the schema definition for the TenantMenu entity.
type TenantMenu struct {
	ent.Schema
}

// Annotations of the TenantMenu.
func (TenantMenu) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "tenant_menu"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantMenu.
func (TenantMenu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.MenuID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantMenu.
func (TenantMenu) Fields() []ent.Field {
	return nil
}

// Edges of the TenantMenu.
func (TenantMenu) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantMenu.
func (TenantMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "menu_id"),
	}
}
