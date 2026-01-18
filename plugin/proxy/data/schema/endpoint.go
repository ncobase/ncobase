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

// Endpoint holds the schema definition for the Endpoint entity.
type Endpoint struct {
	ent.Schema
}

// Annotations of the Endpoint.
func (Endpoint) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "proxy", "endpoint"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Endpoint.
func (Endpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.Disabled,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the Endpoint.
func (Endpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("base_url").
			Comment("Base URL of the third-party API endpoint").
			NotEmpty(),
		field.String("protocol").
			Comment("Protocol (HTTP, HTTPS, WS, WSS, TCP, UDP)").
			Default("HTTPS"),
		field.String("auth_type").
			Comment("Authentication type (None, Basic, Bearer, OAuth, ApiKey)").
			Default("None"),
		field.String("auth_config").
			Comment("Authentication configuration in JSON format").
			Optional(),
		field.Int("timeout").
			Comment("Request timeout in seconds").
			Default(30),
		field.Bool("use_circuit_breaker").
			Comment("Whether to use circuit breaker for this endpoint").
			Default(true),
		field.Int("retry_count").
			Comment("Number of retry attempts").
			Default(3),
		field.Bool("validate_ssl").
			Comment("Whether to validate SSL certificates").
			Default(true),
		field.Bool("log_requests").
			Comment("Whether to log request details").
			Default(true),
		field.Bool("log_responses").
			Comment("Whether to log response details").
			Default(true),
	}
}

// Edges of the Endpoint.
func (Endpoint) Edges() []ent.Edge {
	return nil
}

// Indexes of the Endpoint.
func (Endpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}
