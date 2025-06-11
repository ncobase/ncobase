package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

const (
	// MediaTypes for content management
	MediaTypeImage string = "image"
	MediaTypeVideo string = "video"
	MediaTypeAudio string = "audio"
	MediaTypeFile  string = "file"
)

// MediaBody represents common fields for creating and updating media
type MediaBody struct {
	Title       string      `json:"title,omitempty"`
	Type        string      `json:"type,omitempty"`        // image, video, audio, file
	ResourceID  string      `json:"resource_id,omitempty"` // Reference to resource plugin file
	URL         string      `json:"url,omitempty"`         // For external resources
	Description string      `json:"description,omitempty"`
	Alt         string      `json:"alt,omitempty"`
	Metadata    *types.JSON `json:"metadata,omitempty"`
	SpaceID     string      `json:"space_id,omitempty"`
	OwnerID     string      `json:"owner_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateMediaBody for creating media
type CreateMediaBody struct {
	MediaBody
}

// UpdateMediaBody for updating media
type UpdateMediaBody struct {
	ID string `json:"id"`
	MediaBody
}

// ReadMedia represents output schema for retrieving media
type ReadMedia struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Alt         string                 `json:"alt"`
	Metadata    *types.JSON            `json:"metadata,omitempty"`
	SpaceID     string                 `json:"space_id"`
	OwnerID     string                 `json:"owner_id"`
	Resource    *ResourceFileReference `json:"resource,omitempty"` // Resource file reference
	CreatedBy   *string                `json:"created_by,omitempty"`
	CreatedAt   *int64                 `json:"created_at,omitempty"`
	UpdatedBy   *string                `json:"updated_by,omitempty"`
	UpdatedAt   *int64                 `json:"updated_at,omitempty"`
}

// ResourceFileReference represents reference to resource plugin file
type ResourceFileReference struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"`
	Size         *int   `json:"size"`
	Storage      string `json:"storage"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	IsExpired    bool   `json:"is_expired,omitempty"`
}

// GetCursorValue returns cursor value
func (r *ReadMedia) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListMediaParams for listing media
type ListMediaParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Search    string `form:"search,omitempty" json:"search,omitempty"`
	SpaceID   string `form:"space_id,omitempty" json:"space_id,omitempty"`
	OwnerID   string `form:"owner_id,omitempty" json:"owner_id,omitempty"`
}

// FindMedia for finding media
type FindMedia struct {
	Media   string `json:"media,omitempty"`
	SpaceID string `json:"space_id,omitempty"`
	OwnerID string `json:"owner_id,omitempty"`
}
