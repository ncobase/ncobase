package repository

import (
	"context"
	"fmt"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	dictionaryEnt "ncobase/core/system/data/ent/dictionary"
	"ncobase/core/system/structs"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/data/search/meili"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// DictionaryRepositoryInterface represents the dictionary repository interface.
type DictionaryRepositoryInterface interface {
	Create(context.Context, *structs.DictionaryBody) (*ent.Dictionary, error)
	Get(context.Context, *structs.FindDictionary) (*ent.Dictionary, error)
	Update(context.Context, *structs.UpdateDictionaryBody) (*ent.Dictionary, error)
	Delete(context.Context, *structs.FindDictionary) error
	List(context.Context, *structs.ListDictionaryParams) ([]*ent.Dictionary, error)
	CountX(context.Context, *structs.ListDictionaryParams) int
}

// dictionaryRepository implements the DictionaryRepositoryInterface.
type dictionaryRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Dictionary]
}

// NewDictionaryRepository creates a new dictionary repository.
func NewDictionaryRepository(d *data.Data) DictionaryRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &dictionaryRepository{
		ec: ec,
		rc: rc,
		ms: ms,
		c:  cache.NewCache[ent.Dictionary](rc, "ncse_dictionary", false),
	}
}

// Create creates a new dictionary.
func (r *dictionaryRepository) Create(ctx context.Context, body *structs.DictionaryBody) (*ent.Dictionary, error) {
	// create builder
	builder := r.ec.Dictionary.Create()

	// set values
	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Slug) {
		builder.SetNillableSlug(&body.Slug)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}
	if validator.IsNotEmpty(body.CreatedBy) {
		builder.SetNillableCreatedBy(body.CreatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Create error: %v", err)
		return nil, err
	}

	// Create the dictionary in Meilisearch index
	if err = r.ms.IndexDocuments("dictionarys", row); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Create error creating Meilisearch index: %v", err)
	}

	// delete cached dictionary tree
	// _ = r.c.Delete(ctx, cache.Key("dictionary=tree"))

	return row, nil
}

// Get retrieves a specific dictionary.
func (r *dictionaryRepository) Get(ctx context.Context, params *structs.FindDictionary) (*ent.Dictionary, error) {
	cacheKey := fmt.Sprintf("%s", params.Dictionary)

	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.getDictionary(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Get error: %v", err)
		return nil, err
	}

	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Get cache error: %v", err)
	}

	return row, nil
}

// Update updates an existing dictionary.
func (r *dictionaryRepository) Update(ctx context.Context, body *structs.UpdateDictionaryBody) (*ent.Dictionary, error) {
	// query the dictionary.
	row, err := r.getDictionary(ctx, &structs.FindDictionary{
		Dictionary: body.ID,
	})
	if validator.IsNotNil(err) {
		return nil, err
	}

	// create builder.
	builder := row.Update()

	// set values
	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Slug) {
		builder.SetNillableSlug(&body.Slug)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}
	if validator.IsNotEmpty(body.UpdatedBy) {
		builder.SetNillableUpdatedBy(body.UpdatedBy)
	}

	row, err = builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Update error: %v", err)
		return nil, err
	}

	// update cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	if err := r.c.Set(ctx, cacheKey, row); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Update cache error: %v", err)
	}

	// delete dictionary tree cache
	// if err := r.c.Delete(ctx, "dictionary=tree"); err != nil {
	// 	log.Errorf(ctx, "dictionaryRepo.Update cache error: %v", err)
	// }

	return row, nil
}

// Delete deletes a dictionary.
func (r *dictionaryRepository) Delete(ctx context.Context, params *structs.FindDictionary) error {
	// create builder.
	builder := r.ec.Dictionary.Delete()

	// set where conditions.
	builder.Where(dictionaryEnt.Or(
		dictionaryEnt.IDEQ(params.Dictionary),
		dictionaryEnt.SlugEQ(params.Dictionary),
	))

	// match tenant id.
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(dictionaryEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder.
	_, err := builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	cacheKey := fmt.Sprintf("%s", params.Dictionary)
	if err := r.c.Delete(ctx, cacheKey); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Delete cache error: %v", err)
	}

	return nil
}

// List lists dictionarys based on given parameters.
func (r *dictionaryRepository) List(ctx context.Context, params *structs.ListDictionaryParams) ([]*ent.Dictionary, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, err
	}

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if params.Direction == "backward" {
			builder.Where(
				dictionaryEnt.Or(
					dictionaryEnt.CreatedAtGT(timestamp),
					dictionaryEnt.And(
						dictionaryEnt.CreatedAtEQ(timestamp),
						dictionaryEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				dictionaryEnt.Or(
					dictionaryEnt.CreatedAtLT(timestamp),
					dictionaryEnt.And(
						dictionaryEnt.CreatedAtEQ(timestamp),
						dictionaryEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(dictionaryEnt.FieldCreatedAt), ent.Asc(dictionaryEnt.FieldID))
	} else {
		builder.Order(ent.Desc(dictionaryEnt.FieldCreatedAt), ent.Desc(dictionaryEnt.FieldID))
	}

	builder.Limit(params.Limit)

	return r.executeArrayQuery(ctx, builder)
}

// CountX counts dictionarys based on given parameters.
func (r *dictionaryRepository) CountX(ctx context.Context, params *structs.ListDictionaryParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder - create list builder.
func (r *dictionaryRepository) listBuilder(_ context.Context, params *structs.ListDictionaryParams) (*ent.DictionaryQuery, error) {
	// create builder.
	builder := r.ec.Dictionary.Query()

	// match tenant id.
	if params.Tenant != "" {
		builder.Where(dictionaryEnt.TenantIDEQ(params.Tenant))
	}

	// match type.
	if params.Type != "" {
		builder.Where(dictionaryEnt.TypeEQ(params.Type))
	}

	return builder, nil
}

// getDictionary - get dictionary.
// internal method.
func (r *dictionaryRepository) getDictionary(ctx context.Context, params *structs.FindDictionary) (*ent.Dictionary, error) {
	// create builder.
	builder := r.ec.Dictionary.Query()

	// set where conditions.
	if validator.IsNotEmpty(params.Dictionary) {
		builder.Where(dictionaryEnt.Or(
			dictionaryEnt.IDEQ(params.Dictionary),
			dictionaryEnt.SlugEQ(params.Dictionary),
		))
	}

	// match tenant id.
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(dictionaryEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder.
	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// executeArrayQuery - execute the builder query and return results.
func (r *dictionaryRepository) executeArrayQuery(ctx context.Context, builder *ent.DictionaryQuery) ([]*ent.Dictionary, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}
