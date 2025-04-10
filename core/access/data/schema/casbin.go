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

// CasbinRule holds the schema definition for the CasbinRule entity.
type CasbinRule struct {
	ent.Schema
}

// Annotations of the CasbinRule.
func (CasbinRule) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "casbin_rule"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the CasbinRule.
func (CasbinRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.PType,
		mixin.V0,
		mixin.V1,
		mixin.V2,
		mixin.V3,
		mixin.V4,
		mixin.V5,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the CasbinRule.
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{}
}

// Edges of the CasbinRule.
func (CasbinRule) Edges() []ent.Edge {
	return nil
}

// Indexes of the CasbinRule.
func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
	}
}

// Policy of the CasbinRule.
func (CasbinRule) Policy() ent.Policy {
	return nil
}
