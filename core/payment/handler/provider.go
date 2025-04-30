package handler

import (
	"ncobase/core/payment/service"
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
}

// New creates a new handler provider
func New(s *service.Service) *Handler {
	return &Handler{
		Channel:      NewChannelHandler(s.Channel),
		Order:        NewOrderHandler(s.Order),
		Product:      NewProductHandler(s.Product),
		Subscription: NewSubscriptionHandler(s.Subscription),
		Log:          NewLogHandler(s.Log),
		Webhook:      NewWebhookHandler(s.Order),
		Utility:      NewUtilityHandler(s.Provider),
	}
}
