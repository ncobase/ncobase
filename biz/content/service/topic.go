package service

import (
	"context"
	"errors"
	"ncobase/biz/content/data"
	"ncobase/biz/content/data/ent"
	"ncobase/biz/content/data/repository"
	"ncobase/biz/content/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/slug"
	"github.com/ncobase/ncore/validation/validator"
)

// TopicServiceInterface for topic service operations
type TopicServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTopic, error)
	Get(ctx context.Context, slug string) (*structs.ReadTopic, error)
	GetByID(ctx context.Context, id string) (*structs.ReadTopic, error) // Add this method
	List(ctx context.Context, params *structs.ListTopicParams) (paging.Result[*structs.ReadTopic], error)
	Delete(ctx context.Context, slug string) error
}

type topicService struct {
	r  repository.TopicRepositoryInterface
	ts TaxonomyServiceInterface
}

// NewTopicService creates new topic service
func NewTopicService(d *data.Data, ts TaxonomyServiceInterface) TopicServiceInterface {
	return &topicService{
		r:  repository.NewTopicRepository(d),
		ts: ts,
	}
}

// Create creates new topic
func (s *topicService) Create(ctx context.Context, body *structs.CreateTopicBody) (*structs.ReadTopic, error) {
	// Validate taxonomy if provided
	if validator.IsNotEmpty(body.TaxonomyID) && s.ts != nil {
		_, err := s.ts.Get(ctx, body.TaxonomyID)
		if err != nil {
			return nil, errors.New("invalid taxonomy_id: taxonomy not found")
		}
	}

	// Set slug field
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Update updates existing topic
func (s *topicService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTopic, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug / id"))
	}

	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Validate taxonomy if being updated
	if taxonomyID, ok := updates["taxonomy_id"].(string); ok && validator.IsNotEmpty(taxonomyID) {
		if s.ts != nil {
			_, err := s.ts.Get(ctx, taxonomyID)
			if err != nil {
				return nil, errors.New("invalid taxonomy_id: taxonomy not found")
			}
		}
	}

	row, err := s.r.Update(ctx, slug, updates)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Get retrieves topic by slug
func (s *topicService) Get(ctx context.Context, slug string) (*structs.ReadTopic, error) {
	row, err := s.r.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// GetByID retrieves topic by ID
func (s *topicService) GetByID(ctx context.Context, id string) (*structs.ReadTopic, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Delete deletes topic by slug
func (s *topicService) Delete(ctx context.Context, slug string) error {
	err := s.r.Delete(ctx, slug)
	if err := handleEntError(ctx, "Topic", err); err != nil {
		return err
	}

	return nil
}

// List lists all topics
func (s *topicService) List(ctx context.Context, params *structs.ListTopicParams) (paging.Result[*structs.ReadTopic], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTopic, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.r.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing topics: %v", err)
			return nil, 0, err
		}

		total := s.r.CountX(ctx, params)

		return s.serializes(ctx, rows), total, nil
	})
}

// serializes converts multiple ent.Topic to []*structs.ReadTopic
func (s *topicService) serializes(ctx context.Context, rows []*ent.Topic) []*structs.ReadTopic {
	rs := make([]*structs.ReadTopic, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.serialize(ctx, row))
	}
	return rs
}

// serialize converts ent.Topic to structs.ReadTopic
func (s *topicService) serialize(ctx context.Context, row *ent.Topic) *structs.ReadTopic {
	result := &structs.ReadTopic{
		ID:         row.ID,
		Name:       row.Name,
		Title:      row.Title,
		Slug:       row.Slug,
		Content:    row.Content,
		Thumbnail:  row.Thumbnail,
		Temp:       row.Temp,
		Markdown:   row.Markdown,
		Private:    row.Private,
		Status:     row.Status,
		Released:   row.Released,
		TaxonomyID: row.TaxonomyID,
		SpaceID:    row.SpaceID,
		CreatedBy:  &row.CreatedBy,
		CreatedAt:  &row.CreatedAt,
		UpdatedBy:  &row.UpdatedBy,
		UpdatedAt:  &row.UpdatedAt,
	}

	// Extract additional fields from extras if they exist
	if row.Extras != nil {
		if version, ok := row.Extras["version"].(float64); ok {
			result.Version = int(version)
		}
		if contentType, ok := row.Extras["content_type"].(string); ok {
			result.ContentType = contentType
		}
		if seoTitle, ok := row.Extras["seo_title"].(string); ok {
			result.SEOTitle = seoTitle
		}
		if seoDescription, ok := row.Extras["seo_description"].(string); ok {
			result.SEODescription = seoDescription
		}
		if seoKeywords, ok := row.Extras["seo_keywords"].(string); ok {
			result.SEOKeywords = seoKeywords
		}
		if excerptAuto, ok := row.Extras["excerpt_auto"].(bool); ok {
			result.ExcerptAuto = excerptAuto
		}
		if excerpt, ok := row.Extras["excerpt"].(string); ok {
			result.Excerpt = excerpt
		}
		if featuredMedia, ok := row.Extras["featured_media"].(string); ok {
			result.FeaturedMedia = featuredMedia
		}
		if tags, ok := row.Extras["tags"].([]any); ok {
			tagStrings := make([]string, 0, len(tags))
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					tagStrings = append(tagStrings, tagStr)
				}
			}
			result.Tags = tagStrings
		}
		result.Metadata = &row.Extras
	}

	// Load taxonomy if available and service exists
	if validator.IsNotEmpty(row.TaxonomyID) && s.ts != nil {
		if taxonomy, err := s.ts.Get(ctx, row.TaxonomyID); err == nil {
			result.Taxonomy = taxonomy
		} else {
			logger.Warnf(ctx, "Failed to load taxonomy %s: %v", row.TaxonomyID, err)
		}
	}

	return result
}
