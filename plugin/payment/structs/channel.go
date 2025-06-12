package structs

import "fmt"

// ChannelStatus represents the status of a payment channel
type ChannelStatus string

// Channel statuses
const (
	ChannelStatusActive   ChannelStatus = "active"
	ChannelStatusDisabled ChannelStatus = "disabled"
	ChannelStatusTesting  ChannelStatus = "testing"
)

// Channel represents a payment channel configuration
type Channel struct {
	ID            string          `json:"id,omitempty"`
	Name          string          `json:"name"`
	Provider      PaymentProvider `json:"provider"`
	Status        ChannelStatus   `json:"status"`
	IsDefault     bool            `json:"is_default"`
	SupportedType []PaymentType   `json:"supported_types"`
	Config        ProviderConfig  `json:"config"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
	SpaceID       string          `json:"space_id,omitempty"`
	CreatedAt     int64           `json:"created_at,omitempty"`
	UpdatedAt     int64           `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value for pagination
func (c *Channel) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", c.ID, c.CreatedAt)
}

// CreateChannelInput represents input for creating a payment channel
type CreateChannelInput struct {
	Name          string          `json:"name" binding:"required"`
	Provider      PaymentProvider `json:"provider" binding:"required"`
	Status        ChannelStatus   `json:"status" binding:"required"`
	IsDefault     bool            `json:"is_default"`
	SupportedType []PaymentType   `json:"supported_types" binding:"required"`
	Config        ProviderConfig  `json:"config" binding:"required"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
	SpaceID       string          `json:"space_id,omitempty"`
}

// UpdateChannelInput represents input for updating a payment channel
type UpdateChannelInput struct {
	ID            string          `json:"id,omitempty"`
	Name          string          `json:"name,omitempty"`
	Provider      PaymentProvider `json:"provider,omitempty"`
	Status        ChannelStatus   `json:"status,omitempty"`
	IsDefault     *bool           `json:"is_default,omitempty"`
	SupportedType []PaymentType   `json:"supported_types,omitempty"`
	Config        ProviderConfig  `json:"config,omitempty"`
	Metadata      map[string]any  `json:"metadata,omitempty"`
	SpaceID       string          `json:"space_id,omitempty"`
}

// ChannelQuery represents query parameters for listing channels
type ChannelQuery struct {
	Provider PaymentProvider `form:"provider" json:"provider,omitempty"`
	Status   ChannelStatus   `form:"status" json:"status,omitempty"`
	SpaceID  string          `form:"space_id" json:"space_id,omitempty"`
	PaginationQuery
}
