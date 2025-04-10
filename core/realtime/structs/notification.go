package structs

import (
	"fmt"
	"github.com/ncobase/ncore/pkg/types"
)

// FindNotification defines the parameters for finding a notification
type FindNotification struct {
	ID string `json:"id"`
}

// NotificationBody defines the body of a notification
type NotificationBody struct {
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Type      string       `json:"type"`
	UserID    string       `json:"user_id"`
	Status    int          `json:"status"`
	ChannelID string       `json:"channel_id,omitempty"`
	Links     []types.JSON `json:"links,omitempty"`
}

// CreateNotification defines the body of a notification
type CreateNotification struct {
	Notification NotificationBody `json:"notification"`
}

// UpdateNotification defines the body of a notification
type UpdateNotification struct {
	ID           string           `json:"id"`
	Notification NotificationBody `json:"notification"`
}

// ReadNotification defines the body of a notification
type ReadNotification struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Content   string       `json:"content"`
	Type      string       `json:"type"`
	UserID    string       `json:"user_id"`
	Status    int          `json:"status"`
	ChannelID string       `json:"channel_id,omitempty"`
	Links     []types.JSON `json:"links,omitempty"`
	CreatedAt int64        `json:"created_at"`
	UpdatedAt int64        `json:"updated_at"`
}

// GetCursorValue returns the cursor value
func (r *ReadNotification) GetCursorValue() string {
	return fmt.Sprintf("%s", r.ID)
}

// ListNotificationParams defines parameters for listing notifications
type ListNotificationParams struct {
	UserID    string `json:"user_id"`
	Status    *int   `json:"status,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Direction string `json:"direction,omitempty"`
}
