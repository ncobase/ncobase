package structs

import (
	"fmt"

	"github.com/ncobase/ncore/utils/convert"
)

// TopicMediaBody represents the common fields for creating and updating topic media relation.
type TopicMediaBody struct {
	TopicID   string  `json:"topic_id,omitempty"`
	MediaID   string  `json:"media_id,omitempty"`
	Type      string  `json:"type,omitempty"` // featured, gallery, attachment, etc.
	Order     int     `json:"order,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
}

// CreateTopicMediaBody represents the body for creating a topic media relation.
type CreateTopicMediaBody struct {
	TopicMediaBody
}

// UpdateTopicMediaBody represents the body for updating a topic media relation.
type UpdateTopicMediaBody struct {
	ID string `json:"id"`
	TopicMediaBody
}

// ReadTopicMedia represents the output schema for retrieving a topic media relation.
type ReadTopicMedia struct {
	ID        string     `json:"id"`
	TopicID   string     `json:"topic_id"`
	MediaID   string     `json:"media_id"`
	Type      string     `json:"type"`
	Order     int        `json:"order"`
	Media     *ReadMedia `json:"media,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt *int64     `json:"created_at,omitempty"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	UpdatedAt *int64     `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadTopicMedia) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListTopicMediaParams represents the parameters for listing topic media relations.
type ListTopicMediaParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	TopicID   string `form:"topic_id,omitempty" json:"topic_id,omitempty"`
	MediaID   string `form:"media_id,omitempty" json:"media_id,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	WithMedia bool   `form:"with_media,omitempty" json:"with_media,omitempty"`
}

// FindTopicMedia represents the parameters for finding a topic media relation.
type FindTopicMedia struct {
	TopicMedia string `json:"topic_media,omitempty"`
	TopicID    string `json:"topic_id,omitempty"`
	MediaID    string `json:"media_id,omitempty"`
}
