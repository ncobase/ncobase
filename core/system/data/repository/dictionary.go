package repository

import (
	"context"
	"fmt"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	dictionaryEnt "ncobase/core/system/data/ent/dictionary"
	"ncobase/core/system/structs"
	"time"

	nd "github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/ncobase/ncore/data/search"
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
	data             *data.Data
	searchClient     *search.Client
	dictionaryCache  cache.ICache[ent.Dictionary]
	slugMappingCache cache.ICache[string] // Maps slug to dictionary ID
	dictionaryTTL    time.Duration
}

// NewDictionaryRepository creates a new dictionary repository.
func NewDictionaryRepository(d *data.Data) DictionaryRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)
	searchClient := nd.NewSearchClient(d.Data)

	return &dictionaryRepository{
		data:             d,
		searchClient:     searchClient,
		dictionaryCache:  cache.NewCache[ent.Dictionary](redisClient, "ncse_system:dictionaries"),
		slugMappingCache: cache.NewCache[string](redisClient, "ncse_system:dict_mappings"),
		dictionaryTTL:    time.Hour * 4, // 4 hours cache TTL
	}
}

// Create creates a new dictionary.
func (r *dictionaryRepository) Create(ctx context.Context, body *structs.DictionaryBody) (*ent.Dictionary, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().Dictionary.Create()

	// Set values
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
	if validator.IsNotEmpty(body.Description) {
		builder.SetNillableDescription(&body.Description)
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
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "dictionaries", Document: row}); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Create error creating Meilisearch index: %v", err)
	}

	// Cache the dictionary
	go r.cacheDictionary(context.Background(), row)

	return row, nil
}

// Get retrieves a specific dictionary.
func (r *dictionaryRepository) Get(ctx context.Context, params *structs.FindDictionary) (*ent.Dictionary, error) {
	// Try to get dictionary ID from slug mapping cache if searching by slug
	if params.Dictionary != "" {
		if dictID, err := r.getDictIDBySlug(ctx, params.Dictionary); err == nil && dictID != "" {
			// Try to get from dictionary cache
			cacheKey := fmt.Sprintf("id:%s", dictID)
			if cached, err := r.dictionaryCache.Get(ctx, cacheKey); err == nil && cached != nil {
				return cached, nil
			}
		}
	}

	// Fallback to database
	row, err := r.getDictionary(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Get error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheDictionary(context.Background(), row)

	return row, nil
}

// Update updates an existing dictionary.
func (r *dictionaryRepository) Update(ctx context.Context, body *structs.UpdateDictionaryBody) (*ent.Dictionary, error) {
	// Query the dictionary
	originalDict, err := r.getDictionary(ctx, &structs.FindDictionary{
		Dictionary: body.ID,
	})
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Use master for writes
	builder := originalDict.Update()

	// Set values
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
	if validator.IsNotEmpty(body.Description) {
		builder.SetNillableDescription(&body.Description)
	}
	if validator.IsNotEmpty(body.UpdatedBy) {
		builder.SetNillableUpdatedBy(body.UpdatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.searchClient.Index(ctx, &search.IndexRequest{Index: "dictionaries", Document: row, DocumentID: row.ID}); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Update error updating Meilisearch index: %v", err)
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateDictionaryCache(context.Background(), originalDict)
		r.cacheDictionary(context.Background(), row)
	}()

	return row, nil
}

// Delete deletes a dictionary.
func (r *dictionaryRepository) Delete(ctx context.Context, params *structs.FindDictionary) error {
	// Get dictionary first for cache invalidation
	dict, err := r.getDictionary(ctx, params)
	if err != nil {
		return err
	}

	// Use master for writes
	builder := r.data.GetMasterEntClient().Dictionary.Delete()

	// Set where conditions
	builder.Where(dictionaryEnt.Or(
		dictionaryEnt.IDEQ(params.Dictionary),
		dictionaryEnt.SlugEQ(params.Dictionary),
	))

	// Execute the builder
	_, err = builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	// Delete from Meilisearch index
	if err = r.searchClient.Delete(ctx, "dictionaries", dict.ID); err != nil {
		logger.Errorf(ctx, "dictionaryRepo.Delete index error: %v", err)
	}

	// Invalidate cache
	go r.invalidateDictionaryCache(context.Background(), dict)

	return nil
}

// List lists dictionaries based on given parameters.
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

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
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

	rows, err := r.executeArrayQuery(ctx, builder)
	if err != nil {
		return nil, err
	}

	// Cache dictionaries in background
	go func() {
		for _, dict := range rows {
			r.cacheDictionary(context.Background(), dict)
		}
	}()

	return rows, nil
}

// CountX counts dictionaries based on given parameters.
func (r *dictionaryRepository) CountX(ctx context.Context, params *structs.ListDictionaryParams) int {
	// Create list builder using slave
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder - create list builder.
func (r *dictionaryRepository) listBuilder(_ context.Context, params *structs.ListDictionaryParams) (*ent.DictionaryQuery, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Dictionary.Query()

	// Match type
	if params.Type != "" {
		builder.Where(dictionaryEnt.TypeEQ(params.Type))
	}

	return builder, nil
}

// getDictionary - get dictionary.
// internal method.
func (r *dictionaryRepository) getDictionary(ctx context.Context, params *structs.FindDictionary) (*ent.Dictionary, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Dictionary.Query()

	// Set where conditions
	if validator.IsNotEmpty(params.Dictionary) {
		builder.Where(dictionaryEnt.Or(
			dictionaryEnt.IDEQ(params.Dictionary),
			dictionaryEnt.SlugEQ(params.Dictionary),
		))
	}

	// Execute the builder
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

// cacheDictionary - cache dictionary.
func (r *dictionaryRepository) cacheDictionary(ctx context.Context, dict *ent.Dictionary) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", dict.ID)
	if err := r.dictionaryCache.Set(ctx, idKey, dict, r.dictionaryTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache dictionary by ID %s: %v", dict.ID, err)
	}

	// Cache slug to ID mapping
	if dict.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", dict.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &dict.ID, r.dictionaryTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", dict.Slug, err)
		}
	}
}

// invalidateDictionaryCache invalidates dictionary cache
func (r *dictionaryRepository) invalidateDictionaryCache(ctx context.Context, dict *ent.Dictionary) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", dict.ID)
	if err := r.dictionaryCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate dictionary ID cache %s: %v", dict.ID, err)
	}

	// Invalidate slug mapping
	if dict.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", dict.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", dict.Slug, err)
		}
	}
}

// getDictIDBySlug - get dictionary ID by slug
func (r *dictionaryRepository) getDictIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	dictID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || dictID == nil {
		return "", err
	}
	return *dictID, nil
}
