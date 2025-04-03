package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/realtime/data"
	"ncobase/core/realtime/data/ent"
	"ncobase/core/realtime/data/repository"
	"ncobase/core/realtime/structs"
	"ncobase/ncore/ecode"
	"ncobase/ncore/logger"
	"ncobase/ncore/paging"
	"ncobase/ncore/types"
)

type NotificationService interface {
	Create(ctx context.Context, body *structs.CreateNotification) (*structs.ReadNotification, error)
	Get(ctx context.Context, params *structs.FindNotification) (*structs.ReadNotification, error)
	Update(ctx context.Context, body *structs.UpdateNotification) (*structs.ReadNotification, error)
	Delete(ctx context.Context, params *structs.FindNotification) error
	List(ctx context.Context, params *structs.ListNotificationParams) (paging.Result[*structs.ReadNotification], error)
	MarkAsRead(ctx context.Context, params *structs.FindNotification) error
	MarkAllAsRead(ctx context.Context, userID string) error
	MarkAsUnread(ctx context.Context, params *structs.FindNotification) error
	MarkAllAsUnread(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type notificationService struct {
	data *data.Data
	repo repository.NotificationRepositoryInterface
	ws   WebSocketService
}

func NewNotificationService(d *data.Data, ws WebSocketService) NotificationService {
	return &notificationService{
		data: d,
		repo: repository.NewNotificationRepository(d),
		ws:   ws,
	}
}

// Create creates a new notification
func (s *notificationService) Create(ctx context.Context, body *structs.CreateNotification) (*structs.ReadNotification, error) {
	n := body.Notification
	if n.Title == "" || n.Content == "" || n.UserID == "" {
		return nil, errors.New("title, content and user_id are required")
	}

	// create notification
	notification, err := s.repo.Create(ctx, s.data.EC.Notification.Create().
		SetTitle(n.Title).
		SetContent(n.Content).
		SetType(n.Type).
		SetUserID(n.UserID).
		SetStatus(n.Status).
		SetNillableChannelID(&n.ChannelID).
		SetLinks(n.Links),
	)

	if err != nil {
		logger.Errorf(ctx, "Failed to create notification: %v", err)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	result := s.serializeNotification(notification)

	// send realtime notification
	s.sendRealtimeNotification(result)

	return result, nil
}

// Get retrieves a notification
func (s *notificationService) Get(ctx context.Context, params *structs.FindNotification) (*structs.ReadNotification, error) {
	notification, err := s.repo.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return s.serializeNotification(notification), nil
}

// Update updates a notification
func (s *notificationService) Update(ctx context.Context, body *structs.UpdateNotification) (*structs.ReadNotification, error) {
	// get existing notification
	existing, err := s.repo.Get(ctx, body.ID)
	if err != nil {
		return nil, err
	}

	n := body.Notification
	update := s.data.EC.Notification.UpdateOneID(body.ID).
		SetTitle(n.Title).
		SetContent(n.Content).
		SetType(n.Type).
		SetStatus(n.Status).
		SetNillableChannelID(&n.ChannelID).
		SetLinks(n.Links)

	notification, err := s.repo.Update(ctx, body.ID, update)
	if err != nil {
		return nil, err
	}

	result := s.serializeNotification(notification)

	// send status change notification
	if existing.Status != notification.Status {
		s.sendStatusChangeNotification(result)
	}

	return result, nil
}

// Delete deletes a notification
func (s *notificationService) Delete(ctx context.Context, params *structs.FindNotification) error {
	return s.repo.Delete(ctx, params.ID)
}

// List lists notifications
func (s *notificationService) List(ctx context.Context, params *structs.ListNotificationParams) (paging.Result[*structs.ReadNotification], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadNotification, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.repo.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing permissions: %v", err)
			return nil, 0, err
		}

		total := s.repo.CountX(ctx, params)

		return s.serializeNotifications(rows), total, nil
	})
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, params *structs.FindNotification) error {
	return s.repo.UpdateStatus(ctx, params.ID, 1) // 1 read
}

// MarkAllAsRead marks all notifications as read for a user
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.UpdateStatusBatch(ctx, userID, 1) // 1 read
}

// MarkAsUnread marks a notification as unread
func (s *notificationService) MarkAsUnread(ctx context.Context, params *structs.FindNotification) error {
	return s.repo.UpdateStatus(ctx, params.ID, 0) // 0 unread
}

// MarkAllAsUnread marks all notifications as unread for a user
func (s *notificationService) MarkAllAsUnread(ctx context.Context, userID string) error {
	return s.repo.UpdateStatusBatch(ctx, userID, 0) // 0 unread
}

// GetUnreadCount gets unread notification count for a user
func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.Count(ctx, &structs.ListNotificationParams{
		UserID: userID,
		Status: types.ToPointer(0), // 0 unread
	})
}

// serializeNotification converts ent.Notification to structs.ReadNotification
func (s *notificationService) serializeNotification(n *ent.Notification) *structs.ReadNotification {
	return &structs.ReadNotification{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      n.Type,
		UserID:    n.UserID,
		Status:    n.Status,
		ChannelID: n.ChannelID,
		Links:     n.Links,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// serializeNotifications converts []*ent.Notification to []*structs.ReadNotification
func (s *notificationService) serializeNotifications(notifications []*ent.Notification) []*structs.ReadNotification {
	result := make([]*structs.ReadNotification, len(notifications))
	for i, n := range notifications {
		result[i] = s.serializeNotification(n)
	}
	return result
}

// sendRealtimeNotification sends a notification through WebSocket
func (s *notificationService) sendRealtimeNotification(n *structs.ReadNotification) {
	message := &WebSocketMessage{
		Type: "notification",
		Data: n,
	}

	if n.ChannelID != "" {
		err := s.ws.BroadcastToChannel(n.ChannelID, message)
		if err != nil {
			logger.Errorf(context.Background(), "Failed to broadcast notification to channel: %v", err)
		}
	} else {
		err := s.ws.BroadcastToUser(n.UserID, message)
		if err != nil {
			logger.Errorf(context.Background(), "Failed to send notification to user: %v", err)
		}
	}
}

// sendStatusChangeNotification sends a status change notification
func (s *notificationService) sendStatusChangeNotification(n *structs.ReadNotification) {
	message := &WebSocketMessage{
		Type: "notification.status_changed",
		Data: n,
	}

	err := s.ws.BroadcastToUser(n.UserID, message)
	if err != nil {
		logger.Errorf(context.Background(), "Failed to send status change notification: %v", err)
	}
}
