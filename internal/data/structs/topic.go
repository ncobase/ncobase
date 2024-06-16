package structs

import (
	"time"
)

// FindTopic represents the parameters for finding a topic.
type FindTopic struct {
	ID         string `json:"id,omitempty"`
	Slug       string `json:"slug,omitempty"`
	TaxonomyID string `json:"taxonomy_id,omitempty"`
	TenantID   string `json:"tenant_id,omitempty"`
}

// TopicBody represents the common fields for creating and updating topics.
type TopicBody struct {
	BaseEntity
	Name       string    `json:"name,omitempty"`
	Title      string    `json:"title,omitempty"`
	Slug       string    `json:"slug,omitempty"`
	Content    string    `json:"content,omitempty"`
	Thumbnail  string    `json:"thumbnail,omitempty"`
	Temp       bool      `json:"temp,omitempty"`
	Markdown   bool      `json:"markdown,omitempty"`
	Private    bool      `json:"private,omitempty"`
	Status     int       `json:"status,omitempty"`
	Released   time.Time `json:"released,omitempty"`
	TaxonomyID string    `json:"taxonomy_id,omitempty"`
	TenantID   string    `json:"tenant_id,omitempty"`
}

// CreateTopicBody represents the body for creating a topic.
type CreateTopicBody struct {
	TopicBody
}

// UpdateTopicBody represents the body for updating a topic.
type UpdateTopicBody struct {
	ID string `json:"id"`
	TopicBody
}

// ReadTopic represents the output schema for retrieving a topic.
type ReadTopic struct {
	BaseEntity
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Content    string    `json:"content"`
	Thumbnail  string    `json:"thumbnail"`
	Temp       bool      `json:"temp"`
	Markdown   bool      `json:"markdown"`
	Private    bool      `json:"private"`
	Status     int       `json:"status"`
	Released   time.Time `json:"released"`
	TaxonomyID string    `json:"taxonomy_id"`
	TenantID   string    `json:"tenant_id"`
}

// ListTopicParams represents the parameters for listing topics.
type ListTopicParams struct {
	Cursor     string `form:"cursor,omitempty" json:"cursor,omitempty"`
	Limit      int    `form:"limit,omitempty" json:"limit,omitempty"`
	TaxonomyID string `form:"taxonomy_id,omitempty" json:"taxonomy_id,omitempty"`
	TenantID   string `form:"tenant_id,omitempty" json:"tenant_id,omitempty"`
}
