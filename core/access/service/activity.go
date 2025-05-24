package service

import (
	"context"
	"errors"
	"ncobase/access/data"
	"ncobase/access/data/ent"
	"ncobase/access/data/repository"
	"ncobase/access/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
)

// ActivityServiceInterface defines service operations for activity logs
type ActivityServiceInterface interface {
	LogActivity(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.ActivityEntry, error)
	GetActivity(ctx context.Context, id string) (*structs.ActivityEntry, error)
	ListActivity(ctx context.Context, params *structs.ListActivityParams) ([]*structs.ActivityEntry, int, error)
	GetUserActivity(ctx context.Context, username string, limit int) ([]*structs.ActivityEntry, error)
}

// activityService implements ActivityServiceInterface
type activityService struct {
	activity repository.ActivityRepositoryInterface
}

// NewActivityService creates a new activity log service
func NewActivityService(d *data.Data) ActivityServiceInterface {
	return &activityService{
		activity: repository.NewActivityRepository(d),
	}
}

// LogActivity records a new activity
func (s *activityService) LogActivity(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.ActivityEntry, error) {
	row, err := s.activity.Create(ctx, userID, log)
	if err != nil {
		logger.Errorf(ctx, "activityService.LogActivity error: %v", err)
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetActivity retrieves an activity log entry
func (s *activityService) GetActivity(ctx context.Context, id string) (*structs.ActivityEntry, error) {
	row, err := s.activity.GetByID(ctx, id)
	if err = handleEntError(ctx, "Activity", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// ListActivity lists activity logs with filters
func (s *activityService) ListActivity(ctx context.Context, params *structs.ListActivityParams) ([]*structs.ActivityEntry, int, error) {
	rows, total, err := s.activity.List(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "activityService.ListActivity error: %v", err)
		return nil, 0, err
	}

	return s.Serializes(rows), total, nil
}

// GetUserActivity retrieves activity logs for a user
func (s *activityService) GetUserActivity(ctx context.Context, username string, limit int) ([]*structs.ActivityEntry, error) {
	currentUserID := ctxutil.GetUserID(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)
	if currentUserID != username && !isAdmin {
		return nil, errors.New("you don't have permission to view this user's activity logs")
	}
	rows, err := s.activity.GetRecentByUserID(ctx, username, limit)
	if err != nil {
		logger.Errorf(ctx, "activityService.GetUserActivity error: %v", err)
		return nil, err
	}

	return s.Serializes(rows), nil
}

// Serializes converts multiple activities to DTOs
func (s *activityService) Serializes(rows []*ent.Activity) []*structs.ActivityEntry {
	var rs []*structs.ActivityEntry
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize converts an activity to DTO
func (s *activityService) Serialize(row *ent.Activity) *structs.ActivityEntry {
	return &structs.ActivityEntry{
		ID:        row.ID,
		UserID:    row.UserID,
		Type:      row.Type,
		Timestamp: row.CreatedAt,
		Details:   row.Details,
		Metadata:  &row.Metadata,
	}
}
