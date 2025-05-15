package repository

import (
	"ncobase/payment/data"
)

// Repository represents the payment repository collection
type Repository struct {
	Channel      ChannelRepositoryInterface
	Order        OrderRepositoryInterface
	Log          LogRepositoryInterface
	Product      ProductRepositoryInterface
	Subscription SubscriptionRepositoryInterface
}

// New creates a new repository provider
func New(d *data.Data) *Repository {
	return &Repository{
		Channel:      NewChannelRepository(d),
		Order:        NewOrderRepository(d),
		Log:          NewLogRepository(d),
		Product:      NewProductRepository(d),
		Subscription: NewSubscriptionRepository(d),
	}
}
