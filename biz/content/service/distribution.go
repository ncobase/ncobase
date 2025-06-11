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

// DistributionServiceInterface for distribution service operations
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

type distributionService struct {
	r  repository.DistributionRepositoryInterface
	ts TopicServiceInterface   // Topic service dependency
	cs ChannelServiceInterface // Channel service dependency
}

// NewDistributionService creates new distribution service
func NewDistributionService(d *data.Data, ts TopicServiceInterface, cs ChannelServiceInterface) DistributionServiceInterface {
	return &distributionService{
		r:  repository.NewDistributionRepository(d),
		ts: ts,
		cs: cs,
	}
}

// Create creates new distribution
func (s *distributionService) Create(ctx context.Context, body *structs.CreateDistributionBody) (*structs.ReadDistribution, error) {
	if validator.IsEmpty(body.TopicID) {
		return nil, errors.New(ecode.FieldIsRequired("topic_id"))
	}
	if validator.IsEmpty(body.ChannelID) {
		return nil, errors.New(ecode.FieldIsRequired("channel_id"))
	}

	// Validate topic exists
	if s.ts != nil {
		_, err := s.ts.GetByID(ctx, body.TopicID)
		if err != nil {
			return nil, errors.New("invalid topic_id: topic not found")
		}
	}

	// Validate channel exists
	if s.cs != nil {
		_, err := s.cs.Get(ctx, body.ChannelID)
		if err != nil {
			return nil, errors.New("invalid channel_id: channel not found")
		}
	}

	// Check if distribution already exists for this topic and channel
	existing, err := s.r.GetByTopicAndChannel(ctx, body.TopicID, body.ChannelID)
	if err == nil && existing != nil {
		return nil, errors.New(ecode.AlreadyExist("distribution for this topic and channel"))
	}

	// Set default status
	if body.Status == 0 {
		body.Status = structs.DistributionStatusDraft
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Update updates existing distribution
func (s *distributionService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadDistribution, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Validate topic if being updated
	if topicID, ok := updates["topic_id"].(string); ok && validator.IsNotEmpty(topicID) {
		if s.ts != nil {
			_, err := s.ts.GetByID(ctx, topicID)
			if err != nil {
				return nil, errors.New("invalid topic_id: topic not found")
			}
		}
	}

	// Validate channel if being updated
	if channelID, ok := updates["channel_id"].(string); ok && validator.IsNotEmpty(channelID) {
		if s.cs != nil {
			_, err := s.cs.Get(ctx, channelID)
			if err != nil {
				return nil, errors.New("invalid channel_id: channel not found")
			}
		}
	}

	// Handle space_id/tenant_id compatibility
	if spaceID, ok := updates["space_id"].(string); ok {
		updates["tenant_id"] = spaceID
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Get retrieves distribution by ID
func (s *distributionService) Get(ctx context.Context, id string) (*structs.ReadDistribution, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Delete deletes distribution by ID
func (s *distributionService) Delete(ctx context.Context, id string) error {
	err := s.r.Delete(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return err
	}

	return nil
}

// List lists all distributions
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

		return s.serializes(ctx, rows), count, nil
	})
}

// GetByTopicAndChannel gets distribution by topic and channel IDs
func (s *distributionService) GetByTopicAndChannel(ctx context.Context, topicID string, channelID string) (*structs.ReadDistribution, error) {
	row, err := s.r.GetByTopicAndChannel(ctx, topicID, channelID)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// GetPendingDistributions gets list of pending distributions
func (s *distributionService) GetPendingDistributions(ctx context.Context, limit int) ([]*structs.ReadDistribution, error) {
	rows, err := s.r.GetPendingDistributions(ctx, limit)
	if err != nil {
		logger.Errorf(ctx, "Error getting pending distributions: %v", err)
		return nil, err
	}

	return s.serializes(ctx, rows), nil
}

// GetScheduledDistributions gets list of scheduled distributions
func (s *distributionService) GetScheduledDistributions(ctx context.Context, before int64, limit int) ([]*structs.ReadDistribution, error) {
	rows, err := s.r.GetScheduledDistributions(ctx, before, limit)
	if err != nil {
		logger.Errorf(ctx, "Error getting scheduled distributions: %v", err)
		return nil, err
	}

	return s.serializes(ctx, rows), nil
}

// Publish publishes distribution
func (s *distributionService) Publish(ctx context.Context, id string) (*structs.ReadDistribution, error) {
	dist, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	if dist.Status == structs.DistributionStatusPublished {
		return s.serialize(ctx, dist), nil
	}

	updates := types.JSON{
		"status":       structs.DistributionStatusPublished,
		"published_at": time.Now().Unix(),
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// Cancel cancels distribution
func (s *distributionService) Cancel(ctx context.Context, id string, reason string) (*structs.ReadDistribution, error) {
	dist, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	if dist.Status == structs.DistributionStatusCancelled {
		return s.serialize(ctx, dist), nil
	}

	updates := types.JSON{
		"status":        structs.DistributionStatusCancelled,
		"error_details": reason,
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Distribution", err); err != nil {
		return nil, err
	}

	return s.serialize(ctx, row), nil
}

// serializes converts multiple ent.Distribution to []*structs.ReadDistribution
func (s *distributionService) serializes(ctx context.Context, rows []*ent.Distribution) []*structs.ReadDistribution {
	rs := make([]*structs.ReadDistribution, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.serialize(ctx, row))
	}
	return rs
}

// serialize converts ent.Distribution to structs.ReadDistribution
func (s *distributionService) serialize(ctx context.Context, row *ent.Distribution) *structs.ReadDistribution {
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
		SpaceID:      row.SpaceID,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}

	// Load related entities if services are available and eager loading is not done
	if validator.IsNotEmpty(row.TopicID) && s.ts != nil {
		if topic, err := s.ts.GetByID(ctx, row.TopicID); err == nil {
			result.Topic = topic
		} else {
			logger.Warnf(ctx, "Failed to load topic %s: %v", row.TopicID, err)
		}
	}

	if validator.IsNotEmpty(row.ChannelID) && s.cs != nil {
		if channel, err := s.cs.Get(ctx, row.ChannelID); err == nil {
			result.Channel = channel
		} else {
			logger.Warnf(ctx, "Failed to load channel %s: %v", row.ChannelID, err)
		}
	}

	return result
}
