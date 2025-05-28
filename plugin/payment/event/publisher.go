package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// provides methods to publish payment-related events
type publisher struct {
	em ext.ManagerInterface
}

// NewPublisher creates a new event publisher
func NewPublisher(em ext.ManagerInterface) PublisherInterface {
	return &publisher{
		em: em,
	}
}

// Generic publish method
func (p *publisher) publish(ctx context.Context, t string, data any) {
	// Log the event based on its type
	switch d := data.(type) {
	case *PaymentEventData:
		logger.Infof(ctx, "Publishing payment event: %s, order: %s, status: %s",
			t, d.OrderID, d.Status)
	case *SubscriptionEventData:
		logger.Infof(ctx, "Publishing subscription event: %s, subscription: %s, status: %s",
			t, d.SubscriptionID, d.Status)
	case *ChannelEventData:
		logger.Infof(ctx, "Publishing channel event: %s, channel: %s, provider: %s",
			t, d.ChannelID, d.Provider)
	case *ProductEventData:
		logger.Infof(ctx, "Publishing product event: %s, product: %s, status: %s",
			t, d.ProductID, d.Status)
	}

	// Publish the event through the extension manager
	if p.em != nil {
		p.em.PublishEvent(string(t), data)
	}
}

// Payment event publishing methods

func (p *publisher) PublishPaymentCreated(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentCreated, data)
}

func (p *publisher) PublishPaymentSucceeded(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentSucceeded, data)
}

func (p *publisher) PublishPaymentFailed(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentFailed, data)
}

func (p *publisher) PublishPaymentCancelled(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentCancelled, data)
}

func (p *publisher) PublishPaymentExpired(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentExpired, data)
}

func (p *publisher) PublishPaymentRefunded(ctx context.Context, data *PaymentEventData) {
	p.publish(ctx, PaymentRefunded, data)
}

// Subscription event publishing methods

func (p *publisher) PublishSubscriptionCreated(ctx context.Context, data *SubscriptionEventData) {
	p.publish(ctx, SubscriptionCreated, data)
}

func (p *publisher) PublishSubscriptionRenewed(ctx context.Context, data *SubscriptionEventData) {
	p.publish(ctx, SubscriptionRenewed, data)
}

func (p *publisher) PublishSubscriptionUpdated(ctx context.Context, data *SubscriptionEventData) {
	p.publish(ctx, SubscriptionUpdated, data)
}

func (p *publisher) PublishSubscriptionCancelled(ctx context.Context, data *SubscriptionEventData) {
	p.publish(ctx, SubscriptionCancelled, data)
}

func (p *publisher) PublishSubscriptionExpired(ctx context.Context, data *SubscriptionEventData) {
	p.publish(ctx, SubscriptionExpired, data)
}

// Product event publishing methods

func (p *publisher) PublishProductCreated(ctx context.Context, data *ProductEventData) {
	p.publish(ctx, ProductCreated, data)
}

func (p *publisher) PublishProductUpdated(ctx context.Context, data *ProductEventData) {
	p.publish(ctx, ProductUpdated, data)
}

func (p *publisher) PublishProductDeleted(ctx context.Context, data *ProductEventData) {
	p.publish(ctx, ProductDeleted, data)
}

// Channel event publishing methods

func (p *publisher) PublishChannelCreated(ctx context.Context, data *ChannelEventData) {
	p.publish(ctx, ChannelCreated, data)
}

func (p *publisher) PublishChannelUpdated(ctx context.Context, data *ChannelEventData) {
	p.publish(ctx, ChannelUpdated, data)
}

func (p *publisher) PublishChannelDeleted(ctx context.Context, data *ChannelEventData) {
	p.publish(ctx, ChannelDeleted, data)
}

func (p *publisher) PublishChannelActivated(ctx context.Context, data *ChannelEventData) {
	p.publish(ctx, ChannelActivated, data)
}

func (p *publisher) PublishChannelDisabled(ctx context.Context, data *ChannelEventData) {
	p.publish(ctx, ChannelDisabled, data)
}
