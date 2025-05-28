package schema

import (
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
)

// TenantBilling holds the schema definition for the TenantBilling entity
type TenantBilling struct {
	ent.Schema
}

// Annotations of the TenantBilling
func (TenantBilling) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "iam", "tenant_billing"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entsql.WithComments(true),
	}
}

// Mixin of the TenantBilling
func (TenantBilling) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.TenantID,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the TenantBilling
func (TenantBilling) Fields() []ent.Field {
	return []ent.Field{
		field.String("billing_period").
			Default("monthly").
			Comment("Billing period type (monthly, yearly, one_time, usage_based)"),
		field.Int64("period_start").
			Optional().
			Comment("Start timestamp of billing period"),
		field.Int64("period_end").
			Optional().
			Comment("End timestamp of billing period"),
		field.Float("amount").
			Positive().
			Comment("Billing amount"),
		field.String("currency").
			Default("USD").
			Comment("Currency code (USD, EUR, etc.)"),
		field.String("status").
			Default("pending").
			Comment("Billing status (pending, paid, overdue, cancelled, refunded)"),
		field.String("invoice_number").
			Optional().
			Comment("Invoice or reference number"),
		field.String("payment_method").
			Optional().
			Comment("Payment method used"),
		field.Int64("paid_at").
			Optional().
			Comment("Payment timestamp"),
		field.Int64("due_date").
			Optional().
			Comment("Payment due date timestamp"),
		field.JSON("usage_details", map[string]any{}).
			Optional().
			Comment("Detailed usage information for billing period"),
	}
}

// Edges of the TenantBilling
func (TenantBilling) Edges() []ent.Edge {
	return nil
}

// Indexes of the TenantBilling
func (TenantBilling) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "created_at").Unique(),
		index.Fields("tenant_id", "billing_period"),
		index.Fields("tenant_id", "status"),
		index.Fields("status", "due_date"),
		index.Fields("invoice_number").Unique(),
	}
}
