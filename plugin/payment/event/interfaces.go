package event

import (
	"context"
)

const (
	// Payment events

	PaymentCreated   = "payment.created"
	PaymentSucceeded = "payment.succeeded"
	PaymentFailed    = "payment.failed"
	PaymentCancelled = "payment.cancelled"
	PaymentExpired   = "payment.expired"
	PaymentRefunded  = "payment.refunded"

	// Subscription events

	SubscriptionCreated   = "subscription.created"
	SubscriptionRenewed   = "subscription.renewed"
	SubscriptionUpdated   = "subscription.updated"
	SubscriptionCancelled = "subscription.cancelled"
	SubscriptionExpired   = "subscription.expired"

	// Product events

	ProductCreated = "product.created"
	ProductUpdated = "product.updated"
	ProductDeleted = "product.deleted"

	// Channel events

	ChannelCreated   = "payment_channel.created"
	ChannelUpdated   = "payment_channel.updated"
	ChannelDeleted   = "payment_channel.deleted"
	ChannelActivated = "payment_channel.activated"
	ChannelDisabled  = "payment_channel.disabled"
)

// PublisherInterface defines an interface for publishing events
type PublisherInterface interface {
	// Payment events

	PublishPaymentCreated(ctx context.Context, data *PaymentEventData)
	PublishPaymentSucceeded(ctx context.Context, data *PaymentEventData)
	PublishPaymentFailed(ctx context.Context, data *PaymentEventData)
	PublishPaymentCancelled(ctx context.Context, data *PaymentEventData)
	PublishPaymentRefunded(ctx context.Context, data *PaymentEventData)
	PublishPaymentExpired(ctx context.Context, data *PaymentEventData)

	// Subscription events

	PublishSubscriptionCreated(ctx context.Context, data *SubscriptionEventData)
	PublishSubscriptionRenewed(ctx context.Context, data *SubscriptionEventData)
	PublishSubscriptionUpdated(ctx context.Context, data *SubscriptionEventData)
	PublishSubscriptionCancelled(ctx context.Context, data *SubscriptionEventData)
	PublishSubscriptionExpired(ctx context.Context, data *SubscriptionEventData)

	// Product events

	PublishProductCreated(ctx context.Context, data *ProductEventData)
	PublishProductUpdated(ctx context.Context, data *ProductEventData)
	PublishProductDeleted(ctx context.Context, data *ProductEventData)

	// Channel events

	PublishChannelCreated(ctx context.Context, data *ChannelEventData)
	PublishChannelUpdated(ctx context.Context, data *ChannelEventData)
	PublishChannelDeleted(ctx context.Context, data *ChannelEventData)
	PublishChannelActivated(ctx context.Context, data *ChannelEventData)
	PublishChannelDisabled(ctx context.Context, data *ChannelEventData)
}

// Handler defines a generic event handler function
type Handler func(any)

// HandlerProvider defines an interface for providing event handlers
type HandlerProvider interface {
	// GetHandlers Get event handlers
	GetHandlers() map[string]Handler
}
