package handler

import (
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/service"

	ext "github.com/ncobase/ncore/extension/types"
)

// Handler represents the payment handler provider
type Handler struct {
	Channel      ChannelHandlerInterface
	Order        OrderHandlerInterface
	Product      ProductHandlerInterface
	Subscription SubscriptionHandlerInterface
	Log          LogHandlerInterface
	Webhook      WebhookHandlerInterface
	Utility      UtilityHandlerInterface
	Event        EventHandlerInterface
}

// New creates a new handler provider
func New(em ext.ManagerInterface, s *service.Service) *Handler {
	// Register event handlers if event manager exists
	handlerProvider := NewEventProvider(em, s)
	registrar := event.NewRegistrar(em)
	registrar.RegisterHandlers(handlerProvider)
	return &Handler{
		Channel:      NewChannelHandler(s.Channel),
		Order:        NewOrderHandler(s.Order),
		Product:      NewProductHandler(s.Product),
		Subscription: NewSubscriptionHandler(s.Subscription),
		Log:          NewLogHandler(s.Log),
		Webhook:      NewWebhookHandler(s.Order),
		Utility:      NewUtilityHandler(s.Provider),
		Event:        handlerProvider,
	}
}
