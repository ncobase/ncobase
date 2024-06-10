package schema

import (
	"stocms/internal/data/schema/mixin"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
)

// Domain holds the schema definition for the Domain entity.
type Domain struct {
	ent.Schema
}

// Annotations of the Domain.
func (Domain) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "domain"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Domain.
func (Domain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey{},
		mixin.Name{},
		mixin.Title{},
		mixin.URL{},
		mixin.Logo{},
		mixin.LogoAlt{},
		mixin.Keywords{},
		mixin.Copyright{},
		mixin.Description{},
		mixin.Order{},
		mixin.Disabled{},
		mixin.ExtraProps{},
		mixin.CreatedBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Domain.
func (Domain) Fields() []ent.Field {
	return nil
}

// Edges of the Domain.
func (Domain) Edges() []ent.Edge {
	return nil
}

// Indexes of the Domain.
func (Domain) Indexes() []ent.Index {
	return nil
}
