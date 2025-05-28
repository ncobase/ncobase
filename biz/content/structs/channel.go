package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

const (
	// ChannelTypes

	ChannelTypeWebsite     string = "website"
	ChannelTypeWechat      string = "wechat"
	ChannelTypeDouyin      string = "douyin"
	ChannelTypeTiktok      string = "tiktok"
	ChannelTypeXiaohongshu string = "xiaohongshu"
	ChannelTypeTwitter     string = "twitter"
	ChannelTypeFacebook    string = "facebook"
	ChannelTypeCustom      string = "custom"

	// ContentTypes

	ContentTypeArticle string = "article"
	ContentTypeVideo   string = "video"
	ContentTypeImage   string = "image"
	ContentTypeAudio   string = "audio"
	ContentTypeMixed   string = "mixed"
)

// ChannelBody represents the common fields for creating and updating a distribution channel.
type ChannelBody struct {
	Name          string      `json:"name,omitempty"`
	Type          string      `json:"type,omitempty"` // website, wechat, douyin, tiktok, xiaohongshu, twitter, facebook, custom
	Slug          string      `json:"slug,omitempty"`
	Icon          string      `json:"icon,omitempty"`
	Status        int         `json:"status,omitempty"`        // 0: active, 1: inactive
	AllowedTypes  []string    `json:"allowed_types,omitempty"` // article, video, image, audio, mixed
	Config        *types.JSON `json:"config,omitempty"`        // API keys, secrets, endpoints, etc.
	Description   string      `json:"description,omitempty"`
	Logo          string      `json:"logo,omitempty"`
	WebhookURL    string      `json:"webhook_url,omitempty"`
	AutoPublish   bool        `json:"auto_publish,omitempty"`
	RequireReview bool        `json:"require_review,omitempty"`
	TenantID      string      `json:"tenant_id,omitempty"`
	CreatedBy     *string     `json:"created_by,omitempty"`
	UpdatedBy     *string     `json:"updated_by,omitempty"`
}

// CreateChannelBody represents the body for creating a channel.
type CreateChannelBody struct {
	ChannelBody
}

// UpdateChannelBody represents the body for updating a channel.
type UpdateChannelBody struct {
	ID string `json:"id"`
	ChannelBody
}

// ReadChannel represents the output schema for retrieving a channel.
type ReadChannel struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Slug          string      `json:"slug"`
	Icon          string      `json:"icon"`
	Status        int         `json:"status"`
	AllowedTypes  []string    `json:"allowed_types"`
	Config        *types.JSON `json:"config,omitempty"`
	Description   string      `json:"description"`
	Logo          string      `json:"logo"`
	WebhookURL    string      `json:"webhook_url"`
	AutoPublish   bool        `json:"auto_publish"`
	RequireReview bool        `json:"require_review"`
	TenantID      string      `json:"tenant_id"`
	CreatedBy     *string     `json:"created_by,omitempty"`
	CreatedAt     *int64      `json:"created_at,omitempty"`
	UpdatedBy     *string     `json:"updated_by,omitempty"`
	UpdatedAt     *int64      `json:"updated_at,omitempty"`
}

// GetCursorValue returns the cursor value.
func (r *ReadChannel) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListChannelParams represents the parameters for listing channels.
type ListChannelParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Type      string `form:"type,omitempty" json:"type,omitempty"`
	Status    int    `form:"status,omitempty" json:"status,omitempty"`
	Tenant    string `form:"tenant,omitempty" json:"tenant,omitempty"`
}

// FindChannel represents the parameters for finding a channel.
type FindChannel struct {
	Channel string `json:"channel,omitempty"`
	Type    string `json:"type,omitempty"`
	Tenant  string `json:"tenant,omitempty"`
}
