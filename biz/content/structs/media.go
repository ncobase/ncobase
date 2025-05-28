package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

const (
	// MediaTypes
	MediaTypeImage string = "image"
	MediaTypeVideo string = "video"
	MediaTypeAudio string = "audio"
	MediaTypeFile  string = "file"
)

// MediaBody represents the common fields for creating and updating media.
type MediaBody struct {
	Title       string      `json:"title,omitempty"`
	Type        string      `json:"type,omitempty"` // image, video, audio, file
	URL         string      `json:"url,omitempty"`
	Path        string      `json:"path,omitempty"`
	MimeType    string      `json:"mime_type,omitempty"`
	Size        int64       `json:"size,omitempty"`
	Width       int         `json:"width,omitempty"`
	Height      int         `json:"height,omitempty"`
	Duration    float64     `json:"duration,omitempty"`
	Description string      `json:"description,omitempty"`
	Alt         string      `json:"alt,omitempty"`
	Metadata    *types.JSON `json:"metadata,omitempty"`
	TenantID    string      `json:"tenant_id,omitempty"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

// CreateMediaBody represents the body for creating media.
type CreateMediaBody struct {
	MediaBody
}

// UpdateMediaBody represents the body for updating media.
type UpdateMediaBody struct {
	ID string `json:"id"`
	MediaBody
}

// ReadMedia represents the output schema for retrieving media.
type ReadMedia struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	URL         string      `json:"url"`
	Path        string      `json:"path"`
	MimeType    string      `json:"mime_type"`
	Size        int64       `json:"size"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	Duration    float64     `json:"duration"`
	Description string      `json:"description"`
	Alt         string      `json:"alt"`
	Metadata    *types.JSON `json:"metadata,omitempty"`
	TenantID    string      `json:"tenant_id"`
	CreatedBy   *string     `json:"created_by,omitempty"`
	CreatedAt   *int64      `json:"created_at,omitempty"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
	UpdatedAt   *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadMedia) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListMediaParams represents the parameters for listing media.
type ListMediaParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Search    string `form:"search,omitempty" json:"search,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
}

// FindMedia represents the parameters for finding media.
type FindMedia struct {
	Media  string `json:"media,omitempty"`
	Tenant string `json:"tenant,omitempty"`
}
