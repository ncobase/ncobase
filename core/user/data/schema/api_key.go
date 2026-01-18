package schema

import (
	"strings"

	"github.com/ncobase/ncore/data/entgo/mixin"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ApiKey schema for API keys
type ApiKey struct {
	ent.Schema
}

// Annotations of the ApiKey
func (ApiKey) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "user", "api_key"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the ApiKey
func (ApiKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.TimeAt{},
		mixin.UserID,
	}
}

// Fields of the ApiKey
func (ApiKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").NotEmpty().Unique(),
		field.Int64("last_used").Optional(),
	}
}

// Indexes of the ApiKey
func (ApiKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("user_id", "created_at"),
		index.Fields("key").Unique(),
	}
}
