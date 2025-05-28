package service

import (
	"context"
	"errors"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	"ncobase/content/data/repository"
	"ncobase/content/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// DistributionServiceInterface is the interface for the service.
type DistributionServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateDistributionBody) (*structs.ReadDistribution, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadDistribution, error)
	Get(ctx context.Context, id string) (*structs.ReadDistribution, error)
	List(ctx context.Context, params *structs.ListDistributionParams) (paging.Result[*structs.ReadDistribution], error)
	Delete(ctx context.Context, id string) error
	GetByTopicAndChannel(ctx context.Context, topicID string, channelID string) (*structs.ReadDistribution, error)
	GetPendingDistributions(ctx context.Context, limit int) ([]*structs.ReadDistribution, error)
	GetScheduledDistributions(ctx context.Context, before int64, limit int) ([]*structs.ReadDistribution, error)
	Publish(ctx context.Context, id string) (*structs.ReadDistribution, error)
	Cancel(ctx context.Context, id string, reason string) (*structs.ReadDistribution, error)
}

// distributionService is the struct for the service.
type distributionService struct {
	r repository.DistributionRepositoryInterface
}

// NewDistributionService creates a new service.
func NewDistributionService(d *data.Data) DistributionServiceInterface {
	return &distributionService{
		r: repository.NewDistributionRepository(d),
	}
}

// Create creates a new distribution.
func (s *distributionService) Create(ctx context.Context, body *structs.CreateDistributionBody) (*structs.ReadDistribution, error) {
	if validator.IsEmpty(body.TopicID) {
		return nil, errors.New(ecode.FieldIsRequired("topic_id"))
	}
	if validator.IsEmpty(body.ChannelID) {
		return nil, errors.New(ecode.FieldIsRequired("channel_id"))
	}

	// Check if distribution already exists for this topic and channel
	existing, err := s.r.GetByTopicAndChannel(ctx, body.TopicID, body.ChannelID)
	if err == nil && existing != nil {
		return nil, errors.New(ecode.AlreadyExist("distribution for this topic and channel"))
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing distribution.
func (s *distributionService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadDistribution, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a distribution by ID.
func (s *distributionService) Get(ctx context.Context, id string) (*structs.ReadDistribution, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a distribution by ID.
func (s *distributionService) Delete(ctx context.Context, id string) error {
	err := s.r.Delete(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return err
	}

	return nil
}

// List lists all distributions with pagination.
func (s *distributionService) List(ctx context.Context, params *structs.ListDistributionParams) (paging.Result[*structs.ReadDistribution], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadDistribution, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, count, err := s.r.ListWithCount(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing distributions: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), count, nil
	})
}

// GetByTopicAndChannel gets a distribution by topic ID and channel ID.
func (s *distributionService) GetByTopicAndChannel(ctx context.Context, topicID string, channelID string) (*structs.ReadDistribution, error) {
	row, err := s.r.GetByTopicAndChannel(ctx, topicID, channelID)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetPendingDistributions gets a list of pending distributions.
func (s *distributionService) GetPendingDistributions(ctx context.Context, limit int) ([]*structs.ReadDistribution, error) {
	rows, err := s.r.GetPendingDistributions(ctx, limit)
	if err != nil {
		logger.Errorf(ctx, "Error getting pending distributions: %v", err)
		return nil, err
	}

	return s.Serializes(rows), nil
}

// GetScheduledDistributions gets a list of scheduled distributions.
func (s *distributionService) GetScheduledDistributions(ctx context.Context, before int64, limit int) ([]*structs.ReadDistribution, error) {
	rows, err := s.r.GetScheduledDistributions(ctx, before, limit)
	if err != nil {
		logger.Errorf(ctx, "Error getting scheduled distributions: %v", err)
		return nil, err
	}

	return s.Serializes(rows), nil
}

// Publish publishes a distribution.
func (s *distributionService) Publish(ctx context.Context, id string) (*structs.ReadDistribution, error) {
	dist, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	if dist.Status == structs.DistributionStatusPublished {
		return s.Serialize(dist), nil
	}

	updates := types.JSON{
		"status":       structs.DistributionStatusPublished,
		"published_at": time.Now().Unix(),
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Cancel cancels a distribution.
func (s *distributionService) Cancel(ctx context.Context, id string, reason string) (*structs.ReadDistribution, error) {
	dist, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	if dist.Status == structs.DistributionStatusCancelled {
		return s.Serialize(dist), nil
	}

	updates := types.JSON{
		"status":        structs.DistributionStatusCancelled,
		"error_details": reason,
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Serializes converts multiple ent.Distribution to []*structs.ReadDistribution.
func (s *distributionService) Serializes(rows []*ent.Distribution) []*structs.ReadDistribution {
	var rs []*structs.ReadDistribution
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize converts an ent.Distribution to a structs.ReadDistribution.
func (s *distributionService) Serialize(row *ent.Distribution) *structs.ReadDistribution {
	result := &structs.ReadDistribution{
		ID:           row.ID,
		TopicID:      row.TopicID,
		ChannelID:    row.ChannelID,
		Status:       row.Status,
		ScheduledAt:  row.ScheduledAt,
		PublishedAt:  row.PublishedAt,
		MetaData:     &row.Extras,
		ExternalID:   row.ExternalID,
		ExternalURL:  row.ExternalURL,
		CustomData:   &row.Extras,
		ErrorDetails: row.ErrorDetails,
		TenantID:     row.TenantID,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}

	// Load related entities if available
	if row.Edges.Topic != nil {
		topicService := &topicService{}
		result.Topic = topicService.Serialize(row.Edges.Topic)
	}

	if row.Edges.Channel != nil {
		channelService := &channelService{}
		result.Channel = channelService.Serialize(row.Edges.Channel)
	}

	return result
}
