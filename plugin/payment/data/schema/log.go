package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
)

// PaymentLog holds the schema definition for the PaymentLog entity.
type PaymentLog struct {
	ent.Schema
}

// Annotations of the PaymentLog.
func (PaymentLog) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "pay", "log"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate()),
		entsql.WithComments(true),
	}
}

// Mixin of the PaymentLog.
func (PaymentLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.ExtraProps,
		mixin.TimeAt{},
	}
}

// Fields of the PaymentLog.
func (PaymentLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_id").
			Comment("Payment order ID").
			NotEmpty(),
		field.String("channel_id").
			Comment("Payment channel ID").
			NotEmpty(),
		field.String("type").
			Comment("Log type (create, update, verify, callback, notify, refund, error)").
			NotEmpty(),
		field.String("status_before").
			Comment("Payment status before the action").
			Optional(),
		field.String("status_after").
			Comment("Payment status after the action").
			Optional(),
		field.Text("request_data").
			Comment("Request data in JSON format").
			Optional(),
		field.Text("response_data").
			Comment("Response data in JSON format").
			Optional(),
		field.String("ip").
			Comment("IP address").
			Optional(),
		field.Text("user_agent").
			Comment("User agent").
			Optional(),
		field.String("user_id").
			Comment("User ID").
			Optional(),
		field.Text("error").
			Comment("Error message if any").
			Optional(),
	}
}

// Edges of the PaymentLog.
func (PaymentLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("order", PaymentOrder.Type).
			Ref("logs").
			Field("order_id").
			Unique().
			Required(),
	}
}

// Indexes of the PaymentLog.
func (PaymentLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_id"),
		index.Fields("channel_id"),
		index.Fields("type"),
		index.Fields("user_id"),
	}
}
