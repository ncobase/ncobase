package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ncobase/plugin/counter/data/ent"
	"ncobase/plugin/counter/data/repository"
	"ncobase/plugin/counter/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/types"
)

type mockCounterRepository struct {
	createFn   func(ctx context.Context, body *structs.CreateCounterBody) (*ent.Counter, error)
	getByIDFn  func(ctx context.Context, id string) (*ent.Counter, error)
	getByIDsFn func(ctx context.Context, ids []string) ([]*ent.Counter, error)
	updateFn   func(ctx context.Context, id string, updates types.JSON) (*ent.Counter, error)
	listFn     func(ctx context.Context, params *structs.ListCounterParams) ([]*ent.Counter, error)
	deleteFn   func(ctx context.Context, id string) error
	countXFn   func(ctx context.Context, params *structs.ListCounterParams) int
}

func (m *mockCounterRepository) Create(ctx context.Context, body *structs.CreateCounterBody) (*ent.Counter, error) {
	if m.createFn != nil {
		return m.createFn(ctx, body)
	}
	return nil, nil
}

func (m *mockCounterRepository) GetByID(ctx context.Context, id string) (*ent.Counter, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockCounterRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Counter, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockCounterRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Counter, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, updates)
	}
	return nil, nil
}

func (m *mockCounterRepository) List(ctx context.Context, params *structs.ListCounterParams) ([]*ent.Counter, error) {
	if m.listFn != nil {
		return m.listFn(ctx, params)
	}
	return nil, nil
}

func (m *mockCounterRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockCounterRepository) FindCounter(context.Context, *structs.FindCounter) (*ent.Counter, error) {
	return nil, nil
}

func (m *mockCounterRepository) ListBuilder(context.Context, *structs.ListCounterParams) (*ent.CounterQuery, error) {
	return nil, nil
}

func (m *mockCounterRepository) CountX(ctx context.Context, params *structs.ListCounterParams) int {
	if m.countXFn != nil {
		return m.countXFn(ctx, params)
	}
	return 0
}

func testCounterRow(name string) *ent.Counter {
	return &ent.Counter{
		ID:            "counter-1",
		Identifier:    "COUNTER",
		Name:          name,
		Prefix:        "PX",
		Suffix:        "SX",
		StartValue:    1,
		IncrementStep: 1,
		DateFormat:    "20060102",
		CurrentValue:  10,
		Disabled:      false,
		Description:   "desc",
		SpaceID:       "space-1",
		CreatedBy:     "user-1",
		UpdatedBy:     "user-1",
		CreatedAt:     100,
		UpdatedAt:     100,
	}
}

func TestCounterServiceCreateValidation(t *testing.T) {
	svc := &counterService{}

	if _, err := svc.Create(context.Background(), nil); err == nil {
		t.Fatal("expected error when create body is nil")
	}

	_, err := svc.Create(context.Background(), &structs.CreateCounterBody{CounterBody: structs.CounterBody{Name: "   "}})
	if err == nil {
		t.Fatal("expected error when counter name is blank")
	}
}

func TestCounterServiceCreateTrimsName(t *testing.T) {
	var capturedName string
	svc := &counterService{
		counter: &mockCounterRepository{
			createFn: func(_ context.Context, body *structs.CreateCounterBody) (*ent.Counter, error) {
				capturedName = body.Name
				return testCounterRow(body.Name), nil
			},
		},
	}

	result, err := svc.Create(context.Background(), &structs.CreateCounterBody{
		CounterBody: structs.CounterBody{Name: "  counter name  "},
	})
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if capturedName != "counter name" {
		t.Fatalf("expected trimmed name to be persisted, got %q", capturedName)
	}
	if result == nil || result.Name != "counter name" {
		t.Fatalf("unexpected create result: %#v", result)
	}
}

func TestCounterServiceIDAndQueryValidation(t *testing.T) {
	svc := &counterService{}

	if _, err := svc.Update(context.Background(), "   ", types.JSON{}); err == nil {
		t.Fatal("expected update error for empty counter id")
	}
	if err := svc.Delete(context.Background(), "   "); err == nil {
		t.Fatal("expected delete error for empty counter id")
	}

	if _, err := svc.Get(context.Background(), nil); err == nil {
		t.Fatal("expected get error for nil query")
	}
	if _, err := svc.Get(context.Background(), &structs.FindCounter{Counter: "   "}); err == nil {
		t.Fatal("expected get error for empty counter id in query")
	}

	if _, err := svc.List(context.Background(), nil); err == nil {
		t.Fatal("expected list error for nil params")
	}
}

func TestCounterServiceGetByIDsAndCountValidation(t *testing.T) {
	svc := &counterService{
		counter: &mockCounterRepository{
			getByIDsFn: func(_ context.Context, _ []string) ([]*ent.Counter, error) {
				t.Fatal("repository GetByIDs should not be called for empty input")
				return nil, nil
			},
		},
	}

	rows, err := svc.GetByIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetByIDs returned unexpected error: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected empty result for empty ids, got %#v", rows)
	}

	if got := svc.CountX(context.Background(), nil); got != 0 {
		t.Fatalf("expected zero CountX for nil params, got %d", got)
	}
}

func TestCounterServiceListNotFoundCursorMapping(t *testing.T) {
	svc := &counterService{
		counter: &mockCounterRepository{
			listFn: func(_ context.Context, _ *structs.ListCounterParams) ([]*ent.Counter, error) {
				return nil, &ent.NotFoundError{}
			},
			countXFn: func(_ context.Context, _ *structs.ListCounterParams) int {
				return 0
			},
		},
	}

	_, err := svc.List(context.Background(), &structs.ListCounterParams{Limit: 1})
	if err == nil {
		t.Fatal("expected list error for not found cursor")
	}

	expected := ecode.FieldIsInvalid("cursor")
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected error to contain %q, got %q", expected, err.Error())
	}
}

func TestCounterServiceErrorMappingFromRepository(t *testing.T) {
	svc := &counterService{
		counter: &mockCounterRepository{
			getByIDFn: func(_ context.Context, _ string) (*ent.Counter, error) {
				return nil, &ent.NotFoundError{}
			},
		},
	}

	_, err := svc.Get(context.Background(), &structs.FindCounter{Counter: "counter-1"})
	if err == nil {
		t.Fatal("expected mapped not-found error")
	}
	if err.Error() != ecode.NotExist("Counter") {
		t.Fatalf("unexpected mapped error: %q", err.Error())
	}

	internalErr := errors.New("internal error")
	svc.counter = &mockCounterRepository{
		getByIDFn: func(_ context.Context, _ string) (*ent.Counter, error) {
			return nil, internalErr
		},
	}

	_, err = svc.Get(context.Background(), &structs.FindCounter{Counter: "counter-1"})
	if err == nil || !errors.Is(err, internalErr) {
		t.Fatalf("expected internal error passthrough, got %v", err)
	}
}

var _ repository.CounterRepositoryInterface = (*mockCounterRepository)(nil)
