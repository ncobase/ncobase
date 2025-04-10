package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/paging"
	"ncobase/core/realtime/data"
	"ncobase/core/realtime/data/ent"
	"ncobase/core/realtime/data/repository"
	"ncobase/core/realtime/structs"
)

type ChannelService interface {
	Create(ctx context.Context, body *structs.CreateChannel) (*structs.ReadChannel, error)
	Get(ctx context.Context, params *structs.FindChannel) (*structs.ReadChannel, error)
	Update(ctx context.Context, body *structs.UpdateChannel) (*structs.ReadChannel, error)
	Delete(ctx context.Context, params *structs.FindChannel) error
	List(ctx context.Context, params *structs.ListChannelParams) (paging.Result[*structs.ReadChannel], error)

	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error

	Subscribe(ctx context.Context, body *structs.CreateSubscription) (*structs.ReadSubscription, error)
	Unsubscribe(ctx context.Context, userID string, channelID string) error
	GetSubscribers(ctx context.Context, channelID string) ([]string, error)
	GetUserChannels(ctx context.Context, userID string) ([]*structs.ReadChannel, error)
	IsSubscribed(ctx context.Context, userID string, channelID string) (bool, error)
}

type channelService struct {
	data        *data.Data
	channelRepo repository.ChannelRepositoryInterface
	subRepo     repository.SubscriptionRepositoryInterface
	ws          WebSocketService
}

func NewChannelService(
	d *data.Data,
	ws WebSocketService,
) ChannelService {
	return &channelService{
		data:        d,
		channelRepo: repository.NewChannelRepository(d),
		subRepo:     repository.NewSubscriptionRepository(d),
		ws:          ws,
	}
}

// Create creates a new channel
func (s *channelService) Create(ctx context.Context, body *structs.CreateChannel) (*structs.ReadChannel, error) {
	ch := body.Channel
	if ch.Name == "" {
		return nil, errors.New("channel name is required")
	}

	// Check name is existed
	exists, err := s.channelRepo.FindByName(ctx, ch.Name)
	if err == nil && exists != nil {
		return nil, fmt.Errorf("channel with name %s already exists", ch.Name)
	}

	channel, err := s.channelRepo.Create(ctx, s.data.EC.Channel.Create().
		SetName(ch.Name).
		SetDescription(ch.Description).
		SetType(ch.Type).
		SetStatus(ch.Status).
		SetExtras(ch.Extras),
	)

	if err != nil {
		logger.Errorf(ctx, "Failed to create channel: %v", err)
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	result := s.serializeChannel(channel)

	// Send created event
	s.broadcastChannelEvent("channel.created", result)

	return result, nil
}

// Get retrieves a channel
func (s *channelService) Get(ctx context.Context, params *structs.FindChannel) (*structs.ReadChannel, error) {
	channel, err := s.channelRepo.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return s.serializeChannel(channel), nil
}

// Update updates a channel
func (s *channelService) Update(ctx context.Context, body *structs.UpdateChannel) (*structs.ReadChannel, error) {
	ch := body.Channel

	// Check name is existed
	if ch.Name != "" {
		existing, err := s.channelRepo.FindByName(ctx, ch.Name)
		if err == nil && existing != nil && existing.ID != body.ID {
			return nil, fmt.Errorf("channel name %s is already taken", ch.Name)
		}
	}

	update := s.data.EC.Channel.UpdateOneID(body.ID).
		SetName(ch.Name).
		SetDescription(ch.Description).
		SetType(ch.Type).
		SetStatus(ch.Status).
		SetExtras(ch.Extras)

	channel, err := s.channelRepo.Update(ctx, body.ID, update)
	if err != nil {
		return nil, err
	}

	result := s.serializeChannel(channel)

	// Send updated event
	s.broadcastChannelEvent("channel.updated", result)

	return result, nil
}

// Delete deletes a channel
func (s *channelService) Delete(ctx context.Context, params *structs.FindChannel) error {
	// Remove channel and related subscriptions
	err := s.channelRepo.Delete(ctx, params.ID)
	if err != nil {
		return err
	}

	// Send deleted event
	s.broadcastChannelEvent("channel.deleted", &structs.ReadChannel{ID: params.ID})

	return nil
}

// List lists channels
func (s *channelService) List(ctx context.Context, params *structs.ListChannelParams) (paging.Result[*structs.ReadChannel], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadChannel, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.channelRepo.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing permissions: %v", err)
			return nil, 0, err
		}

		total := s.channelRepo.CountX(ctx, params)

		return s.serializeChannels(rows), total, nil
	})
}

// Enable enables a channel
func (s *channelService) Enable(ctx context.Context, id string) error {
	err := s.channelRepo.UpdateStatus(ctx, id, 1)
	if err != nil {
		return err
	}

	// Send status change event
	s.broadcastChannelEvent("channel.enabled", &structs.ReadChannel{ID: id})
	return nil
}

// Disable disables a channel
func (s *channelService) Disable(ctx context.Context, id string) error {
	err := s.channelRepo.UpdateStatus(ctx, id, 0)
	if err != nil {
		return err
	}

	// Send status change event
	s.broadcastChannelEvent("channel.disabled", &structs.ReadChannel{ID: id})
	return nil
}

// Subscribe subscribes a user to a channel
func (s *channelService) Subscribe(ctx context.Context, body *structs.CreateSubscription) (*structs.ReadSubscription, error) {
	sub := body.Subscription

	// Check channel is existd and enabled
	channel, err := s.channelRepo.Get(ctx, sub.ChannelID)
	if err != nil {
		return nil, err
	}
	if channel.Status != 1 {
		return nil, errors.New("channel is disabled")
	}

	subscription, err := s.subRepo.Create(ctx, s.data.EC.Subscription.Create().
		SetUserID(sub.UserID).
		SetChannelID(sub.ChannelID).
		SetStatus(sub.Status),
	)

	if err != nil {
		return nil, err
	}

	result := s.serializeSubscription(subscription)

	// Create WebSocket subscription
	err = s.ws.SubscribeToChannel(sub.UserID, sub.ChannelID)
	if err != nil {
		return nil, err
	}

	// Send subscribed event
	s.broadcastChannelEvent("channel.subscribed", map[string]any{
		"channel_id": sub.ChannelID,
		"user_id":    sub.UserID,
	})

	return result, nil
}

// Unsubscribe unsubscribes a user from a channel
func (s *channelService) Unsubscribe(ctx context.Context, userID string, channelID string) error {
	err := s.subRepo.DeleteByUserAndChannel(ctx, userID, channelID)
	if err != nil {
		return err
	}

	// Remove WebSocket subscription
	err = s.ws.UnsubscribeFromChannel(userID, channelID)
	if err != nil {
		return err
	}

	// Send unsubscribed event
	s.broadcastChannelEvent("channel.unsubscribed", map[string]any{
		"channel_id": channelID,
		"user_id":    userID,
	})

	return nil
}

// GetSubscribers gets all subscribers of a channel
func (s *channelService) GetSubscribers(ctx context.Context, channelID string) ([]string, error) {
	subs, err := s.subRepo.GetChannelSubscribers(ctx, channelID)
	if err != nil {
		return nil, err
	}

	subscribers := make([]string, len(subs))
	for i, sub := range subs {
		subscribers[i] = sub.UserID
	}

	return subscribers, nil
}

// GetUserChannels gets all channels a user has subscribed to
func (s *channelService) GetUserChannels(ctx context.Context, userID string) ([]*structs.ReadChannel, error) {
	subs, err := s.subRepo.GetUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var channels []*structs.ReadChannel
	for _, sub := range subs {
		channel, err := s.channelRepo.Get(ctx, sub.ChannelID)
		if err != nil {
			continue
		}
		channels = append(channels, s.serializeChannel(channel))
	}

	return channels, nil
}

// IsSubscribed checks if a user is subscribed to a channel
func (s *channelService) IsSubscribed(ctx context.Context, userID string, channelID string) (bool, error) {
	sub, err := s.subRepo.FindByUserAndChannel(ctx, userID, channelID)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return sub.Status == 1, nil
}

// Serialization helpers
func (s *channelService) serializeChannel(ch *ent.Channel) *structs.ReadChannel {
	return &structs.ReadChannel{
		ID:          ch.ID,
		Name:        ch.Name,
		Description: ch.Description,
		Type:        ch.Type,
		Status:      ch.Status,
		Extras:      ch.Extras,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
}

func (s *channelService) serializeChannels(channels []*ent.Channel) []*structs.ReadChannel {
	result := make([]*structs.ReadChannel, len(channels))
	for i, ch := range channels {
		result[i] = s.serializeChannel(ch)
	}
	return result
}

func (s *channelService) serializeSubscription(sub *ent.Subscription) *structs.ReadSubscription {
	return &structs.ReadSubscription{
		ID:        sub.ID,
		UserID:    sub.UserID,
		ChannelID: sub.ChannelID,
		Status:    sub.Status,
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}
}

// broadcastChannelEvent broadcasts a channel-related event
func (s *channelService) broadcastChannelEvent(eventType string, data any) {
	message := &WebSocketMessage{
		Type: eventType,
		Data: data,
	}
	if err := s.ws.BroadcastToAll(message); err != nil {
		logger.Errorf(context.Background(), "Failed to broadcast channel event: %v", err)
	}
}
