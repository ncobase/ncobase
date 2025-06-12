package service

import (
	"context"
	"errors"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/ecode"
)

// UserSpaceServiceInterface is the interface for the service.
type UserSpaceServiceInterface interface {
	UserBelongSpace(ctx context.Context, uid string) (*structs.ReadSpace, error)
	UserBelongSpaces(ctx context.Context, uid string) ([]*structs.ReadSpace, error)
	AddUserToSpace(ctx context.Context, u, t string) (*structs.UserSpace, error)
	RemoveUserFromSpace(ctx context.Context, u, t string) error
	IsSpaceInUser(ctx context.Context, t, u string) (bool, error)
}

// userSpaceService is the struct for the service.
type userSpaceService struct {
	ts        SpaceServiceInterface
	userSpace repository.UserSpaceRepositoryInterface
}

// NewUserSpaceService creates a new service.
func NewUserSpaceService(d *data.Data, ts SpaceServiceInterface) UserSpaceServiceInterface {
	return &userSpaceService{
		ts:        ts,
		userSpace: repository.NewUserSpaceRepository(d),
	}
}

// UserBelongSpace user belong space service
func (s *userSpaceService) UserBelongSpace(ctx context.Context, uid string) (*structs.ReadSpace, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	// Try to get space from user-space relationship
	userSpace, err := s.userSpace.GetByUserID(ctx, uid)
	if err != nil {
		// If no specific space found, try to get the first available space for the user
		spaces, err := s.userSpace.GetSpacesByUserID(ctx, uid)
		if err != nil || len(spaces) == 0 {
			// If user doesn't belong to any space, check if they created a space
			return s.ts.GetByUser(ctx, uid)
		}
		// Return the first space
		return s.ts.Serialize(spaces[0]), nil
	}

	row, err := s.ts.Find(ctx, userSpace.SpaceID)
	if err := handleEntError(ctx, "Space", err); err != nil {
		return nil, err
	}

	return row, nil
}

// UserBelongSpaces user belong spaces service
func (s *userSpaceService) UserBelongSpaces(ctx context.Context, uid string) ([]*structs.ReadSpace, error) {
	if uid == "" {
		return nil, errors.New(ecode.FieldIsInvalid("User ID"))
	}

	userSpaces, err := s.userSpace.GetSpacesByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	var spaces []*structs.ReadSpace
	for _, userSpace := range userSpaces {
		space, err := s.ts.Find(ctx, userSpace.ID)
		if err != nil {
			return nil, errors.New("space not found")
		}
		spaces = append(spaces, space)
	}

	return spaces, nil
}

// AddUserToSpace adds a user to a space.
func (s *userSpaceService) AddUserToSpace(ctx context.Context, u string, t string) (*structs.UserSpace, error) {
	row, err := s.userSpace.Create(ctx, &structs.UserSpace{UserID: u, SpaceID: t})
	if err := handleEntError(ctx, "UserSpace", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// RemoveUserFromSpace removes a user from a space.
func (s *userSpaceService) RemoveUserFromSpace(ctx context.Context, u string, t string) error {
	err := s.userSpace.Delete(ctx, u, t)
	if err := handleEntError(ctx, "UserSpace", err); err != nil {
		return err
	}
	return nil
}

// IsSpaceInUser checks if a space is in a user.
func (s *userSpaceService) IsSpaceInUser(ctx context.Context, t, u string) (bool, error) {
	isValid, err := s.userSpace.IsSpaceInUser(ctx, t, u)
	if err = handleEntError(ctx, "UserSpace", err); err != nil {
		return false, err

	}
	return isValid, nil
}

// Serializes serializes user spaces.
func (s *userSpaceService) Serializes(rows []*ent.UserSpace) []*structs.UserSpace {
	rs := make([]*structs.UserSpace, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a user space.
func (s *userSpaceService) Serialize(row *ent.UserSpace) *structs.UserSpace {
	return &structs.UserSpace{
		UserID:  row.UserID,
		SpaceID: row.SpaceID,
	}
}
