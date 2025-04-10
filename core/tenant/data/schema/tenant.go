package schema

import (
	"github.com/ncobase/ncore/pkg/data/entgo/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// Tenant holds the schema definition for the Tenant entity.
type Tenant struct {
	ent.Schema
}

// Annotations of the Tenant.
func (Tenant) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "tenant"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Tenant.
func (Tenant) Mixin() []ent.Mixin {
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

// Fields of the Tenant.
func (Tenant) Fields() []ent.Field {
	return nil
}

// Edges of the Tenant.
func (Tenant) Edges() []ent.Edge {
	return nil
}

// Indexes of the Tenant.
func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}
