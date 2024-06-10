package schema

import (
	"stocms/pkg/nanoid"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// CasbinRule holds the schema definition for the CasbinRule entity.
type CasbinRule struct {
	ent.Schema
}

// Annotations of the CasbinRule.
func (CasbinRule) Annotations() []schema.Annotation {
	table := strings.Join([]string{"sc", "casbin_rule"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Fields of the CasbinRule.
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").DefaultFunc(nanoid.PrimaryKey()),      // primary key
		field.String("p_type").Nillable().Optional().MaxLen(255), // p_type, ref: https://casbin.org/docs/zh-CN/policy-storage
		field.String("v0").Nillable().Optional().MaxLen(255),     // v0
		field.String("v1").Nillable().Optional().MaxLen(255),     // v1
		field.String("v2").Nillable().Optional().MaxLen(255),     // v2
		field.String("v3").Nillable().Optional().MaxLen(255),     // v3
		field.String("v4").Nillable().Optional().MaxLen(255),     // v4
		field.String("v5").Nillable().Optional().MaxLen(255),     // v5
	}
}

// Edges of the CasbinRule.
func (CasbinRule) Edges() []ent.Edge {
	return nil
}

// Indexes of the CasbinRule.
func (CasbinRule) Indexes() []ent.Index {
	return []ent.Index{}
}

// Policy of the CasbinRule.
func (CasbinRule) Policy() ent.Policy {
	return nil
}
