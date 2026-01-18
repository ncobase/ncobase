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

// PaymentProduct holds the schema definition for the PaymentProduct entity.
type PaymentProduct struct {
	ent.Schema
}

// Annotations of the PaymentProduct.
func (PaymentProduct) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "pay", "product"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the PaymentProduct.
func (PaymentProduct) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the PaymentProduct.
func (PaymentProduct) Fields() []ent.Field {
	return []ent.Field{
		field.String("status").
			Comment("PaymentProduct status (active, disabled, draft)").
			Default("active"),
		field.String("pricing_type").
			Comment("Pricing type (one_time, recurring, usage_based, tiered_usage)").
			Default("one_time"),
		field.Float("price").
			Comment("Base price amount").
			Positive(),
		field.String("currency").
			Comment("Currency code (USD, EUR, GBP, etc.)").
			Default("USD"),
		field.String("billing_interval").
			Comment("Billing interval for recurring payments (daily, weekly, monthly, yearly)").
			Optional(),
		field.Int("trial_days").
			Comment("Number of trial days for recurring subscriptions").
			Default(0),
		field.JSON("features", []string{}).
			Comment("List of features included in the product"),
		field.String("space_id").
			Comment("Space ID").
			Optional(),
	}
}

// Edges of the PaymentProduct.
func (PaymentProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("subscriptions", PaymentSubscription.Type),
	}
}

// Indexes of the PaymentProduct.
func (PaymentProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("pricing_type"),
		index.Fields("space_id"),
	}
}
