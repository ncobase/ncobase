package service

import (
	"context"
	"errors"
	"ncobase/biz/content/data"
	"ncobase/biz/content/data/repository"
	"ncobase/biz/content/structs"
	"ncobase/biz/content/wrapper"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// MediaServiceInterface for media service operations
type MediaServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateMediaBody) (*structs.ReadMedia, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadMedia, error)
	Get(ctx context.Context, id string) (*structs.ReadMedia, error)
	List(ctx context.Context, params *structs.ListMediaParams) (paging.Result[*structs.ReadMedia], error)
	Delete(ctx context.Context, id string) error
}

type mediaService struct {
	r   repository.MediaRepositoryInterface
	rsw *wrapper.ResourceServiceWrapper
}

// NewMediaService creates new media service
func NewMediaService(d *data.Data, rsw *wrapper.ResourceServiceWrapper) MediaServiceInterface {
	return &mediaService{
		r:   repository.NewMediaRepository(d),
		rsw: rsw,
	}
}

// Create creates new media
func (s *mediaService) Create(ctx context.Context, body *structs.CreateMediaBody) (*structs.ReadMedia, error) {
	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	if validator.IsEmpty(body.SpaceID) {
		return nil, errors.New(ecode.FieldIsRequired("space_id"))
	}

	if validator.IsEmpty(body.OwnerID) {
		return nil, errors.New(ecode.FieldIsRequired("owner_id"))
	}

	// Validate that either resource_id or url is provided
	if validator.IsEmpty(body.ResourceID) && validator.IsEmpty(body.URL) {
		return nil, errors.New("either resource_id or url must be provided")
	}

	// Validate resource_id if provided
	if validator.IsNotEmpty(body.ResourceID) && s.rsw.HasFileService() {
		_, err := s.rsw.GetFile(ctx, body.ResourceID)
		if err != nil {
			return nil, errors.New("invalid resource_id: resource not found")
		}
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.enrichMedia(ctx, repository.SerializeMedia(row)), nil
}

// Update updates existing media
func (s *mediaService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadMedia, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Validate resource_id if being updated
	if resourceID, ok := updates["resource_id"].(string); ok && validator.IsNotEmpty(resourceID) {
		if s.rsw.HasFileService() {
			_, err := s.rsw.GetFile(ctx, resourceID)
			if err != nil {
				return nil, errors.New("invalid resource_id: resource not found")
			}
		}
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.enrichMedia(ctx, repository.SerializeMedia(row)), nil
}

// Get retrieves media by ID
func (s *mediaService) Get(ctx context.Context, id string) (*structs.ReadMedia, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.enrichMedia(ctx, repository.SerializeMedia(row)), nil
}

// Delete deletes media by ID
func (s *mediaService) Delete(ctx context.Context, id string) error {
	err := s.r.Delete(ctx, id)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return err
	}

	return nil
}

// List lists all media
func (s *mediaService) List(ctx context.Context, params *structs.ListMediaParams) (paging.Result[*structs.ReadMedia], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadMedia, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, count, err := s.r.ListWithCount(ctx, &lp)
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing media: %v", err)
			return nil, 0, err
		}

		return s.enrichMedias(ctx, repository.SerializeMedias(rows)), count, nil
	})
}

// enrichMedias enriches media rows with related data.
func (s *mediaService) enrichMedias(ctx context.Context, rows []*structs.ReadMedia) []*structs.ReadMedia {
	rs := make([]*structs.ReadMedia, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.enrichMedia(ctx, row))
	}
	return rs
}

// enrichMedia enriches a media row with related resource data.
func (s *mediaService) enrichMedia(ctx context.Context, media *structs.ReadMedia) *structs.ReadMedia {
	if media == nil {
		return nil
	}
	if media.ResourceID != "" && s.rsw != nil && s.rsw.HasFileService() {
		if resource, err := s.rsw.GetFile(ctx, media.ResourceID); err == nil {
			media.Resource = &structs.ResourceFileReference{
				ID:           resource.ID,
				Name:         resource.Name,
				Path:         resource.Path,
				Type:         resource.Type,
				Size:         resource.Size,
				Storage:      resource.Storage,
				DownloadURL:  resource.DownloadURL,
				ThumbnailURL: resource.ThumbnailURL,
				IsExpired:    resource.IsExpired,
			}
		} else {
			logger.Warnf(ctx, "Failed to load resource %s: %v", media.ResourceID, err)
		}
	}
	return media
}
