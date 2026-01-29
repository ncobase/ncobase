package service

import (
	"context"
	"errors"
	"ncobase/biz/content/data"
	"ncobase/biz/content/data/repository"
	"ncobase/biz/content/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/slug"
	"github.com/ncobase/ncore/validation/validator"
)

// ChannelServiceInterface is the interface for the service.
type ChannelServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateChannelBody) (*structs.ReadChannel, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadChannel, error)
	Get(ctx context.Context, slug string) (*structs.ReadChannel, error)
	List(ctx context.Context, params *structs.ListChannelParams) (paging.Result[*structs.ReadChannel], error)
	Delete(ctx context.Context, slug string) error
}

// channelService is the struct for the service.
type channelService struct {
	r repository.ChannelRepositoryInterface
}

// NewChannelService creates a new service.
func NewChannelService(d *data.Data) ChannelServiceInterface {
	return &channelService{
		r: repository.NewChannelRepository(d),
	}
}

// Create creates a new channel.
func (s *channelService) Create(ctx context.Context, body *structs.CreateChannelBody) (*structs.ReadChannel, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsRequired("name"))
	}
	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}
	// Set slug field if empty
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Channel", err); err != nil {
		return nil, err
	}

	return repository.SerializeChannel(row), nil
}

// Update updates an existing channel.
func (s *channelService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadChannel, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug / id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.r.Update(ctx, slug, updates)
	if err := handleEntError(ctx, "Channel", err); err != nil {
		return nil, err
	}

	return repository.SerializeChannel(row), nil
}

// Get retrieves a channel by slug or ID.
func (s *channelService) Get(ctx context.Context, slug string) (*structs.ReadChannel, error) {
	row, err := s.r.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Channel", err); err != nil {
		return nil, err
	}

	return repository.SerializeChannel(row), nil
}

// Delete deletes a channel by slug or ID.
func (s *channelService) Delete(ctx context.Context, slug string) error {
	err := s.r.Delete(ctx, slug)
	if err := handleEntError(ctx, "Channel", err); err != nil {
		return err
	}

	return nil
}

// List lists all channels.
func (s *channelService) List(ctx context.Context, params *structs.ListChannelParams) (paging.Result[*structs.ReadChannel], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadChannel, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, count, err := s.r.ListWithCount(ctx, &lp)
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing channels: %v", err)
			return nil, 0, err
		}

		return repository.SerializeChannels(rows), count, nil
	})
}
