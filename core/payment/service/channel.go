package service

import (
	"context"
	"fmt"
	"ncobase/core/payment/data/repository"
	"ncobase/core/payment/event"
	"ncobase/core/payment/structs"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// ChannelServiceInterface defines the interface for channel service operations
type ChannelServiceInterface interface {
	Create(ctx context.Context, input *structs.CreateChannelInput) (*structs.Channel, error)
	GetByID(ctx context.Context, id string) (*structs.Channel, error)
	Update(ctx context.Context, id string, updates *structs.UpdateChannelInput) (*structs.Channel, error)
	Delete(ctx context.Context, id string) error
	ChangeStatus(ctx context.Context, id string, status structs.ChannelStatus) (*structs.Channel, error)
	List(ctx context.Context, query *structs.ChannelQuery) (paging.Result[*structs.Channel], error)
	GetDefault(ctx context.Context, provider structs.PaymentProvider, tenantID string) (*structs.Channel, error)
	Serialize(channel *structs.Channel) *structs.Channel
	Serializes(channels []*structs.Channel) []*structs.Channel
}

// channelService provides operations for payment channels
type channelService struct {
	repo      repository.ChannelRepositoryInterface
	publisher event.PublisherInterface
}

// NewChannelService creates a new channel service
func NewChannelService(repo repository.ChannelRepositoryInterface, publisher event.PublisherInterface) ChannelServiceInterface {
	return &channelService{
		repo:      repo,
		publisher: publisher,
	}
}

// Create creates a new payment channel
func (s *channelService) Create(ctx context.Context, input *structs.CreateChannelInput) (*structs.Channel, error) {
	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("name"))
	}

	if input.Provider == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("provider"))
	}

	if len(input.SupportedType) == 0 {
		return nil, fmt.Errorf(ecode.FieldIsRequired("supported_types"))
	}

	if input.Config == nil {
		return nil, fmt.Errorf(ecode.FieldIsRequired("config"))
	}

	// Create channel entity
	channel := &structs.CreateChannelInput{
		Name:          input.Name,
		Provider:      input.Provider,
		Status:        input.Status,
		IsDefault:     input.IsDefault,
		SupportedType: input.SupportedType,
		Config:        input.Config,
		Metadata:      input.Metadata,
		TenantID:      input.TenantID,
	}

	// If this channel is set as default, unset any existing default for this provider
	if channel.IsDefault {
		if err := s.repo.UnsetDefault(ctx, string(channel.Provider), channel.TenantID); err != nil {
			return nil, fmt.Errorf("failed to unset existing default channel: %w", err)
		}
	}

	// Save to database
	created, err := s.repo.Create(ctx, channel)
	if err != nil {
		return nil, handleEntError(ctx, "Channel", err)
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ChannelEventData{
			ChannelID: created.ID,
			Name:      created.Name,
			Provider:  created.Provider,
			Status:    created.Status,
			IsDefault: created.IsDefault,
			TenantID:  created.TenantID,
			Metadata:  created.Metadata,
		}

		s.publisher.PublishChannelCreated(ctx, eventData)
	}

	return s.Serialize(created), nil
}

// GetByID gets a payment channel by ID
func (s *channelService) GetByID(ctx context.Context, id string) (*structs.Channel, error) {
	if id == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("id"))
	}

	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, handleEntError(ctx, "Channel", err)
	}

	return s.Serialize(channel), nil
}

// Update updates a payment channel
func (s *channelService) Update(ctx context.Context, id string, updates *structs.UpdateChannelInput) (*structs.Channel, error) {
	if id == "" {
		return nil, fmt.Errorf(ecode.FieldIsRequired("id"))
	}

	// Get existing channel
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, handleEntError(ctx, "Channel", err)
	}

	// Apply updates
	if updates.Name != "" {
		existing.Name = updates.Name
	}

	if updates.Status != "" {
		existing.Status = updates.Status
	}

	if updates.IsDefault != nil {
		// If setting as default, unset any existing default
		if *updates.IsDefault && !existing.IsDefault {
			if err := s.repo.UnsetDefault(ctx, string(existing.Provider), existing.TenantID); err != nil {
				return nil, fmt.Errorf("failed to unset existing default channel: %w", err)
			}
		}
		existing.IsDefault = *updates.IsDefault
	}

	if updates.SupportedType != nil {
		existing.SupportedType = updates.SupportedType
	}

	if updates.Config != nil {
		existing.Config = updates.Config
	}

	if updates.Metadata != nil {
		existing.Metadata = updates.Metadata
	}

	existing.UpdatedAt = time.Now().UnixMilli()

	// Save to database
	updated, err := s.repo.Update(ctx, &structs.UpdateChannelInput{
		ID:            existing.ID,
		Name:          existing.Name,
		Status:        existing.Status,
		IsDefault:     &existing.IsDefault,
		SupportedType: existing.SupportedType,
		Config:        existing.Config,
		Metadata:      existing.Metadata,
	})
	if err != nil {
		return nil, handleEntError(ctx, "Channel", err)
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ChannelEventData{
			ChannelID: updated.ID,
			Name:      updated.Name,
			Provider:  updated.Provider,
			Status:    updated.Status,
			IsDefault: updated.IsDefault,
			TenantID:  updated.TenantID,
			Metadata:  updated.Metadata,
		}

		s.publisher.PublishChannelUpdated(ctx, eventData)
	}

	return s.Serialize(updated), nil
}

// Delete deletes a payment channel
func (s *channelService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("channel ID is required")
	}

	// Get existing channel for event data
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if the channel is in use
	inUse, err := s.repo.IsInUse(ctx, id)
	if err != nil {
		return err
	}

	if inUse {
		return fmt.Errorf("channel is in use and cannot be deleted")
	}

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ChannelEventData{
			ChannelID: existing.ID,
			Name:      existing.Name,
			Provider:  existing.Provider,
			Status:    existing.Status,
			IsDefault: existing.IsDefault,
			TenantID:  existing.TenantID,
			Metadata:  existing.Metadata,
		}

		s.publisher.PublishChannelDeleted(ctx, eventData)
	}

	return nil
}

// ChangeStatus changes the status of a payment channel
func (s *channelService) ChangeStatus(ctx context.Context, id string, status structs.ChannelStatus) (*structs.Channel, error) {
	if id == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// Get existing channel
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply status change
	existing.Status = status

	// Save to database
	updated, err := s.repo.Update(ctx, &structs.UpdateChannelInput{
		Status: existing.Status,
	})
	if err != nil {
		return nil, err
	}

	// Publish event
	if s.publisher != nil {
		eventData := &event.ChannelEventData{
			ChannelID: updated.ID,
			Name:      updated.Name,
			Provider:  updated.Provider,
			Status:    updated.Status,
			IsDefault: updated.IsDefault,
			TenantID:  updated.TenantID,
			Metadata:  updated.Metadata,
		}

		if status == structs.ChannelStatusActive {
			s.publisher.PublishChannelActivated(ctx, eventData)
		} else if status == structs.ChannelStatusDisabled {
			s.publisher.PublishChannelDisabled(ctx, eventData)
		} else if status == event.ChannelUpdated {
			s.publisher.PublishChannelUpdated(ctx, eventData)
		}
	}

	return updated, nil
}

// List lists payment channels with pagination
func (s *channelService) List(ctx context.Context, query *structs.ChannelQuery) (paging.Result[*structs.Channel], error) {
	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: "forward", // Default direction
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Channel, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit

		channels, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing channels: %v", err)
			return nil, 0, err
		}

		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting channels: %v", err)
			return nil, 0, err
		}

		return s.Serializes(channels), int(total), nil
	})
}

// GetDefault gets the default payment channel for a provider
func (s *channelService) GetDefault(ctx context.Context, provider structs.PaymentProvider, tenantID string) (*structs.Channel, error) {
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}

	return s.repo.GetDefault(ctx, string(provider), tenantID)
}

// Serialize serializes a channel entity to a response format
func (s *channelService) Serialize(channel *structs.Channel) *structs.Channel {
	if channel == nil {
		return nil
	}

	return &structs.Channel{
		ID:            channel.ID,
		Name:          channel.Name,
		Provider:      channel.Provider,
		Status:        channel.Status,
		IsDefault:     channel.IsDefault,
		SupportedType: channel.SupportedType,
		Config:        channel.Config,
		Metadata:      channel.Metadata,
		TenantID:      channel.TenantID,
		CreatedAt:     channel.CreatedAt,
		UpdatedAt:     channel.UpdatedAt,
	}
}

// Serializes serializes multiple channel entities to response format
func (s *channelService) Serializes(channels []*structs.Channel) []*structs.Channel {
	result := make([]*structs.Channel, len(channels))
	for i, channel := range channels {
		result[i] = s.Serialize(channel)
	}
	return result
}
