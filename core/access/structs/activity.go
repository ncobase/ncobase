package structs

import "github.com/ncobase/ncore/types"

// ActivityEntry represents a user activity log entry
type ActivityEntry struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Details   string      `json:"details"`
	Metadata  *types.JSON `json:"metadata,omitempty"`
}

// CreateActivityRequest represents a request to create an activity
type CreateActivityRequest struct {
	Type     string      `json:"type" validate:"required"`
	Details  string      `json:"details" validate:"required"`
	Metadata *types.JSON `json:"metadata,omitempty"`
}

// ListActivityParams represents parameters for listing activities
type ListActivityParams struct {
	UserID   string `form:"user_id,omitempty" json:"user_id,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty"`
	FromDate int64  `form:"from_date,omitempty" json:"from_date,omitempty"`
	ToDate   int64  `form:"to_date,omitempty" json:"to_date,omitempty"`
	Limit    int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset   int    `form:"offset,omitempty" json:"offset,omitempty"`
}
