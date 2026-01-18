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

// Logs holds the schema definition for the Logs entity.
type Logs struct {
	ent.Schema
}

// Annotations of the Logs.
func (Logs) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "proxy", "log"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the Logs.
func (Logs) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TimeAt{},
	}
}

// Fields of the Logs.
func (Logs) Fields() []ent.Field {
	return []ent.Field{
		field.String("endpoint_id").
			Comment("ID of the associated endpoint"),
		field.String("route_id").
			Comment("ID of the associated route"),
		field.String("request_method").
			Comment("HTTP method of the request"),
		field.String("request_path").
			Comment("Path of the request"),
		field.Text("request_headers").
			Comment("Headers of the request (JSON format)").
			Optional(),
		field.Text("request_body").
			Comment("Body of the request").
			Optional(),
		field.Int("status_code").
			Comment("HTTP status code of the response"),
		field.Text("response_headers").
			Comment("Headers of the response (JSON format)").
			Optional(),
		field.Text("response_body").
			Comment("Body of the response").
			Optional(),
		field.Int("duration").
			Comment("Duration of the request in milliseconds"),
		field.String("error").
			Comment("Error message if any").
			Optional(),
		field.String("client_ip").
			Comment("IP address of the client").
			Optional(),
		field.String("user_id").
			Comment("ID of the user who made the request").
			Optional(),
	}
}

// Edges of the Logs.
func (Logs) Edges() []ent.Edge {
	return nil
}

// Indexes of the Logs.
func (Logs) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("endpoint_id"),
		index.Fields("route_id"),
		index.Fields("created_at"),
	}
}
