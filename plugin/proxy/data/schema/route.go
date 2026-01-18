package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/entgo/mixin"
)

// Route holds the schema definition for the Route entity.
type Route struct {
	ent.Schema
}

// Annotations of the Route.
func (Route) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "tbp", "route"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Route.
func (Route) Mixin() []ent.Mixin {
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

// Fields of the Route.
func (Route) Fields() []ent.Field {
	return []ent.Field{
		field.String("endpoint_id").
			Comment("ID of the associated endpoint"),
		field.String("path_pattern").
			Comment("Path pattern for this route (e.g., /api/users/:id)").
			NotEmpty(),
		field.String("target_path").
			Comment("Target path on the remote API").
			NotEmpty(),
		field.String("method").
			Comment("HTTP method (GET, POST, PUT, DELETE, etc.)").
			Default("GET"),
		field.String("input_transformer_id").
			Comment("ID of the transformer to apply to incoming requests").
			Optional(),
		field.String("output_transformer_id").
			Comment("ID of the transformer to apply to outgoing responses").
			Optional(),
		field.Bool("cache_enabled").
			Comment("Whether to cache responses").
			Default(false),
		field.Int("cache_ttl").
			Comment("Time to live for cached responses in seconds").
			Default(300),
		field.String("rate_limit").
			Comment("Rate limit expression (e.g., 100/minute)").
			Optional(),
		field.Bool("strip_auth_header").
			Comment("Whether to strip authentication header when forwarding").
			Default(false),
	}
}

// Edges of the Route.
func (Route) Edges() []ent.Edge {
	return nil
}

// Indexes of the Route.
func (Route) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint_id", "path_pattern", "method").
			Unique(),
	}
}
