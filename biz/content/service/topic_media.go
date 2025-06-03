package service

import (
	"context"
	"errors"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	"ncobase/content/data/repository"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// TopicMediaServiceInterface is the interface for the service.
type TopicMediaServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTopicMediaBody) (*structs.ReadTopicMedia, error)
	Update(ctx context.Context, id string, topicID string, mediaID string, mediaType string, order int) (*structs.ReadTopicMedia, error)
	Get(ctx context.Context, id string) (*structs.ReadTopicMedia, error)
	List(ctx context.Context, params *structs.ListTopicMediaParams) (paging.Result[*structs.ReadTopicMedia], error)
	Delete(ctx context.Context, id string) error
	GetByTopicAndMedia(ctx context.Context, topicID string, mediaID string) (*structs.ReadTopicMedia, error)
}

// topicMediaService is the struct for the service.
type topicMediaService struct {
	r  repository.TopicMediaRepositoryInterface
	m  repository.MediaRepositoryInterface
	tr repository.TopicRepositoryInterface
}

// NewTopicMediaService creates a new service.
func NewTopicMediaService(d *data.Data) TopicMediaServiceInterface {
	return &topicMediaService{
		r:  repository.NewTopicMediaRepository(d),
		m:  repository.NewMediaRepository(d),
		tr: repository.NewTopicRepository(d),
	}
}

// Create creates a new topic media relation.
func (s *topicMediaService) Create(ctx context.Context, body *structs.CreateTopicMediaBody) (*structs.ReadTopicMedia, error) {
	if validator.IsEmpty(body.TopicID) {
		return nil, errors.New(ecode.FieldIsRequired("topic_id"))
	}

	if validator.IsEmpty(body.MediaID) {
		return nil, errors.New(ecode.FieldIsRequired("media_id"))
	}

	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	// Check if topic exists
	_, err := s.tr.GetByID(ctx, body.TopicID)
	if err != nil {
		logger.Errorf(ctx, "Topic not found: %v", err)
		return nil, errors.New(ecode.NotExist("topic"))
	}

	// Check if media exists
	_, err = s.m.GetByID(ctx, body.MediaID)
	if err != nil {
		logger.Errorf(ctx, "Media not found: %v", err)
		return nil, errors.New(ecode.NotExist("media"))
	}

	// Check if relation already exists
	existing, err := s.r.GetByTopicAndMedia(ctx, body.TopicID, body.MediaID)
	if err == nil && existing != nil {
		return nil, errors.New(ecode.AlreadyExist("topic-media relation"))
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "TopicMedia", err); err != nil {
		return nil, err
	}

	// Load the associated media to return a complete response
	return s.loadMediaForTopicMedia(ctx, row)
}

// Update updates an existing topic media relation.
func (s *topicMediaService) Update(ctx context.Context, id string, topicID string, mediaID string, mediaType string, order int) (*structs.ReadTopicMedia, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate if the topic and media exist if they're being updated
	if validator.IsNotEmpty(topicID) {
		_, err := s.tr.GetByID(ctx, topicID)
		if err != nil {
			logger.Errorf(ctx, "Topic not found: %v", err)
			return nil, errors.New(ecode.NotExist("topic"))
		}
	}

	if validator.IsNotEmpty(mediaID) {
		_, err := s.m.GetByID(ctx, mediaID)
		if err != nil {
			logger.Errorf(ctx, "Media not found: %v", err)
			return nil, errors.New(ecode.NotExist("media"))
		}
	}

	row, err := s.r.Update(ctx, id, topicID, mediaID, mediaType, order)
	if err := handleEntError(ctx, "TopicMedia", err); err != nil {
		return nil, err
	}

	// Load the associated media to return a complete response
	return s.loadMediaForTopicMedia(ctx, row)
}

// Get retrieves a topic media relation by ID.
func (s *topicMediaService) Get(ctx context.Context, id string) (*structs.ReadTopicMedia, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "TopicMedia", err); err != nil {
		return nil, err
	}

	// Load the associated media to return a complete response
	return s.loadMediaForTopicMedia(ctx, row)
}

// Delete deletes a topic media relation by ID.
func (s *topicMediaService) Delete(ctx context.Context, id string) error {
	err := s.r.Delete(ctx, id)
	if err := handleEntError(ctx, "TopicMedia", err); err != nil {
		return err
	}

	return nil
}

// List lists all topic media relations.
func (s *topicMediaService) List(ctx context.Context, params *structs.ListTopicMediaParams) (paging.Result[*structs.ReadTopicMedia], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTopicMedia, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.r.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing topic media relations: %v", err)
			return nil, 0, err
		}

		count, err := s.r.Count(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error counting topic media relations: %v", err)
			return nil, 0, err
		}

		// Load media for each relation
		result := make([]*structs.ReadTopicMedia, 0, len(rows))
		for _, row := range rows {
			topicMedia, err := s.loadMediaForTopicMedia(ctx, row)
			if err != nil {
				logger.Warnf(ctx, "Error loading media for topic media relation %s: %v", row.ID, err)
				// Continue with next item even if one fails
				result = append(result, s.Serialize(row))
				continue
			}
			result = append(result, topicMedia)
		}

		return result, count, nil
	})
}

// GetByTopicAndMedia gets a topic media relation by topic ID and media ID.
func (s *topicMediaService) GetByTopicAndMedia(ctx context.Context, topicID string, mediaID string) (*structs.ReadTopicMedia, error) {
	row, err := s.r.GetByTopicAndMedia(ctx, topicID, mediaID)
	if err := handleEntError(ctx, "TopicMedia", err); err != nil {
		return nil, err
	}

	// Load the associated media to return a complete response
	return s.loadMediaForTopicMedia(ctx, row)
}

// loadMediaForTopicMedia loads the related media for a topic media relation.
func (s *topicMediaService) loadMediaForTopicMedia(ctx context.Context, row *ent.TopicMedia) (*structs.ReadTopicMedia, error) {
	result := s.Serialize(row)

	// Load the associated media if withMedia is true
	if row.MediaID != "" {
		media, err := s.m.GetByID(ctx, row.MediaID)
		if err != nil {
			logger.Warnf(ctx, "Failed to load media for topic media relation: %v", err)
			// Continue with no media loaded
		} else {
			mediaService := &mediaService{}
			result.Media = mediaService.Serialize(media)
		}
	}

	return result, nil
}

// Serializes converts multiple ent.TopicMedia to []*structs.ReadTopicMedia.
func (s *topicMediaService) Serializes(rows []*ent.TopicMedia) []*structs.ReadTopicMedia {
	rs := make([]*structs.ReadTopicMedia, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize converts an ent.TopicMedia to a structs.ReadTopicMedia.
func (s *topicMediaService) Serialize(row *ent.TopicMedia) *structs.ReadTopicMedia {
	return &structs.ReadTopicMedia{
		ID:        row.ID,
		TopicID:   row.TopicID,
		MediaID:   row.MediaID,
		Type:      row.Type,
		Order:     row.Order,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
