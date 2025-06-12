package structs

import (
	"fmt"
	"time"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

// Subscription statuses
const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusTrialing  SubscriptionStatus = "trialing"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusPastDue   SubscriptionStatus = "past_due"
)

// Subscription represents a subscription to a product
type Subscription struct {
	ID                 string             `json:"id,omitempty"`
	Status             SubscriptionStatus `json:"status"`
	UserID             string             `json:"user_id"`
	SpaceID            string             `json:"space_id,omitempty"`
	ProductID          string             `json:"product_id"`
	ChannelID          string             `json:"channel_id"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelAt           *time.Time         `json:"cancel_at,omitempty"`
	CancelledAt        *time.Time         `json:"cancelled_at,omitempty"`
	TrialStart         *time.Time         `json:"trial_start,omitempty"`
	TrialEnd           *time.Time         `json:"trial_end,omitempty"`
	ProviderRef        string             `json:"provider_ref,omitempty"`
	Metadata           map[string]any     `json:"metadata,omitempty"`
	CreatedAt          int64              `json:"created_at,omitempty"`
	UpdatedAt          int64              `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (s *Subscription) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", s.ID, s.CreatedAt)
}

// CreateSubscriptionInput represents input for creating a subscription
type CreateSubscriptionInput struct {
	UserID    string         `json:"user_id" binding:"required"`
	SpaceID   string         `json:"space_id,omitempty"`
	ProductID string         `json:"product_id" binding:"required"`
	ChannelID string         `json:"channel_id" binding:"required"`
	TrialDays int            `json:"trial_days,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// UpdateSubscriptionInput represents the input for updating a subscription
type UpdateSubscriptionInput struct {
	ID string `json:"id,omitempty"`
}

// CancelSubscriptionInput represents input for cancelling a subscription
type CancelSubscriptionInput struct {
	Immediate bool   `json:"immediate"`
	Reason    string `json:"reason"`
}

// SubscriptionQuery represents query parameters for listing subscriptions
type SubscriptionQuery struct {
	Status    SubscriptionStatus `form:"status" json:"status,omitempty"`
	UserID    string             `form:"user_id" json:"user_id,omitempty"`
	SpaceID   string             `form:"space_id" json:"space_id,omitempty"`
	ProductID string             `form:"product_id" json:"product_id,omitempty"`
	ChannelID string             `form:"channel_id" json:"channel_id,omitempty"`
	Active    *bool              `form:"active" json:"active,omitempty"`
	PaginationQuery
}

// SubscriptionSummary represents a summary of subscription stats
type SubscriptionSummary struct {
	TotalCount     int64   `json:"total_count"`
	ActiveCount    int64   `json:"active_count"`
	TrialingCount  int64   `json:"trialing_count"`
	CancelledCount int64   `json:"cancelled_count"`
	ExpiredCount   int64   `json:"expired_count"`
	PastDueCount   int64   `json:"past_due_count"`
	MRR            float64 `json:"mrr"` // Monthly Recurring Revenue
	ARR            float64 `json:"arr"` // Annual Recurring Revenue
}
