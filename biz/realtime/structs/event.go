package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
)

// FindEvent defines parameters for finding an event
type FindEvent struct {
	ID string `json:"id"`
}

// EventBody defines the body of an event
type EventBody struct {
	Type     string     `json:"type"`
	Source   string     `json:"source"`
	Payload  types.JSON `json:"payload,omitempty"`
	Priority string     `json:"priority,omitempty"`
}

// CreateEvent defines parameters for creating an event
type CreateEvent struct {
	Event EventBody `json:"event"`
}

// ReadEvent defines the response structure for events
type ReadEvent struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Source       string     `json:"source"`
	Payload      types.JSON `json:"payload,omitempty"`
	Status       string     `json:"status"`
	Priority     string     `json:"priority,omitempty"`
	CreatedAt    int64      `json:"created_at"`
	ProcessedAt  *int64     `json:"processed_at,omitempty"`
	RetryCount   int        `json:"retry_count,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

// GetCursorValue returns the cursor value
func (r *ReadEvent) GetCursorValue() string {
	return fmt.Sprintf("%s", r.ID)
}

// ListEventParams defines parameters for listing events
type ListEventParams struct {
	Type      string  `json:"type,omitempty"`
	Source    string  `json:"source,omitempty"`
	Status    string  `json:"status,omitempty"`
	Cursor    string  `json:"cursor,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Direction string  `json:"direction,omitempty"`
	TimeRange []int64 `json:"time_range,omitempty"`
}

// SearchQuery defines search query parameters
type SearchQuery struct {
	Query        map[string]any   `json:"query,omitempty"`
	Filters      map[string]any   `json:"filters,omitempty"`
	Aggregations map[string]any   `json:"aggregations,omitempty"`
	TimeRange    *TimeRange       `json:"time_range,omitempty"`
	Size         int              `json:"size,omitempty"`
	From         int              `json:"from,omitempty"`
	Sort         []map[string]any `json:"sort,omitempty"`
}

// TimeRange defines time range for queries
type TimeRange struct {
	Start string `json:"start"` // ISO 8601 format
	End   string `json:"end"`   // ISO 8601 format
}

// SearchResult defines search result structure
type SearchResult struct {
	Total        int64          `json:"total"`
	Events       []*ReadEvent   `json:"events"`
	Aggregations map[string]any `json:"aggregations,omitempty"`
	ScrollID     string         `json:"scroll_id,omitempty"`
}

// RealtimeStats defines real-time statistics structure
type RealtimeStats struct {
	Timestamp string                    `json:"timestamp"`
	Interval  string                    `json:"interval,omitempty"`
	Metrics   map[string]any            `json:"metrics"`
	Breakdown map[string]map[string]any `json:"breakdown,omitempty"`
}

// StatsParams defines parameters for statistics queries
type StatsParams struct {
	Interval  string     `json:"interval,omitempty"`
	Type      string     `json:"type,omitempty"`
	TimeRange *TimeRange `json:"time_range,omitempty"`
}

// RetryParams defines parameters for event retry
type RetryParams struct {
	Reason       string        `json:"reason,omitempty"`
	Priority     string        `json:"priority,omitempty"`
	RetryOptions *RetryOptions `json:"retry_options,omitempty"`
}

// RetryOptions defines retry configuration
type RetryOptions struct {
	MaxAttempts  int `json:"max_attempts,omitempty"`
	DelaySeconds int `json:"delay_seconds,omitempty"`
}

// RetryResult defines retry operation result
type RetryResult struct {
	RetryID     string `json:"retry_id"`
	ScheduledAt string `json:"scheduled_at"`
}
