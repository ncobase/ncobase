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
	"github.com/ncobase/ncore/types"
)

// AIRun holds the schema definition for AI execution records.
type AIRun struct {
	ent.Schema
}

func (AIRun) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "ai", "run"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

func (AIRun) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

func (AIRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("operation_id").
			Optional().
			Comment("Client or server operation id for correlating retries and streams"),
		field.String("action").
			Optional().
			Comment("Business action key, for example content.summary or builder.schema"),
		field.String("mode").
			Default("complete").
			Comment("AI execution mode: complete, stream, embed, action"),
		field.String("status").
			Default("running").
			Comment("AI execution status: running, succeeded, failed, canceled"),
		field.String("provider").
			Optional().
			Comment("Resolved provider name"),
		field.String("model").
			Optional().
			Comment("Resolved provider model"),
		field.String("fallback_model").
			Optional().
			Comment("Fallback model used when it differs from the configured primary"),
		field.Int("input_tokens").
			Default(0).
			Comment("Input token count reported by the provider"),
		field.Int("output_tokens").
			Default(0).
			Comment("Output token count reported by the provider"),
		field.Int("total_tokens").
			Default(0).
			Comment("Total token count reported by the provider"),
		field.Int("reasoning_tokens").
			Default(0).
			Comment("Reasoning token count reported by models that expose it"),
		field.Int("cache_created_tokens").
			Default(0).
			Comment("Tokens written to prompt cache"),
		field.Int("cache_read_tokens").
			Default(0).
			Comment("Tokens read from prompt cache"),
		field.Int64("duration_ms").
			Default(0).
			Comment("Execution duration in milliseconds"),
		field.String("error_code").
			Optional().
			Comment("Classified error code"),
		field.Text("error_message").
			Optional().
			Comment("Sanitized error message"),
		field.String("request_hash").
			Optional().
			Comment("SHA-256 hash of sanitized request payload"),
		field.String("response_hash").
			Optional().
			Comment("SHA-256 hash of response content"),
		field.Float("estimated_cost").
			Default(0).
			Comment("Estimated cost using configured token pricing"),
		field.String("currency").
			Optional().
			Comment("Cost currency"),
		field.JSON("metadata", types.JSON{}).
			Optional().
			Comment("Non-secret request metadata"),
		field.String("space_id").
			Optional().
			Comment("Active space id"),
		field.String("user_id").
			Optional().
			Comment("Actor user id"),
	}
}

func (AIRun) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (AIRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("operation_id"),
		index.Fields("action"),
		index.Fields("mode"),
		index.Fields("status"),
		index.Fields("provider"),
		index.Fields("model"),
		index.Fields("space_id", "created_at"),
		index.Fields("user_id", "created_at"),
		index.Fields("action", "created_at"),
		index.Fields("status", "created_at"),
		index.Fields("created_at"),
	}
}
