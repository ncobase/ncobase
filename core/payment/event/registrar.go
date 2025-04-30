package event

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
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

	// Payment events
	if handler, ok := handlers["payment_created"]; ok {
		r.manager.SubscribeEvent(PaymentCreated, handler)
	}

	if handler, ok := handlers["payment_succeeded"]; ok {
		r.manager.SubscribeEvent(PaymentSucceeded, handler)
	}

	if handler, ok := handlers["payment_failed"]; ok {
		r.manager.SubscribeEvent(PaymentFailed, handler)
	}

	if handler, ok := handlers["payment_cancelled"]; ok {
		r.manager.SubscribeEvent(PaymentCancelled, handler)
	}

	if handler, ok := handlers["payment_expired"]; ok {
		r.manager.SubscribeEvent(PaymentExpired, handler)
	}

	if handler, ok := handlers["payment_refunded"]; ok {
		r.manager.SubscribeEvent(PaymentRefunded, handler)
	}

	// Subscription events
	if handler, ok := handlers["subscription_created"]; ok {
		r.manager.SubscribeEvent(SubscriptionCreated, handler)
	}

	if handler, ok := handlers["subscription_renewed"]; ok {
		r.manager.SubscribeEvent(SubscriptionRenewed, handler)
	}

	if handler, ok := handlers["subscription_updated"]; ok {
		r.manager.SubscribeEvent(SubscriptionUpdated, handler)
	}

	if handler, ok := handlers["subscription_cancelled"]; ok {
		r.manager.SubscribeEvent(SubscriptionCancelled, handler)
	}

	if handler, ok := handlers["subscription_expired"]; ok {
		r.manager.SubscribeEvent(SubscriptionExpired, handler)
	}

	// Product events
	if handler, ok := handlers["product_created"]; ok {
		r.manager.SubscribeEvent(ProductCreated, handler)
	}

	if handler, ok := handlers["product_updated"]; ok {
		r.manager.SubscribeEvent(ProductUpdated, handler)
	}

	if handler, ok := handlers["product_deleted"]; ok {
		r.manager.SubscribeEvent(ProductDeleted, handler)
	}

	// Channel events
	if handler, ok := handlers["channel_created"]; ok {
		r.manager.SubscribeEvent(ChannelCreated, handler)
	}

	if handler, ok := handlers["channel_updated"]; ok {
		r.manager.SubscribeEvent(ChannelUpdated, handler)
	}

	if handler, ok := handlers["channel_deleted"]; ok {
		r.manager.SubscribeEvent(ChannelDeleted, handler)
	}

	if handler, ok := handlers["channel_activated"]; ok {
		r.manager.SubscribeEvent(ChannelActivated, handler)
	}

	if handler, ok := handlers["channel_disabled"]; ok {
		r.manager.SubscribeEvent(ChannelDisabled, handler)
	}

	logger.Info(context.Background(), "Payment event handlers registered")
}
