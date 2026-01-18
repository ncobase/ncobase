package repository

import (
	"context"
	"fmt"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	optionsEnt "ncobase/core/system/data/ent/options"
	"ncobase/core/system/structs"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	nd "github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/search"
	"github.com/redis/go-redis/v9"
)

// OptionRepositoryInterface represents the option repository interface.
type OptionRepositoryInterface interface {
	Create(context.Context, *structs.OptionBody) (*ent.Options, error)
	Get(context.Context, *structs.FindOptions) (*ent.Options, error)
	Update(context.Context, *structs.UpdateOptionBody) (*ent.Options, error)
	Delete(context.Context, *structs.FindOptions) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	List(context.Context, *structs.ListOptionParams) ([]*ent.Options, error)
	ListWithCount(ctx context.Context, params *structs.ListOptionParams) ([]*ent.Options, int, error)
	CountX(context.Context, *structs.ListOptionParams) int
}

// optionRepository implements the OptionRepositoryInterface.
type optionRepository struct {
	data             *data.Data
	sc               *search.Client
	optionsCache     cache.ICache[ent.Options]
	nameMappingCache cache.ICache[string] // Maps name to option ID
	optionsTTL       time.Duration
}

// NewOptionRepository creates a new option repository.
func NewOptionRepository(d *data.Data) OptionRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)
	sc := nd.NewSearchClient(d.Data)

	return &optionRepository{
		data:             d,
		sc:               sc,
		optionsCache:     cache.NewCache[ent.Options](redisClient, "ncse_system:options"),
		nameMappingCache: cache.NewCache[string](redisClient, "ncse_system:option_mappings"),
		optionsTTL:       time.Hour * 6, // 6 hours cache TTL
	}
}

// Create creates a new option.
func (r *optionRepository) Create(ctx context.Context, body *structs.OptionBody) (*ent.Options, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().Options.Create()

	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotNil(body.Autoload) {
		builder.SetAutoload(body.Autoload)
	}
	if validator.IsNotEmpty(body.CreatedBy) {
		builder.SetNillableCreatedBy(body.CreatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Create error: %v", err)
		return nil, err
	}

	// Create in Meilisearch index
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "options", Document: row}); err != nil {
			logger.Errorf(ctx, "optionsRepo.Create error creating Meilisearch index: %v", err)
		}
	}

	// Cache the option
	go r.cacheOption(context.Background(), row)

	return row, nil
}

// Get retrieves a specific option.
func (r *optionRepository) Get(ctx context.Context, params *structs.FindOptions) (*ent.Options, error) {
	// Try to get option ID from name mapping cache if searching by name
	if params.Option != "" {
		if optionID, err := r.getOptionIDByName(ctx, params.Option); err == nil && optionID != "" {
			// Try to get from option cache
			cacheKey := fmt.Sprintf("id:%s", optionID)
			if cached, err := r.optionsCache.Get(ctx, cacheKey); err == nil && cached != nil {
				return cached, nil
			}
		}
	}

	// Fallback to database
	row, err := r.getOption(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Get error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheOption(context.Background(), row)

	return row, nil
}

// Update updates an existing option.
func (r *optionRepository) Update(ctx context.Context, body *structs.UpdateOptionBody) (*ent.Options, error) {
	// Get original option for cache invalidation
	originalOption, err := r.getOption(ctx, &structs.FindOptions{
		Option: body.ID,
	})
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Use master for writes
	builder := originalOption.Update()

	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotNil(body.Autoload) {
		builder.SetAutoload(body.Autoload)
	}
	if validator.IsNotEmpty(body.UpdatedBy) {
		builder.SetNillableUpdatedBy(body.UpdatedBy)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "options", Document: row, DocumentID: row.ID}); err != nil {
			logger.Errorf(ctx, "optionsRepo.Update error updating Meilisearch index: %v", err)
		}
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateOptionCache(context.Background(), originalOption)
		r.cacheOption(context.Background(), row)
	}()

	return row, nil
}

// Delete deletes an option.
func (r *optionRepository) Delete(ctx context.Context, params *structs.FindOptions) error {
	// Get option first for cache invalidation
	option, err := r.getOption(ctx, params)
	if err != nil {
		return err
	}

	// Use master for writes
	builder := r.data.GetMasterEntClient().Options.Delete()

	builder.Where(optionsEnt.Or(
		optionsEnt.IDEQ(params.Option),
		optionsEnt.NameEQ(params.Option),
	))

	_, err = builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	if r.sc != nil {
		if err = r.sc.Delete(ctx, "options", option.ID); err != nil {
			logger.Errorf(ctx, "optionRepo.Delete index error: %v", err)
		}
	}

	// Invalidate cache
	go r.invalidateOptionCache(context.Background(), option)

	return nil
}

// DeleteByPrefix deletes options by prefix.
func (r *optionRepository) DeleteByPrefix(ctx context.Context, prefix string) error {
	if validator.IsEmpty(prefix) {
		return fmt.Errorf("prefix is required")
	}

	// Get options to be deleted for cache invalidation
	options, err := r.data.GetSlaveEntClient().Options.Query().Where(optionsEnt.NameHasPrefix(prefix)).All(ctx)
	if err != nil {
		logger.Debugf(ctx, "Failed to get options for cache invalidation: %v", err)
	}

	// Use master for writes - Delete all options with the given prefix in one operation
	_, err = r.data.GetMasterEntClient().Options.Delete().Where(optionsEnt.NameHasPrefix(prefix)).Exec(ctx)
	if err != nil {
		return err
	}

	// Invalidate caches for deleted options
	go func() {
		for _, option := range options {
			r.invalidateOptionCache(context.Background(), option)
		}
	}()

	return nil
}

// List returns a slice of options based on the provided parameters.
func (r *optionRepository) List(ctx context.Context, params *structs.ListOptionParams) ([]*ent.Options, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("building list query: %w", err)
	}

	builder = r.applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decoding cursor: %w", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		builder = r.applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	builder.Limit(params.Limit)

	rows, err := r.executeArrayQuery(ctx, builder)
	if err != nil {
		return nil, err
	}

	// Cache options in background
	go func() {
		for _, option := range rows {
			r.cacheOption(context.Background(), option)
		}
	}()

	return rows, nil
}

// CountX returns the total count of options based on the provided parameters.
func (r *optionRepository) CountX(ctx context.Context, params *structs.ListOptionParams) int {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Error building count query: %v", err)
		return 0
	}
	return builder.CountX(ctx)
}

// ListWithCount returns both a slice of options and the total count based on the provided parameters.
func (r *optionRepository) ListWithCount(ctx context.Context, params *structs.ListOptionParams) ([]*ent.Options, int, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("building list query: %w", err)
	}

	builder = r.applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("decoding cursor: %w", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, 0, fmt.Errorf("invalid id in cursor: %s", id)
		}

		builder = r.applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting options: %w", err)
	}

	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching options: %w", err)
	}

	// Cache options in background
	go func() {
		for _, option := range rows {
			r.cacheOption(context.Background(), option)
		}
	}()

	return rows, total, nil
}

// applySorting applies the specified sorting to the query builder.
func (r *optionRepository) applySorting(builder *ent.OptionsQuery, sortBy string) *ent.OptionsQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		return builder.Order(ent.Desc(optionsEnt.FieldCreatedAt), ent.Desc(optionsEnt.FieldID))
	case structs.SortByName:
		return builder.Order(ent.Asc(optionsEnt.FieldName), ent.Desc(optionsEnt.FieldID))
	default:
		return builder.Order(ent.Desc(optionsEnt.FieldCreatedAt), ent.Desc(optionsEnt.FieldID))
	}
}

// applyCursorCondition applies the cursor-based condition to the query builder.
func (r *optionRepository) applyCursorCondition(builder *ent.OptionsQuery, id string, value any, direction string, sortBy string) *ent.OptionsQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				optionsEnt.Or(
					optionsEnt.CreatedAtGT(timestamp),
					optionsEnt.And(
						optionsEnt.CreatedAtEQ(timestamp),
						optionsEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			optionsEnt.Or(
				optionsEnt.CreatedAtLT(timestamp),
				optionsEnt.And(
					optionsEnt.CreatedAtEQ(timestamp),
					optionsEnt.IDLT(id),
				),
			),
		)
	case structs.SortByName:
		name, ok := value.(string)
		if !ok {
			logger.Errorf(context.Background(), "Invalid name value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				optionsEnt.Or(
					optionsEnt.NameGT(name),
					optionsEnt.And(
						optionsEnt.NameEQ(name),
						optionsEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			optionsEnt.Or(
				optionsEnt.NameLT(name),
				optionsEnt.And(
					optionsEnt.NameEQ(name),
					optionsEnt.IDLT(id),
				),
			),
		)
	default:
		return r.applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// listBuilder - create list builder.
func (r *optionRepository) listBuilder(_ context.Context, params *structs.ListOptionParams) (*ent.OptionsQuery, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Options.Query()

	if params.Type != "" {
		builder.Where(optionsEnt.TypeEQ(params.Type))
	}

	if params.Autoload != nil {
		builder.Where(optionsEnt.AutoloadEQ(*params.Autoload))
	}

	if params.Prefix != "" {
		builder.Where(optionsEnt.NameHasPrefix(params.Prefix))
	}

	return builder, nil
}

// getOption - get option.
// internal method.
func (r *optionRepository) getOption(ctx context.Context, params *structs.FindOptions) (*ent.Options, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().Options.Query()

	if validator.IsNotEmpty(params.Option) {
		builder.Where(optionsEnt.Or(
			optionsEnt.IDEQ(params.Option),
			optionsEnt.NameEQ(params.Option),
		))
	}

	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// executeArrayQuery - execute the builder query and return results.
func (r *optionRepository) executeArrayQuery(ctx context.Context, builder *ent.OptionsQuery) ([]*ent.Options, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}

// cacheOption - cache option.
func (r *optionRepository) cacheOption(ctx context.Context, option *ent.Options) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", option.ID)
	if err := r.optionsCache.Set(ctx, idKey, option, r.optionsTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache option by ID %s: %v", option.ID, err)
	}

	// Cache name to ID mapping
	if option.Name != "" {
		nameKey := fmt.Sprintf("name:%s", option.Name)
		if err := r.nameMappingCache.Set(ctx, nameKey, &option.ID, r.optionsTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache name mapping %s: %v", option.Name, err)
		}
	}
}

// invalidateOptionCache invalidates option cache
func (r *optionRepository) invalidateOptionCache(ctx context.Context, option *ent.Options) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", option.ID)
	if err := r.optionsCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate option ID cache %s: %v", option.ID, err)
	}

	// Invalidate name mapping
	if option.Name != "" {
		nameKey := fmt.Sprintf("name:%s", option.Name)
		if err := r.nameMappingCache.Delete(ctx, nameKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate name mapping cache %s: %v", option.Name, err)
		}
	}
}

// getOptionIDByName - get option ID by name
func (r *optionRepository) getOptionIDByName(ctx context.Context, name string) (string, error) {
	cacheKey := fmt.Sprintf("name:%s", name)
	optionID, err := r.nameMappingCache.Get(ctx, cacheKey)
	if err != nil || optionID == nil {
		return "", err
	}
	return *optionID, nil
}
