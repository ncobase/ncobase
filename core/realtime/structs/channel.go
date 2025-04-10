package structs

import (
	"fmt"
	"github.com/ncobase/ncore/pkg/types"
)

// FindChannel defines parameters for finding a channel
type FindChannel struct {
	ID string `json:"id"`
}

// ChannelBody defines the body of a channel
type ChannelBody struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`   // public/private/direct
	Status      int        `json:"status"` // 0: disabled, 1: enabled
	Extras      types.JSON `json:"extras,omitempty"`
}

// CreateChannel defines parameters for creating a channel
type CreateChannel struct {
	Channel ChannelBody `json:"channel"`
}

// UpdateChannel defines parameters for updating a channel
type UpdateChannel struct {
	ID      string      `json:"id"`
	Channel ChannelBody `json:"channel"`
}

// ReadChannel defines the response structure for channels
type ReadChannel struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Status      int        `json:"status"`
	Extras      types.JSON `json:"extras,omitempty"`
	CreatedAt   int64      `json:"created_at"`
	UpdatedAt   int64      `json:"updated_at"`
}

// GetCursorValue returns the cursor value
func (r *ReadChannel) GetCursorValue() string {
	return fmt.Sprintf("%s", r.ID)
}

// ListChannelParams defines parameters for listing channels
type ListChannelParams struct {
	Type      string `json:"type,omitempty"`
	Status    *int   `json:"status,omitempty"`
	Cursor    string `json:"cursor,omitempty"`
	Limit     int    `json:"limit,omitempty"`
	Direction string `json:"direction,omitempty"`
}

// SubscriptionBody defines the body of a subscription
type SubscriptionBody struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Status    int    `json:"status"` // 0: disabled, 1: enabled
}

// CreateSubscription defines parameters for creating a subscription
type CreateSubscription struct {
	Subscription SubscriptionBody `json:"subscription"`
}

// ReadSubscription defines the response structure for subscriptions
type ReadSubscription struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
