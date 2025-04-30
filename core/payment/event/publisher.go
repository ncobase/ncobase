package event

import (
	"context"
)

// DefaultPublisher provides methods to publish payment-related events
type DefaultPublisher struct {
	em *Manager
}

// NewPublisher creates a new event publisher
func NewPublisher(manager *Manager) Publisher {
	return &DefaultPublisher{
		em: manager,
	}
}

// PublishPaymentCreated publishes a payment.created event
func (p *DefaultPublisher) PublishPaymentCreated(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentCreated, data)
}

// PublishPaymentSucceeded publishes a payment.succeeded event
func (p *DefaultPublisher) PublishPaymentSucceeded(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentSucceeded, data)
}

// PublishPaymentFailed publishes a payment.failed event
func (p *DefaultPublisher) PublishPaymentFailed(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentFailed, data)
}

// PublishPaymentCancelled publishes a payment.cancelled event
func (p *DefaultPublisher) PublishPaymentCancelled(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentCancelled, data)
}

// PublishPaymentExpired publishes a payment.expired event
func (p *DefaultPublisher) PublishPaymentExpired(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentExpired, data)
}

// PublishPaymentRefunded publishes a payment.refunded event
func (p *DefaultPublisher) PublishPaymentRefunded(ctx context.Context, data *PaymentEventData) {
	p.em.PublishPaymentEvent(ctx, PaymentRefunded, data)
}

// PublishSubscriptionCreated publishes a subscription.created event
func (p *DefaultPublisher) PublishSubscriptionCreated(ctx context.Context, data *SubscriptionEventData) {
	p.em.PublishSubscriptionEvent(ctx, SubscriptionCreated, data)
}

// PublishSubscriptionRenewed publishes a subscription.renewed event
func (p *DefaultPublisher) PublishSubscriptionRenewed(ctx context.Context, data *SubscriptionEventData) {
	p.em.PublishSubscriptionEvent(ctx, SubscriptionRenewed, data)
}

// PublishSubscriptionUpdated publishes a subscription.updated event
func (p *DefaultPublisher) PublishSubscriptionUpdated(ctx context.Context, data *SubscriptionEventData) {
	p.em.PublishSubscriptionEvent(ctx, SubscriptionUpdated, data)
}

// PublishSubscriptionCancelled publishes a subscription.cancelled event
func (p *DefaultPublisher) PublishSubscriptionCancelled(ctx context.Context, data *SubscriptionEventData) {
	p.em.PublishSubscriptionEvent(ctx, SubscriptionCancelled, data)
}

// PublishSubscriptionExpired publishes a subscription.expired event
func (p *DefaultPublisher) PublishSubscriptionExpired(ctx context.Context, data *SubscriptionEventData) {
	p.em.PublishSubscriptionEvent(ctx, SubscriptionExpired, data)
}

// PublishProductCreated publishes a product.created event
func (p *DefaultPublisher) PublishProductCreated(ctx context.Context, data *ProductEventData) {
	p.em.PublishProductEvent(ctx, ProductCreated, data)
}

// PublishProductUpdated publishes a product.updated event
func (p *DefaultPublisher) PublishProductUpdated(ctx context.Context, data *ProductEventData) {
	p.em.PublishProductEvent(ctx, ProductUpdated, data)
}

// PublishProductDeleted publishes a product.deleted event
func (p *DefaultPublisher) PublishProductDeleted(ctx context.Context, data *ProductEventData) {
	p.em.PublishProductEvent(ctx, ProductDeleted, data)
}

// PublishChannelCreated publishes a payment_channel.created event
func (p *DefaultPublisher) PublishChannelCreated(ctx context.Context, data *ChannelEventData) {
	p.em.PublishChannelEvent(ctx, ChannelCreated, data)
}

// PublishChannelUpdated publishes a payment_channel.updated event
func (p *DefaultPublisher) PublishChannelUpdated(ctx context.Context, data *ChannelEventData) {
	p.em.PublishChannelEvent(ctx, ChannelUpdated, data)
}

// PublishChannelDeleted publishes a payment_channel.deleted event
func (p *DefaultPublisher) PublishChannelDeleted(ctx context.Context, data *ChannelEventData) {
	p.em.PublishChannelEvent(ctx, ChannelDeleted, data)
}

// PublishChannelActivated publishes a payment_channel.activated event
func (p *DefaultPublisher) PublishChannelActivated(ctx context.Context, data *ChannelEventData) {
	p.em.PublishChannelEvent(ctx, ChannelActivated, data)
}

// PublishChannelDisabled publishes a payment_channel.disabled event
func (p *DefaultPublisher) PublishChannelDisabled(ctx context.Context, data *ChannelEventData) {
	p.em.PublishChannelEvent(ctx, ChannelDisabled, data)
}
