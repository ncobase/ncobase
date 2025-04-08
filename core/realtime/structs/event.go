package structs

import (
	"fmt"
	"ncore/pkg/types"
)

// FindEvent defines parameters for finding an event
type FindEvent struct {
	ID string `json:"id"`
}

// EventBody defines the body of an event
type EventBody struct {
	Type      string     `json:"type"`
	ChannelID string     `json:"channel_id"`
	UserID    string     `json:"user_id,omitempty"`
	Payload   types.JSON `json:"payload,omitempty"`
}

// CreateEvent defines parameters for creating an event
type CreateEvent struct {
	Event EventBody `json:"event"`
}

// ReadEvent defines the response structure for events
type ReadEvent struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	ChannelID string     `json:"channel_id"`
	UserID    string     `json:"user_id,omitempty"`
	Payload   types.JSON `json:"payload,omitempty"`
	CreatedAt int64      `json:"created_at"`
}

// GetCursorValue returns the cursor value
func (r *ReadEvent) GetCursorValue() string {
	return fmt.Sprintf("%s", r.ID)
}

// ListEventParams defines parameters for listing events
type ListEventParams struct {
	ChannelID string  `json:"channel_id,omitempty"`
	Type      string  `json:"type,omitempty"`
	UserID    string  `json:"user_id,omitempty"`
	Cursor    string  `json:"cursor,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Direction string  `json:"direction,omitempty"`
	TimeRange []int64 `json:"time_range,omitempty"`
}
