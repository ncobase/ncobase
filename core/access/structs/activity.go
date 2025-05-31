package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
)

// ActivityDocument represents activity stored in search engine/database
type ActivityDocument struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      string      `json:"type"`
	Details   string      `json:"details"`
	Metadata  *types.JSON `json:"metadata,omitempty"`
	CreatedAt int64       `json:"created_at"`
	UpdatedAt int64       `json:"updated_at"`
}

func (a *ActivityDocument) GetCursorValue() string {
	return fmt.Sprintf("%s-%d", a.ID, a.CreatedAt)
}

// Activity represents a user activity log entry (API response)
type Activity struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Details   string      `json:"details,omitempty"`
	Metadata  *types.JSON `json:"metadata,omitempty"`
	CreatedAt int64       `json:"created_at,omitempty"`
	UpdatedAt int64       `json:"updated_at,omitempty"`
}

func (a *Activity) GetCursorValue() string {
	return fmt.Sprintf("%s-%d", a.ID, a.Timestamp)
}

// CreateActivityRequest represents a request to create an activity
type CreateActivityRequest struct {
	Type     string      `json:"type" validate:"required"`
	Details  string      `json:"details" validate:"required"`
	Metadata *types.JSON `json:"metadata,omitempty"`
}

// ListActivityParams represents parameters for listing activities
type ListActivityParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	UserID    string `form:"user_id,omitempty" json:"user_id,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	FromDate  int64  `form:"from_date,omitempty" json:"from_date,omitempty"`
	ToDate    int64  `form:"to_date,omitempty" json:"to_date,omitempty"`
	Offset    int    `form:"offset,omitempty" json:"offset,omitempty"`
	SortBy    string `form:"sort_by,omitempty" json:"sort_by,omitempty"`
	Order     string `form:"order,omitempty" json:"order,omitempty"`
}

// SearchActivityParams represents parameters for searching activities
type SearchActivityParams struct {
	Query    string `form:"q,omitempty" json:"q,omitempty"`
	UserID   string `form:"user_id,omitempty" json:"user_id,omitempty"`
	Type     string `form:"type,omitempty" json:"type,omitempty"`
	FromDate int64  `form:"from_date,omitempty" json:"from_date,omitempty"`
	ToDate   int64  `form:"to_date,omitempty" json:"to_date,omitempty"`
	From     int    `form:"from,omitempty" json:"from,omitempty"`
	Size     int    `form:"size,omitempty" json:"size,omitempty"`
}
