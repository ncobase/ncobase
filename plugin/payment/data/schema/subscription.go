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
	"github.com/ncobase/ncore/data/entgo/mixin"
)

// PaymentSubscription holds the schema definition for the PaymentSubscription entity.
type PaymentSubscription struct {
	ent.Schema
}

// Annotations of the PaymentSubscription.
func (PaymentSubscription) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "pay", "subscription"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the PaymentSubscription.
func (PaymentSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the PaymentSubscription.
func (PaymentSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.String("status").
			Comment("PaymentSubscription status (active, trialing, cancelled, expired, past_due)").
			Default("active"),
		field.String("user_id").
			Comment("User ID").
			NotEmpty(),
		field.String("space_id").
			Comment("Space ID").
			Optional(),
		field.String("product_id").
			Comment("PaymentProduct ID").
			NotEmpty(),
		field.String("channel_id").
			Comment("Payment channel ID").
			NotEmpty(),
		field.Time("current_period_start").
			Comment("Start of the current billing period"),
		field.Time("current_period_end").
			Comment("End of the current billing period"),
		field.Time("cancel_at").
			Comment("When to cancel the subscription").
			Optional().
			Nillable(),
		field.Time("cancelled_at").
			Comment("When the subscription was cancelled").
			Optional().
			Nillable(),
		field.Time("trial_start").
			Comment("When the trial started").
			Optional().
			Nillable(),
		field.Time("trial_end").
			Comment("When the trial ends").
			Optional().
			Nillable(),
		field.String("provider_ref").
			Comment("Reference ID from the payment provider").
			Optional(),
	}
}

// Edges of the PaymentSubscription.
func (PaymentSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", PaymentProduct.Type).
			Ref("subscriptions").
			Field("product_id").
			Unique().
			Required(),
	}
}

// Indexes of the PaymentSubscription.
func (PaymentSubscription) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("user_id"),
		index.Fields("space_id"),
		index.Fields("product_id"),
		index.Fields("channel_id"),
		index.Fields("current_period_end"),
	}
}
