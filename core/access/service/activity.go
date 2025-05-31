package service

import (
	"context"
	"errors"
	"ncobase/access/data"
	"ncobase/access/data/repository"
	"ncobase/access/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

// ActivityServiceInterface defines service operations for activity
type ActivityServiceInterface interface {
	LogActivity(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.Activity, error)
	GetActivity(ctx context.Context, id string) (*structs.Activity, error)
	ListActivity(ctx context.Context, params *structs.ListActivityParams) (paging.Result[*structs.Activity], error)
	GetUserActivity(ctx context.Context, username string, limit int) ([]*structs.Activity, error)
	SearchActivity(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.Activity, int, error)
	CountX(ctx context.Context, params *structs.ListActivityParams) int
	DocumentToEntry(doc *structs.ActivityDocument) *structs.Activity
	DocumentsToEntries(docs []*structs.ActivityDocument) []*structs.Activity
}

type activityService struct {
	activity repository.ActivityRepositoryInterface
}

func NewActivityService(d *data.Data) ActivityServiceInterface {
	return &activityService{
		activity: repository.NewActivityRepository(d),
	}
}

func (s *activityService) LogActivity(ctx context.Context, userID string, log *structs.CreateActivityRequest) (*structs.Activity, error) {
	if log.Type == "" {
		return nil, errors.New("activity type is required")
	}
	if log.Details == "" {
		return nil, errors.New("activity details is required")
	}

	doc, err := s.activity.Create(ctx, userID, log)
	if err != nil {
		logger.Errorf(ctx, "activityService.LogActivity error: %v", err)
		return nil, err
	}

	return s.DocumentToEntry(doc), nil
}

func (s *activityService) GetActivity(ctx context.Context, id string) (*structs.Activity, error) {
	doc, err := s.activity.GetByID(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "activityService.GetActivity error: %v", err)
		return nil, err
	}

	return s.DocumentToEntry(doc), nil
}

func (s *activityService) ListActivity(ctx context.Context, params *structs.ListActivityParams) (paging.Result[*structs.Activity], error) {
	result, err := s.activity.List(ctx, params)
	if err != nil {
		if errors.Is(err, paging.ErrInvalidCursor) {
			return paging.Result[*structs.Activity]{}, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		logger.Errorf(ctx, "activityService.ListActivity error: %v", err)
		return paging.Result[*structs.Activity]{}, err
	}

	// Convert documents to entries
	entries := s.DocumentsToEntries(result.Items)

	return paging.Result[*structs.Activity]{
		Items:       entries,
		Total:       result.Total,
		NextCursor:  result.NextCursor,
		PrevCursor:  result.PrevCursor,
		HasNextPage: result.HasNextPage,
		HasPrevPage: result.HasPrevPage,
	}, nil
}

func (s *activityService) GetUserActivity(ctx context.Context, username string, limit int) ([]*structs.Activity, error) {
	currentUserID := ctxutil.GetUserID(ctx)
	isAdmin := ctxutil.GetUserIsAdmin(ctx)

	if currentUserID != username && !isAdmin {
		return nil, errors.New("you don't have permission to view this user's activities")
	}

	docs, err := s.activity.GetRecentByUserID(ctx, username, limit)
	if err != nil {
		logger.Errorf(ctx, "activityService.GetUserActivity error: %v", err)
		return nil, err
	}

	return s.DocumentsToEntries(docs), nil
}

func (s *activityService) SearchActivity(ctx context.Context, params *structs.SearchActivityParams) ([]*structs.Activity, int, error) {
	docs, total, err := s.activity.Search(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "activityService.SearchActivity error: %v", err)
		return nil, 0, err
	}

	return s.DocumentsToEntries(docs), total, nil
}

func (s *activityService) CountX(ctx context.Context, params *structs.ListActivityParams) int {
	return s.activity.CountX(ctx, params)
}

func (s *activityService) DocumentsToEntries(docs []*structs.ActivityDocument) []*structs.Activity {
	entries := make([]*structs.Activity, len(docs))
	for i, doc := range docs {
		entries[i] = s.DocumentToEntry(doc)
	}
	return entries
}

func (s *activityService) DocumentToEntry(doc *structs.ActivityDocument) *structs.Activity {
	return &structs.Activity{
		ID:        doc.ID,
		UserID:    doc.UserID,
		Type:      doc.Type,
		Timestamp: doc.CreatedAt,
		Details:   doc.Details,
		Metadata:  doc.Metadata,
	}
}
