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
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// MediaServiceInterface is the interface for the service.
type MediaServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateMediaBody) (*structs.ReadMedia, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadMedia, error)
	Get(ctx context.Context, id string) (*structs.ReadMedia, error)
	List(ctx context.Context, params *structs.ListMediaParams) (paging.Result[*structs.ReadMedia], error)
	Delete(ctx context.Context, id string) error
}

// mediaService is the struct for the service.
type mediaService struct {
	r repository.MediaRepositoryInterface
}

// NewMediaService creates a new service.
func NewMediaService(d *data.Data) MediaServiceInterface {
	return &mediaService{
		r: repository.NewMediaRepository(d),
	}
}

// Create creates a new media.
func (s *mediaService) Create(ctx context.Context, body *structs.CreateMediaBody) (*structs.ReadMedia, error) {
	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	// Validate media type
	validTypes := map[string]bool{
		structs.MediaTypeImage: true,
		structs.MediaTypeVideo: true,
		structs.MediaTypeAudio: true,
		structs.MediaTypeFile:  true,
	}

	if !validTypes[body.Type] {
		return nil, errors.New(ecode.FieldIsInvalid("type"))
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing media.
func (s *mediaService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadMedia, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Validate media type if updating
	if mediaType, ok := updates["type"].(string); ok {
		validTypes := map[string]bool{
			structs.MediaTypeImage: true,
			structs.MediaTypeVideo: true,
			structs.MediaTypeAudio: true,
			structs.MediaTypeFile:  true,
		}

		if !validTypes[mediaType] {
			return nil, errors.New(ecode.FieldIsInvalid("type"))
		}
	}

	row, err := s.r.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a media by ID.
func (s *mediaService) Get(ctx context.Context, id string) (*structs.ReadMedia, error) {
	row, err := s.r.GetByID(ctx, id)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a media by ID.
func (s *mediaService) Delete(ctx context.Context, id string) error {
	err := s.r.Delete(ctx, id)
	if err := handleEntError(ctx, "Media", err); err != nil {
		return err
	}

	return nil
}

// List lists all media.
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
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing media: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), count, nil
	})
}

// Serializes converts multiple ent.Media to []*structs.ReadMedia.
func (s *mediaService) Serializes(rows []*ent.Media) []*structs.ReadMedia {
	var rs []*structs.ReadMedia
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize converts an ent.Media to a structs.ReadMedia.
func (s *mediaService) Serialize(row *ent.Media) *structs.ReadMedia {
	return &structs.ReadMedia{
		ID:          row.ID,
		Title:       row.Title,
		Type:        row.Type,
		URL:         row.URL,
		Path:        row.Path,
		MimeType:    row.MimeType,
		Size:        row.Size,
		Width:       row.Width,
		Height:      row.Height,
		Duration:    row.Duration,
		Description: row.Description,
		Alt:         row.Alt,
		Metadata:    &row.Extras,
		TenantID:    row.TenantID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
