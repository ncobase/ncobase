package event

import (
	ext "github.com/ncobase/ncore/extension/types"
)

// Registrar handles registration of event handlers
type Registrar struct {
	manager ext.ManagerInterface
}

// NewRegistrar creates a new event registrar
func NewRegistrar(manager ext.ManagerInterface) *Registrar {
	return &Registrar{
		manager: manager,
	}
}

// RegisterHandlers registers event handlers with the manager
func (r *Registrar) RegisterHandlers(provider HandlerProvider) {
	handlers := provider.GetHandlers()

	// Define a mapping between handler keys and event types
	eventMapping := map[string]string{
		// Payment events
		"payment_created":   PaymentCreated,
		"payment_succeeded": PaymentSucceeded,
		"payment_failed":    PaymentFailed,
		"payment_cancelled": PaymentCancelled,
		"payment_expired":   PaymentExpired,
		"payment_refunded":  PaymentRefunded,

		// Subscription events
		"subscription_created":   SubscriptionCreated,
		"subscription_renewed":   SubscriptionRenewed,
		"subscription_updated":   SubscriptionUpdated,
		"subscription_cancelled": SubscriptionCancelled,
		"subscription_expired":   SubscriptionExpired,

		// Product events
		"product_created": ProductCreated,
		"product_updated": ProductUpdated,
		"product_deleted": ProductDeleted,

		// Channel events
		"channel_created":   ChannelCreated,
		"channel_updated":   ChannelUpdated,
		"channel_deleted":   ChannelDeleted,
		"channel_activated": ChannelActivated,
		"channel_disabled":  ChannelDisabled,
	}

	// Register handlers using the mapping
	for handlerKey, handler := range handlers {
		if eventType, exists := eventMapping[handlerKey]; exists {
			r.manager.SubscribeEvent(string(eventType), handler)
		}
	}
}
