package structs

import (
	"fmt"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// FindTopic for finding topic
type FindTopic struct {
	Topic    string `json:"topic,omitempty"`
	Taxonomy string `json:"taxonomy,omitempty"`
	SpaceID  string `json:"space_id,omitempty"`
}

// TopicBody represents common fields for creating and updating topics
type TopicBody struct {
	Name           string      `json:"name,omitempty"`
	Title          string      `json:"title,omitempty"`
	Slug           string      `json:"slug,omitempty"`
	Content        string      `json:"content,omitempty"`
	Thumbnail      string      `json:"thumbnail,omitempty"`
	Temp           bool        `json:"temp,omitempty"`
	Markdown       bool        `json:"markdown,omitempty"`
	Private        bool        `json:"private,omitempty"`
	Status         int         `json:"status,omitempty"`
	Version        int         `json:"version,omitempty"`
	ContentType    string      `json:"content_type,omitempty"` // article, video, etc.
	SEOTitle       string      `json:"seo_title,omitempty"`
	SEODescription string      `json:"seo_description,omitempty"`
	SEOKeywords    string      `json:"seo_keywords,omitempty"`
	ExcerptAuto    bool        `json:"excerpt_auto,omitempty"`
	Excerpt        string      `json:"excerpt,omitempty"`
	FeaturedMedia  string      `json:"featured_media,omitempty"`
	Tags           []string    `json:"tags,omitempty"`
	Metadata       *types.JSON `json:"metadata,omitempty"`
	Released       int64       `json:"released,omitempty"`
	TaxonomyID     string      `json:"taxonomy_id,omitempty"`
	SpaceID        string      `json:"space_id,omitempty"`
	CreatedBy      *string     `json:"created_by,omitempty"`
	UpdatedBy      *string     `json:"updated_by,omitempty"`
}

// CreateTopicBody for creating topic
type CreateTopicBody struct {
	TopicBody
}

// UpdateTopicBody for updating topic
type UpdateTopicBody struct {
	ID string `json:"id"`
	TopicBody
}

// ReadTopic represents output schema for retrieving topic
type ReadTopic struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Title          string        `json:"title"`
	Slug           string        `json:"slug"`
	Content        string        `json:"content"`
	Thumbnail      string        `json:"thumbnail"`
	Temp           bool          `json:"temp"`
	Markdown       bool          `json:"markdown"`
	Private        bool          `json:"private"`
	Status         int           `json:"status"`
	Version        int           `json:"version"`
	ContentType    string        `json:"content_type"`
	SEOTitle       string        `json:"seo_title"`
	SEODescription string        `json:"seo_description"`
	SEOKeywords    string        `json:"seo_keywords"`
	ExcerptAuto    bool          `json:"excerpt_auto"`
	Excerpt        string        `json:"excerpt"`
	FeaturedMedia  string        `json:"featured_media"`
	Tags           []string      `json:"tags"`
	Metadata       *types.JSON   `json:"metadata,omitempty"`
	Released       int64         `json:"released"`
	TaxonomyID     string        `json:"taxonomy_id"`
	SpaceID        string        `json:"space_id"`
	Media          []*ReadMedia  `json:"media,omitempty"`
	Taxonomy       *ReadTaxonomy `json:"taxonomy,omitempty"`
	CreatedBy      *string       `json:"created_by,omitempty"`
	CreatedAt      *int64        `json:"created_at,omitempty"`
	UpdatedBy      *string       `json:"updated_by,omitempty"`
	UpdatedAt      *int64        `json:"updated_at,omitempty"`
}

// GetCursorValue returns cursor value
func (r *ReadTopic) GetCursorValue() string {
	return fmt.Sprintf("%s:%d", r.ID, convert.ToValue(r.CreatedAt))
}

// ListTopicParams for listing topics
type ListTopicParams struct {
	Cursor    string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit     int    `form:"limit,omitempty" json:"limit,omitempty"`
	Direction string `form:"direction,omitempty" json:"direction,omitempty"`
	Taxonomy  string `form:"taxonomy,omitempty" json:"taxonomy,omitempty"`
	SpaceID   string `form:"space_id,omitempty" json:"space_id,omitempty"`
}
