package structs

import "fmt"

// LogType represents the type of payment log entry
type LogType string

// Log types
const (
	LogTypeCreate   LogType = "create"
	LogTypeUpdate   LogType = "update"
	LogTypeVerify   LogType = "verify"
	LogTypeCallback LogType = "callback"
	LogTypeNotify   LogType = "notify"
	LogTypeRefund   LogType = "refund"
	LogTypeError    LogType = "error"
)

// Log represents a payment log entry
type Log struct {
	ID           string         `json:"id,omitempty"`
	OrderID      string         `json:"order_id"`
	ChannelID    string         `json:"channel_id"`
	Type         LogType        `json:"type"`
	StatusBefore PaymentStatus  `json:"status_before,omitempty"`
	StatusAfter  PaymentStatus  `json:"status_after,omitempty"`
	RequestData  string         `json:"request_data,omitempty"`
	ResponseData string         `json:"response_data,omitempty"`
	IP           string         `json:"ip,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	UserID       string         `json:"user_id,omitempty"`
	Error        string         `json:"error,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    int64          `json:"created_at,omitempty"`
	UpdatedAt    int64          `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (l *Log) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", l.ID, l.CreatedAt)
}

// CreateLogInput represents input for creating a log entry
type CreateLogInput struct {
	OrderID      string         `json:"order_id"`
	ChannelID    string         `json:"channel_id"`
	Type         LogType        `json:"type"`
	StatusBefore PaymentStatus  `json:"status_before,omitempty"`
	StatusAfter  PaymentStatus  `json:"status_after,omitempty"`
	RequestData  string         `json:"request_data,omitempty"`
	ResponseData string         `json:"response_data,omitempty"`
	IP           string         `json:"ip,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	UserID       string         `json:"user_id,omitempty"`
	Error        string         `json:"error,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// LogQuery represents query parameters for listing logs
type LogQuery struct {
	OrderID   string  `form:"order_id" json:"order_id,omitempty"`
	ChannelID string  `form:"channel_id" json:"channel_id,omitempty"`
	Type      LogType `form:"type" json:"type,omitempty"`
	HasError  *bool   `form:"has_error" json:"has_error,omitempty"`
	StartDate int64   `form:"start_date" json:"start_date,omitempty"`
	EndDate   int64   `form:"end_date" json:"end_date,omitempty"`
	UserID    string  `form:"user_id" json:"user_id,omitempty"`
	PaginationQuery
}
