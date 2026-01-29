package service

import (
	"context"
	"errors"
	"ncobase/plugin/counter/data"
	"ncobase/plugin/counter/data/repository"
	"ncobase/plugin/counter/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// CounterServiceInterface is the interface for the service.
type CounterServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateCounterBody) (*structs.ReadCounter, error)
	Update(ctx context.Context, counterID string, updates types.JSON) (*structs.ReadCounter, error)
	Get(ctx context.Context, params *structs.FindCounter) (*structs.ReadCounter, error)
	GetByIDs(ctx context.Context, counterIDs []string) ([]*structs.ReadCounter, error)
	Delete(ctx context.Context, counterID string) error
	List(ctx context.Context, params *structs.ListCounterParams) (paging.Result[*structs.ReadCounter], error)
	CountX(ctx context.Context, params *structs.ListCounterParams) int
}

// counterService is the struct for the service.
type counterService struct {
	counter repository.CounterRepositoryInterface
}

// NewCounterService creates a new service.
func NewCounterService(d *data.Data) CounterServiceInterface {
	return &counterService{
		counter: repository.NewCounterRepository(d),
	}
}

// Create creates a new counter.
func (s *counterService) Create(ctx context.Context, body *structs.CreateCounterBody) (*structs.ReadCounter, error) {
	if body.Name == "" {
		return nil, errors.New("counter name is required")
	}

	row, err := s.counter.Create(ctx, body)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return repository.SerializeCounter(row), nil
}

// Update updates an existing counter.
func (s *counterService) Update(ctx context.Context, counterID string, updates types.JSON) (*structs.ReadCounter, error) {
	row, err := s.counter.Update(ctx, counterID, updates)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return repository.SerializeCounter(row), nil
}

// Get retrieves a counter by its ID.
func (s *counterService) Get(ctx context.Context, params *structs.FindCounter) (*structs.ReadCounter, error) {
	row, err := s.counter.GetByID(ctx, params.Counter)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return repository.SerializeCounter(row), nil
}

// GetByIDs retrieves counters by their IDs.
func (s *counterService) GetByIDs(ctx context.Context, counterIDs []string) ([]*structs.ReadCounter, error) {
	rows, err := s.counter.GetByIDs(ctx, counterIDs)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return repository.SerializeCounters(rows), nil
}

// Delete deletes a counter by its ID.
func (s *counterService) Delete(ctx context.Context, counterID string) error {
	err := s.counter.Delete(ctx, counterID)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return err
	}

	return nil
}

// List lists all counters.
func (s *counterService) List(ctx context.Context, params *structs.ListCounterParams) (paging.Result[*structs.ReadCounter], error) {

	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadCounter, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.counter.List(ctx, &lp)
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing counters: %v", err)
			return nil, 0, err
		}

		total := s.counter.CountX(ctx, params)

		return repository.SerializeCounters(rows), total, nil
	})
}

// CountX gets a count of counters.
func (s *counterService) CountX(ctx context.Context, params *structs.ListCounterParams) int {
	return s.counter.CountX(ctx, params)
}
