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

// PaymentOrder holds the schema definition for the PaymentOrder entity.
type PaymentOrder struct {
	ent.Schema
}

// Annotations of the PaymentOrder.
func (PaymentOrder) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "pay", "order"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the PaymentOrder.
func (PaymentOrder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the PaymentOrder.
func (PaymentOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_number").
			Comment("Unique order number").
			Unique().
			NotEmpty(),
		field.Float("amount").
			Comment("Payment amount").
			Positive(),
		field.String("currency").
			Comment("Currency code (USD, EUR, GBP, etc.)").
			Default("USD"),
		field.String("status").
			Comment("Payment status (pending, completed, failed, refunded, cancelled)").
			Default("pending"),
		field.String("type").
			Comment("Payment type (one_time, subscription, recurring)").
			Default("one_time"),
		field.String("channel_id").
			Comment("Payment channel ID").
			NotEmpty(),
		field.String("user_id").
			Comment("User ID").
			NotEmpty(),
		field.String("space_id").
			Comment("Space ID").
			Optional(),
		field.String("product_id").
			Comment("PaymentProduct ID if associated with a product").
			Optional(),
		field.String("subscription_id").
			Comment("PaymentSubscription ID if associated with a subscription").
			Optional(),
		field.Time("expires_at").
			Comment("Expiration time for the payment"),
		field.Time("paid_at").
			Comment("Time when the payment was completed").
			Optional().
			Nillable(),
		field.String("provider_ref").
			Comment("Reference ID from the payment provider").
			Optional(),
		field.String("description").
			Comment("Payment description").
			Optional(),
	}
}

// Edges of the PaymentOrder.
func (PaymentOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("logs", PaymentLog.Type),
	}
}

// Indexes of the PaymentOrder.
func (PaymentOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_number").Unique(),
		index.Fields("status"),
		index.Fields("user_id"),
		index.Fields("space_id"),
		index.Fields("channel_id"),
		index.Fields("product_id"),
		index.Fields("subscription_id"),
	}
}
