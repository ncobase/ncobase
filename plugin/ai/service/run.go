package service

import (
	"context"
	"errors"
	"ncobase/plugin/ai/data/repository"
	"ncobase/plugin/ai/structs"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
)

type RunServiceInterface interface {
	GetByID(ctx context.Context, id string, unrestricted bool) (*structs.Run, error)
	List(ctx context.Context, query *structs.RunQuery, unrestricted bool) (paging.Result[*structs.Run], error)
	Usage(ctx context.Context, query *structs.UsageQuery, unrestricted bool) (*structs.UsageSummary, error)
}

type runService struct {
	repo repository.RunRepositoryInterface
}

func NewRunService(repo repository.RunRepositoryInterface) RunServiceInterface {
	return &runService{repo: repo}
}

func (s *runService) GetByID(ctx context.Context, id string, unrestricted bool) (*structs.Run, error) {
	if id == "" {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !unrestricted && run.UserID != "" && run.UserID != ctxutil.GetUserID(ctx) {
		return nil, errors.New(ecode.Text(ecode.AccessDenied))
	}
	return run, nil
}

func (s *runService) List(ctx context.Context, query *structs.RunQuery, unrestricted bool) (paging.Result[*structs.Run], error) {
	if query == nil {
		query = &structs.RunQuery{}
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}
	if !unrestricted {
		query.UserID = ctxutil.GetUserID(ctx)
	}

	pp := paging.Params{
		Cursor:    query.Cursor,
		Limit:     query.PageSize,
		Direction: query.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.Run, int, error) {
		lq := *query
		lq.Cursor = cursor
		lq.PageSize = limit
		lq.Direction = direction

		runs, err := s.repo.List(ctx, &lq)
		if err != nil {
			logger.Errorf(ctx, "Error listing AI runs: %v", err)
			return nil, 0, err
		}
		total, err := s.repo.Count(ctx, query)
		if err != nil {
			logger.Errorf(ctx, "Error counting AI runs: %v", err)
			return nil, 0, err
		}
		return runs, int(total), nil
	})
}

func (s *runService) Usage(ctx context.Context, query *structs.UsageQuery, unrestricted bool) (*structs.UsageSummary, error) {
	if query == nil {
		query = &structs.UsageQuery{}
	}
	if !unrestricted {
		query.UserID = ctxutil.GetUserID(ctx)
	}
	return s.repo.Usage(ctx, query)
}
