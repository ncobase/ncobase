package service

import (
	"context"
	"errors"
	"ncobase/counter/data"
	"ncobase/counter/data/ent"
	"ncobase/counter/data/repository"
	"ncobase/counter/structs"

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
	Serializes(rows []*ent.Counter) []*structs.ReadCounter
	Serialize(counter *ent.Counter) *structs.ReadCounter
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

	return s.Serialize(row), nil
}

// Update updates an existing counter.
func (s *counterService) Update(ctx context.Context, counterID string, updates types.JSON) (*structs.ReadCounter, error) {
	row, err := s.counter.Update(ctx, counterID, updates)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a counter by its ID.
func (s *counterService) Get(ctx context.Context, params *structs.FindCounter) (*structs.ReadCounter, error) {
	row, err := s.counter.GetByID(ctx, params.Counter)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByIDs retrieves counters by their IDs.
func (s *counterService) GetByIDs(ctx context.Context, counterIDs []string) ([]*structs.ReadCounter, error) {
	rows, err := s.counter.GetByIDs(ctx, counterIDs)
	if err := handleEntError(ctx, "Counter", err); err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
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
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing counters: %v", err)
			return nil, 0, err
		}

		total := s.counter.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes counters.
func (s *counterService) Serializes(rows []*ent.Counter) []*structs.ReadCounter {
	var rs []*structs.ReadCounter
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a counter.
func (s *counterService) Serialize(row *ent.Counter) *structs.ReadCounter {
	return &structs.ReadCounter{
		ID:            row.ID,
		Identifier:    row.Identifier,
		Name:          row.Name,
		Prefix:        row.Prefix,
		Suffix:        row.Suffix,
		StartValue:    row.StartValue,
		IncrementStep: row.IncrementStep,
		DateFormat:    row.DateFormat,
		CurrentValue:  row.CurrentValue,
		Disabled:      row.Disabled,
		Description:   row.Description,
		TenantID:      &row.TenantID,
		CreatedBy:     &row.CreatedBy,
		CreatedAt:     &row.CreatedAt,
		UpdatedBy:     &row.UpdatedBy,
		UpdatedAt:     &row.UpdatedAt,
	}
}

// CountX gets a count of counters.
func (s *counterService) CountX(ctx context.Context, params *structs.ListCounterParams) int {
	return s.counter.CountX(ctx, params)
}
