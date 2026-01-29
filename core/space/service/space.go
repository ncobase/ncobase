package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
)

// SpaceServiceInterface is the interface for the service.
type SpaceServiceInterface interface {
	UserOwn(ctx context.Context, uid string) (*structs.ReadSpace, error)
	Create(ctx context.Context, body *structs.CreateSpaceBody) (*structs.ReadSpace, error)
	Update(ctx context.Context, body *structs.UpdateSpaceBody) (*structs.ReadSpace, error)
	Get(ctx context.Context, id string) (*structs.ReadSpace, error)
	GetBySlug(ctx context.Context, id string) (*structs.ReadSpace, error)
	GetByUser(ctx context.Context, uid string) (*structs.ReadSpace, error)
	GetByIDs(ctx context.Context, ids []string) ([]*structs.ReadSpace, error)
	Find(ctx context.Context, id string) (*structs.ReadSpace, error)
	Delete(ctx context.Context, id string) error
	CountX(ctx context.Context, params *structs.ListSpaceParams) int
	List(ctx context.Context, params *structs.ListSpaceParams) (paging.Result[*structs.ReadSpace], error)
}

// spaceService is the struct for the service.
type spaceService struct {
	space     repository.SpaceRepositoryInterface
	userSpace repository.UserSpaceRepositoryInterface
}

// NewSpaceService creates a new service.
func NewSpaceService(d *data.Data) SpaceServiceInterface {
	return &spaceService{
		space:     repository.NewSpaceRepository(d),
		userSpace: repository.NewUserSpaceRepository(d),
	}
}

// UserOwn user own space service
func (s *spaceService) UserOwn(ctx context.Context, uid string) (*structs.ReadSpace, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	row, err := s.space.GetByUser(ctx, uid)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpace(row), nil
}

// Create creates a space service.
func (s *spaceService) Create(ctx context.Context, body *structs.CreateSpaceBody) (*structs.ReadSpace, error) {
	if body.CreatedBy == nil {
		body.CreatedBy = convert.ToPointer(ctxutil.GetUserID(ctx))
	}

	// Create the space
	space, err := s.space.Create(ctx, body)
	if err != nil {
		return nil, err
	}

	return repository.SerializeSpace(space), nil
}

// Update updates space service (full and partial).
func (s *spaceService) Update(ctx context.Context, body *structs.UpdateSpaceBody) (*structs.ReadSpace, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// Check if CreatedBy field is provided and validate user's access to the space
	if body.CreatedBy != nil {
		_, err := s.space.GetByUser(ctx, *body.CreatedBy)
		if err := handleEntError(ctx, "Space", err); err != nil {
			return nil, err
		}
	}

	// If ID is not provided, get the space ID associated with the user
	if body.ID == "" {
		body.ID, _ = s.space.GetIDByUser(ctx, userID)
	}

	// Retrieve the space by ID
	row, err := s.Find(ctx, body.ID)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator or user belongs to the space
	ut, err := s.userSpace.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "UserSpace", err); err != nil {
		return nil, err
	}
	if convert.ToValue(row.CreatedBy) != userID && ut.SpaceID != row.ID {
		return nil, errors.New("this space is not yours or your not belong to this space")
	}

	// set updated by
	body.UpdatedBy = &userID

	// Serialize request body
	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a types.JSON object
	var d types.JSON
	if err := json.Unmarshal(bodyData, &d); err != nil {
		return nil, err
	}

	// Update the space with the provided data
	_, err = s.space.Update(ctx, row.ID, d)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	return row, nil
}

// Get reads space service.
func (s *spaceService) Get(ctx context.Context, id string) (*structs.ReadSpace, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("invalid user ID")
	}

	// If ID is not provided, get the space ID associated with the user
	if id == "" {
		id, _ = s.space.GetIDByUser(ctx, userID)
	}

	// Retrieve the space by ID
	row, err := s.Find(ctx, id)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	// Check if the user is the creator or user belongs to the space
	ut, err := s.userSpace.GetByUserID(ctx, userID)
	if err := handleEntError(ctx, "UserSpace", err); err != nil {
		return nil, err
	}
	if convert.ToValue(row.CreatedBy) != userID && ut.SpaceID != row.ID {
		return nil, errors.New("this space is not yours or your not belong to this space")
	}

	return row, nil
}

// GetBySlug returns the space for the provided slug
func (s *spaceService) GetBySlug(ctx context.Context, slug string) (*structs.ReadSpace, error) {
	if slug == "" {
		return nil, errors.New(ecode.FieldIsInvalid("Slug"))
	}
	space, err := s.space.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}
	return repository.SerializeSpace(space), nil
}

// GetByUser returns the space for the created by user
func (s *spaceService) GetByUser(ctx context.Context, uid string) (*structs.ReadSpace, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}
	space, err := s.space.GetByUser(ctx, uid)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}
	return repository.SerializeSpace(space), nil
}

// GetByIDs retrieves multiple spaces by their IDs
func (s *spaceService) GetByIDs(ctx context.Context, ids []string) ([]*structs.ReadSpace, error) {
	if len(ids) == 0 {
		return []*structs.ReadSpace{}, nil
	}

	spaces, err := s.space.GetByIDs(ctx, ids)
	if err != nil {
		logger.Errorf(ctx, "Failed to get spaces by IDs: %v", err)
		return nil, err
	}

	return repository.SerializeSpaces(spaces), nil
}

// Find finds space service.
func (s *spaceService) Find(ctx context.Context, id string) (*structs.ReadSpace, error) {
	space, err := s.space.GetBySlug(ctx, id)
	if err != nil {
		return nil, err
	}
	return repository.SerializeSpace(space), nil
}

// Delete deletes space service.
func (s *spaceService) Delete(ctx context.Context, id string) error {
	err := s.space.Delete(ctx, id)
	if err != nil {
		return err
	}

	// TODO: Freed all roles / orgs / users that are associated with the space

	return nil
}

// List lists space service.
func (s *spaceService) List(ctx context.Context, params *structs.ListSpaceParams) (paging.Result[*structs.ReadSpace], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadSpace, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.space.List(ctx, &lp)
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing spaces: %v", err)
			return nil, 0, err
		}

		total := s.CountX(ctx, params)

		return repository.SerializeSpaces(rows), total, nil
	})
}

// CountX gets a count of spaces.
func (s *spaceService) CountX(ctx context.Context, params *structs.ListSpaceParams) int {
	return s.space.CountX(ctx, params)
}
