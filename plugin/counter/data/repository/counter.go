package repository

import (
	"context"
	"fmt"
	"ncobase/common/data/cache"
	"ncobase/common/data/meili"
	"ncobase/common/logger"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/plugin/counter/data"
	"ncobase/plugin/counter/data/ent"
	counterEnt "ncobase/plugin/counter/data/ent/counter"
	"ncobase/plugin/counter/structs"

	"github.com/redis/go-redis/v9"
)

// CounterRepositoryInterface represents the counter repository interface.
type CounterRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateCounterBody) (*ent.Counter, error)
	GetByID(ctx context.Context, id string) (*ent.Counter, error)
	GetByIDs(ctx context.Context, counterIDs []string) ([]*ent.Counter, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Counter, error)
	List(ctx context.Context, params *structs.ListCounterParams) ([]*ent.Counter, error)
	Delete(ctx context.Context, slug string) error
	FindCounter(ctx context.Context, params *structs.FindCounter) (*ent.Counter, error)
	ListBuilder(ctx context.Context, params *structs.ListCounterParams) (*ent.CounterQuery, error)
	CountX(ctx context.Context, params *structs.ListCounterParams) int
}

// counterRepository implements the CounterRepositoryInterface.
type counterRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Counter]
}

// NewCounterRepository creates a new counter repository.
func NewCounterRepository(d *data.Data) CounterRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &counterRepository{ec, rc, ms, cache.NewCache[ent.Counter](rc, "ncse_counter")}
}

// Create creates a new counter.
func (r *counterRepository) Create(ctx context.Context, body *structs.CreateCounterBody) (*ent.Counter, error) {
	// create builder.
	builder := r.ec.Counter.Create()
	// set values.
	if body.Identifier != "" {
		builder.SetNillableIdentifier(&body.Identifier)
	}
	if body.Name != "" {
		builder.SetNillableName(&body.Name)
	}
	if body.Prefix != "" {
		builder.SetNillablePrefix(&body.Prefix)
	}
	if body.Suffix != "" {
		builder.SetNillableSuffix(&body.Suffix)
	}
	if body.StartValue != 0 {
		builder.SetNillableStartValue(&body.StartValue)
	}
	if body.IncrementStep != 0 {
		builder.SetNillableIncrementStep(&body.IncrementStep)
	}
	if body.DateFormat != "" {
		builder.SetNillableDateFormat(&body.DateFormat)
	}
	if body.CurrentValue != 0 {
		builder.SetNillableCurrentValue(&body.CurrentValue)
	}
	if body.Disabled {
		builder.SetNillableDisabled(&body.Disabled)
	} else {
		builder.SetDisabled(false)
	}
	if body.Description != "" {
		builder.SetNillableDescription(&body.Description)
	}
	if body.TenantID != nil {
		builder.SetNillableTenantID(body.TenantID)
	}
	if body.CreatedBy != nil {
		builder.SetNillableCreatedBy(body.CreatedBy)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.Create error: %v", err)
		return nil, err
	}

	// Create the counter in Meilisearch index
	if err = r.ms.IndexDocuments("counters", row); err != nil {
		logger.Errorf(context.Background(), "counterRepo.Create error creating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets a counter by ID.
func (r *counterRepository) GetByID(ctx context.Context, id string) (*ent.Counter, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "counters", id, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Counter); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindCounter(ctx, &structs.FindCounter{Counter: id})
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByIDs retrieves counters by their IDs.
func (r *counterRepository) GetByIDs(ctx context.Context, counterIDs []string) ([]*ent.Counter, error) {
	// create builder.
	builder := r.ec.Counter.Query()
	// set conditions.
	builder.Where(counterEnt.IDIn(counterIDs...))

	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.GetByIDs error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Update updates a counter (full or partial).
func (r *counterRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Counter, error) {
	counter, err := r.FindCounter(ctx, &structs.FindCounter{Counter: slug})
	if err != nil {
		return nil, err
	}

	builder := counter.Update()

	for field, value := range updates {
		switch field {
		case "identifier":
			builder.SetNillableIdentifier(types.ToPointer(value.(string)))
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "prefix":
			builder.SetNillablePrefix(types.ToPointer(value.(string)))
		case "suffix":
			builder.SetNillableSuffix(types.ToPointer(value.(string)))
		case "start_value":
			builder.SetNillableStartValue(types.ToPointer(value.(int)))
		case "increment_step":
			builder.SetNillableIncrementStep(types.ToPointer(value.(int)))
		case "date_format":
			builder.SetNillableDateFormat(types.ToPointer(value.(string)))
		case "current_value":
			builder.SetNillableCurrentValue(types.ToPointer(value.(int)))
		case "disabled":
			builder.SetNillableDisabled(types.ToPointer(value.(bool)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "tenant_id":
			builder.SetNillableTenantID(types.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", counter.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, counter.ID)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.Update cache error: %v", err)
	}

	// Update the counter in Meilisearch index
	if err = r.ms.DeleteDocuments("counters", slug); err != nil {
		logger.Errorf(context.Background(), "counterRepo.Update error deleting Meilisearch index: %v", err)
		// return nil, err
	}
	if err = r.ms.IndexDocuments("counters", row); err != nil {
		logger.Errorf(context.Background(), "counterRepo.Update error updating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// List gets a list of counters.
func (r *counterRepository) List(ctx context.Context, params *structs.ListCounterParams) ([]*ent.Counter, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// belong tenant
	if params.Tenant != "" {
		builder.Where(counterEnt.TenantIDEQ(params.Tenant))
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				counterEnt.Or(
					counterEnt.CreatedAtGT(timestamp),
					counterEnt.And(
						counterEnt.CreatedAtEQ(timestamp),
						counterEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				counterEnt.Or(
					counterEnt.CreatedAtLT(timestamp),
					counterEnt.And(
						counterEnt.CreatedAtEQ(timestamp),
						counterEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(counterEnt.FieldCreatedAt), ent.Asc(counterEnt.FieldID))
	} else {
		builder.Order(ent.Desc(counterEnt.FieldCreatedAt), ent.Desc(counterEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a counter.
func (r *counterRepository) Delete(ctx context.Context, slug string) error {
	counter, err := r.FindCounter(ctx, &structs.FindCounter{Counter: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Counter.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(counterEnt.IDEQ(counter.ID)).Exec(ctx); err != nil {
		logger.Errorf(context.Background(), "counterRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", counter.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("counter:slug:%s", counter.ID))
	if err != nil {
		logger.Errorf(context.Background(), "counterRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("counters", counter.ID); err != nil {
		logger.Errorf(context.Background(), "counterRepo.Delete index error: %v", err)
		// return nil, err
	}

	return nil
}

// FindCounter finds a counter.
func (r *counterRepository) FindCounter(ctx context.Context, params *structs.FindCounter) (*ent.Counter, error) {

	// create builder.
	builder := r.ec.Counter.Query()

	if validator.IsNotEmpty(params.Counter) {
		builder = builder.Where(counterEnt.Or(
			counterEnt.ID(params.Counter),
		))
	}
	if validator.IsNotEmpty(params.Tenant) {
		builder = builder.Where(counterEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder creates list builder.
func (r *counterRepository) ListBuilder(_ context.Context, _ *structs.ListCounterParams) (*ent.CounterQuery, error) {
	// create builder.
	builder := r.ec.Counter.Query()

	return builder, nil
}

// CountX gets a count of counters.
func (r *counterRepository) CountX(ctx context.Context, params *structs.ListCounterParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// Count gets a count of counters.
func (r *counterRepository) Count(ctx context.Context, params *structs.ListCounterParams) (int, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0, err
	}
	return builder.Count(ctx)
}
