package structs

import (
	"time"
)

// FindTopic represents the parameters for finding a topic.
type FindTopic struct {
	ID         string `json:"id,omitempty"`
	Slug       string `json:"slug,omitempty"`
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	DomainID   string `json:"domain_id,omitempty"`
}

// TopicBody - Common fields for creating and updating topics
type TopicBody struct {
	Name       string    `json:"name,omitempty"`
	Title      string    `json:"title,omitempty"`
	Slug       string    `json:"slug,omitempty"`
	Content    string    `json:"content,omitempty"`
	Thumbnail  string    `json:"thumbnail,omitempty"`
	Temp       bool      `json:"temp,omitempty"`
	Markdown   bool      `json:"markdown,omitempty"`
	Private    bool      `json:"private,omitempty"`
	Status     int32     `json:"status,omitempty"`
	Released   time.Time `json:"released,omitempty"`
	TaxonomyID string    `json:"taxonomy_id,omitempty"`
	DomainID   string    `json:"domain_id,omitempty"`
	CreatedBy  string    `json:"created_by,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedBy  string    `json:"updated_by,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// CreateTopicBody - Create topic body
type CreateTopicBody struct {
	TopicBody
}

// UpdateTopicBody - Update topic body
type UpdateTopicBody struct {
	TopicBody
	ID string `json:"id"`
}

// ReadTopic - Output topic schema
type ReadTopic struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Content    string    `json:"content"`
	Thumbnail  string    `json:"thumbnail"`
	Temp       bool      `json:"temp"`
	Markdown   bool      `json:"markdown"`
	Private    bool      `json:"private"`
	Status     int32     `json:"status"`
	Released   bool      `json:"released"`
	TaxonomyID string    `json:"taxonomy_id"`
	DomainID   string    `json:"domain_id"`
	CreatedBy  string    `json:"created_by,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedBy  string    `json:"updated_by,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// ListTopicParams - Query topic list params
type ListTopicParams struct {
	Cursor   string `form:"cursor" json:"cursor"`
	Limit    int64  `form:"limit" json:"limit"`
	Taxonomy string `form:"taxonomy,omitempty" json:"taxonomy,omitempty"`
	Domain   string `form:"domain,omitempty" json:"domain,omitempty"`
}
