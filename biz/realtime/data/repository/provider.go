package repository

import "ncobase/biz/realtime/data"

// Repository represents the realtime repository provider
type Repository struct {
	Notification NotificationRepositoryInterface
	Channel      ChannelRepositoryInterface
	Event        EventRepositoryInterface
	Subscription SubscriptionRepositoryInterface
}

// New creates a new repository provider
func New(d *data.Data) *Repository {
	return &Repository{
		Notification: NewNotificationRepository(d),
		Channel:      NewChannelRepository(d),
		Event:        NewEventRepository(d),
		Subscription: NewSubscriptionRepository(d),
	}
}
