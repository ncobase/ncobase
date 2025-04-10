package schema

import (
	"strings"

	"github.com/ncobase/ncore/pkg/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Menu holds the schema definition for the Menu entity.
type Menu struct {
	ent.Schema
}

// Annotations of the Menu.
func (Menu) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "menu"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Menu.
func (Menu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Label, // label, i18n key or custom name
		mixin.SlugUnique,
		mixin.Type,   // type, 1: menu, 2: button, 3: submenu, 4: divider, 5: header, 6: group, 7: dropdown, 8: navbar
		mixin.Path,   // path, eg: /admin/user
		mixin.Target, // target, eg: _blank
		mixin.Icon,
		mixin.Perms, // permissions, menu authority code, eg: admin.user.list
		mixin.Hidden,
		mixin.Order,
		mixin.Disabled,
		mixin.ExtraProps,
		mixin.ParentID,
		mixin.TenantID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Menu.
func (Menu) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the Menu.
func (Menu) Edges() []ent.Edge {
	return nil
}

// Indexes of the Menu.
func (Menu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
