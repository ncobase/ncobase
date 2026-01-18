package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/index"
)

// SpaceDictionary holds the schema definition for the SpaceDictionary entity.
type SpaceDictionary struct {
	ent.Schema
}

// Annotations of the SpaceDictionary.
func (SpaceDictionary) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "space", "dictionary"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the SpaceDictionary.
func (SpaceDictionary) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.SpaceID,
		mixin.DictionaryID,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the SpaceDictionary.
func (SpaceDictionary) Fields() []ent.Field {
	return nil
}

// Edges of the SpaceDictionary.
func (SpaceDictionary) Edges() []ent.Edge {
	return nil
}

// Indexes of the SpaceDictionary.
func (SpaceDictionary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("space_id", "dictionary_id"),
	}
}
