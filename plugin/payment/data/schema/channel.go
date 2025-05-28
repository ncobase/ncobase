package schema

import (
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/ncobase/ncore/data/databases/entgo/mixin"
)

// PaymentChannel holds the schema definition for the PaymentChannel entity.
type PaymentChannel struct {
	ent.Schema
}

// Annotations of the PaymentChannel.
func (PaymentChannel) Annotations() []schema.Annotation {
	table := strings.Join([]string{"ncse", "pay", "channel"}, "_")
	return []schema.Annotation{
		entsql.Annotation{Table: table},
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		entsql.WithComments(true),
	}
}

// Mixin of the PaymentChannel.
func (PaymentChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.PrimaryKey,
		mixin.Name,
		mixin.Description,
		mixin.ExtraProps,
		mixin.OperatorBy{},
		mixin.TimeAt{},
	}
}

// Fields of the PaymentChannel.
func (PaymentChannel) Fields() []ent.Field {
	return []ent.Field{
		field.String("provider").
			Comment("Payment provider (stripe, paypal, alipay, wechatpay, etc.)").
			NotEmpty(),
		field.String("status").
			Comment("Payment channel status (active, disabled, testing)").
			Default("active"),
		field.Bool("is_default").
			Comment("Whether this is the default channel for the provider").
			Default(false),
		field.JSON("supported_types", []string{}).
			Comment("Supported payment types (one_time, subscription, recurring)"),
		field.JSON("config", map[string]any{}).
			Comment("Provider-specific configuration"),
		field.String("tenant_id").
			Comment("Tenant ID if multi-tenant support is enabled").
			Optional(),
	}
}

// Edges of the PaymentChannel.
func (PaymentChannel) Edges() []ent.Edge {
	return nil
}

// Indexes of the PaymentChannel.
func (PaymentChannel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider"),
		index.Fields("status"),
		index.Fields("tenant_id"),
		index.Fields("provider", "is_default"),
	}
}
