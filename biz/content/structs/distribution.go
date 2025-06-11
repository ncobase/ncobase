package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

const (
	// DistributionStatusDraft DistributionStatus represents the status of a distribution
	DistributionStatusDraft     int = 0
	DistributionStatusScheduled int = 1
	DistributionStatusPublished int = 2
	DistributionStatusFailed    int = 3
	DistributionStatusCancelled int = 4
)

// DistributionBody represents common fields for creating and updating distribution
type DistributionBody struct {
	TopicID      string      `json:"topic_id,omitempty"`
	ChannelID    string      `json:"channel_id,omitempty"`
	Status       int         `json:"status,omitempty"` // 0: draft, 1: scheduled, 2: published, 3: failed, 4: cancelled
	ScheduledAt  *int64      `json:"scheduled_at,omitempty"`
	PublishedAt  *int64      `json:"published_at,omitempty"`
	MetaData     *types.JSON `json:"meta_data,omitempty"`     // Platform-specific data
	ExternalID   string      `json:"external_id,omitempty"`   // ID on the external platform
	ExternalURL  string      `json:"external_url,omitempty"`  // URL on the external platform
	CustomData   *types.JSON `json:"custom_data,omitempty"`   // Custom data for the distribution
	ErrorDetails string      `json:"error_details,omitempty"` // Error details if distribution failed
	SpaceID      string      `json:"space_id,omitempty"`
	CreatedBy    *string     `json:"created_by,omitempty"`
	UpdatedBy    *string     `json:"updated_by,omitempty"`
}

// CreateDistributionBody for creating distribution
type CreateDistributionBody struct {
	DistributionBody
}

// UpdateDistributionBody for updating distribution
type UpdateDistributionBody struct {
	ID string `json:"id"`
	DistributionBody
}

// ReadDistribution represents output schema for retrieving distribution
type ReadDistribution struct {
	ID           string       `json:"id"`
	TopicID      string       `json:"topic_id"`
	ChannelID    string       `json:"channel_id"`
	Status       int          `json:"status"`
	ScheduledAt  *int64       `json:"scheduled_at,omitempty"`
	PublishedAt  *int64       `json:"published_at,omitempty"`
	MetaData     *types.JSON  `json:"meta_data,omitempty"`
	ExternalID   string       `json:"external_id"`
	ExternalURL  string       `json:"external_url"`
	CustomData   *types.JSON  `json:"custom_data,omitempty"`
	ErrorDetails string       `json:"error_details"`
	SpaceID      string       `json:"space_id"`
	Topic        *ReadTopic   `json:"topic,omitempty"`
	Channel      *ReadChannel `json:"channel,omitempty"`
	CreatedBy    *string      `json:"created_by,omitempty"`
	CreatedAt    *int64       `json:"created_at,omitempty"`
	UpdatedBy    *string      `json:"updated_by,omitempty"`
	UpdatedAt    *int64       `json:"updated_at,omitempty"`
}

// GetCursorValue returns cursor value
func (r *ReadDistribution) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListDistributionParams for listing distributions
type ListDistributionParams struct {
	Cursor      string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit       int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction   string `form:"direction,omitempty" json:"direction,omitempty"`
	TopicID     string `form:"topic_id,omitempty" json:"topic_id,omitempty"`
	ChannelID   string `form:"channel_id,omitempty" json:"channel_id,omitempty"`
	Status      int    `form:"status,omitempty" json:"status,omitempty"`
	WithTopic   bool   `form:"with_topic,omitempty" json:"with_topic,omitempty"`
	WithChannel bool   `form:"with_channel,omitempty" json:"with_channel,omitempty"`
	SpaceID     string `form:"space_id,omitempty" json:"space_id,omitempty"`
}

// FindDistribution for finding distribution
type FindDistribution struct {
	Distribution string `json:"distribution,omitempty"`
	TopicID      string `json:"topic_id,omitempty"`
	ChannelID    string `json:"channel_id,omitempty"`
	SpaceID      string `json:"space_id,omitempty"`
}
