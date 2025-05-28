package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Manager handles publishing events for the payment module
type Manager struct {
	em ext.ManagerInterface
}

// NewManager creates a new event manager
func NewManager(manager ext.ManagerInterface) *Manager {
	return &Manager{
		em: manager,
	}
}

// PublishPaymentEvent publishes a payment-related event
func (m *Manager) PublishPaymentEvent(ctx context.Context, eventType string, data *PaymentEventData) {
	// Log the event
	logger.Infof(ctx, "Publishing payment event: %s, order: %s, status: %s",
		eventType, data.OrderID, data.Status)

	// Publish the event through the extension manager
	if m.em != nil {
		m.em.PublishEvent(eventType, data)
	}
}

// PublishSubscriptionEvent publishes a subscription-related event
func (m *Manager) PublishSubscriptionEvent(ctx context.Context, eventType string, data *SubscriptionEventData) {
	// Log the event
	logger.Infof(ctx, "Publishing subscription event: %s, subscription: %s, status: %s",
		eventType, data.SubscriptionID, data.Status)

	// Publish the event through the extension manager
	if m.em != nil {
		m.em.PublishEvent(eventType, data)
	}
}

// PublishChannelEvent publishes a payment channel-related event
func (m *Manager) PublishChannelEvent(ctx context.Context, eventType string, data *ChannelEventData) {
	// Log the event
	logger.Infof(ctx, "Publishing channel event: %s, channel: %s, provider: %s",
		eventType, data.ChannelID, data.Provider)

	// Publish the event through the extension manager
	if m.em != nil {
		m.em.PublishEvent(eventType, data)
	}
}

// PublishProductEvent publishes a product-related event
func (m *Manager) PublishProductEvent(ctx context.Context, eventType string, data *ProductEventData) {
	// Log the event
	logger.Infof(ctx, "Publishing product event: %s, product: %s, status: %s",
		eventType, data.ProductID, data.Status)

	// Publish the event through the extension manager
	if m.em != nil {
		m.em.PublishEvent(eventType, data)
	}
}
