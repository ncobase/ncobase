package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/databases/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// TenantDictionary holds the schema definition for the TenantDictionary entity.
type TenantDictionary struct {
	ent.Schema
}

// Annotations of the TenantDictionary.
func (TenantDictionary) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "sys", "tenant_dictionary"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantDictionary.
func (TenantDictionary) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.DictionaryID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantDictionary.
func (TenantDictionary) Fields() []ent.Field {
	return nil
}

// Edges of the TenantDictionary.
func (TenantDictionary) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantDictionary.
func (TenantDictionary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "dictionary_id"),
	}
}
