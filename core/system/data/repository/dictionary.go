package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	dictionaryEnt "ncobase/core/system/data/ent/dictionary"
	menuEnt "ncobase/core/system/data/ent/menu"
	optionsEnt "ncobase/core/system/data/ent/options"
	"ncobase/core/system/data/ent/predicate"
	"ncobase/core/system/structs"
	"strings"
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
	GetUsage(context.Context, *structs.FindDictionary) ([]*structs.DictionaryUsage, error)
	List(context.Context, *structs.ListDictionaryParams) ([]*ent.Dictionary, error)
	CountX(context.Context, *structs.ListDictionaryParams) int
}

// dictionaryRepository implements the DictionaryRepositoryInterface.
type dictionaryRepository struct {
	data             *data.Data
	sc               *search.Client
	dictionaryCache  cache.ICache[ent.Dictionary]
	slugMappingCache cache.ICache[string] // Maps slug to dictionary ID
	dictionaryTTL    time.Duration
}

// NewDictionaryRepository creates a new dictionary repository.
func NewDictionaryRepository(d *data.Data) DictionaryRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)
	sc := nd.NewSearchClient(d.Data)

	return &dictionaryRepository{
		data:             d,
		sc:               sc,
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
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "dictionaries", Document: row}); err != nil {
			logger.Errorf(ctx, "dictionaryRepo.Create error creating Meilisearch index: %v", err)
		}
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
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "dictionaries", Document: row, DocumentID: row.ID}); err != nil {
			logger.Errorf(ctx, "dictionaryRepo.Update error updating Meilisearch index: %v", err)
		}
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
	if r.sc != nil {
		if err = r.sc.Delete(ctx, "dictionaries", dict.ID); err != nil {
			logger.Errorf(ctx, "dictionaryRepo.Delete index error: %v", err)
		}
	}

	// Invalidate cache
	go r.invalidateDictionaryCache(context.Background(), dict)

	return nil
}

// GetUsage finds system records that reference a dictionary id or slug.
func (r *dictionaryRepository) GetUsage(ctx context.Context, params *structs.FindDictionary) ([]*structs.DictionaryUsage, error) {
	dict, err := r.getDictionary(ctx, params)
	if err != nil {
		return nil, err
	}

	terms := dictionaryUsageTerms(dict)
	if len(terms) == 0 {
		return []*structs.DictionaryUsage{}, nil
	}

	usage := make([]*structs.DictionaryUsage, 0)

	optionPredicates := make([]predicate.Options, 0, len(terms)*3)
	for _, term := range terms {
		optionPredicates = append(optionPredicates,
			optionsEnt.NameContainsFold(term),
			optionsEnt.TypeContainsFold(term),
			optionsEnt.ValueContainsFold(term),
		)
	}
	optionsRows, err := r.data.GetSlaveEntClient().Options.Query().
		Where(optionsEnt.Or(optionPredicates...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, option := range optionsRows {
		count := countDictionaryTermMatches(terms, option.Name, option.Type, option.Value)
		if count == 0 {
			continue
		}
		usage = append(usage, &structs.DictionaryUsage{
			Module:      "system.options",
			Location:    fmt.Sprintf("option:%s", option.Name),
			Count:       count,
			ReferenceID: option.ID,
		})
	}

	menuPredicates := make([]predicate.Menu, 0, len(terms)*8)
	for _, term := range terms {
		menuPredicates = append(menuPredicates,
			menuEnt.NameContainsFold(term),
			menuEnt.LabelContainsFold(term),
			menuEnt.SlugContainsFold(term),
			menuEnt.TypeContainsFold(term),
			menuEnt.PathContainsFold(term),
			menuEnt.TargetContainsFold(term),
			menuEnt.PermsContainsFold(term),
			menuEnt.ParentIDContainsFold(term),
		)
	}
	menuRows, err := r.data.GetSlaveEntClient().Menu.Query().
		Where(menuEnt.Or(menuPredicates...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, menu := range menuRows {
		count := countDictionaryTermMatches(
			terms,
			menu.Name,
			menu.Label,
			menu.Slug,
			menu.Type,
			menu.Path,
			menu.Target,
			menu.Perms,
			menu.ParentID,
		)
		if count == 0 {
			continue
		}
		usage = append(usage, &structs.DictionaryUsage{
			Module:      "system.menus",
			Location:    fmt.Sprintf("menu:%s", menu.Slug),
			Count:       count,
			ReferenceID: menu.ID,
		})
	}

	menuExtrasRows, err := r.data.GetSlaveEntClient().Menu.Query().
		Where(menuEnt.ExtrasNotNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	seenMenuExtras := make(map[string]struct{}, len(menuRows))
	for _, menu := range menuRows {
		seenMenuExtras[menu.ID] = struct{}{}
	}
	for _, menu := range menuExtrasRows {
		if _, ok := seenMenuExtras[menu.ID]; ok {
			continue
		}
		raw, err := json.Marshal(menu.Extras)
		if err != nil {
			continue
		}
		count := countDictionaryTermMatches(terms, string(raw))
		if count == 0 {
			continue
		}
		usage = append(usage, &structs.DictionaryUsage{
			Module:      "system.menus",
			Location:    fmt.Sprintf("menu:%s:extras", menu.Slug),
			Count:       count,
			ReferenceID: menu.ID,
		})
	}

	dictionaryPredicates := make([]predicate.Dictionary, 0, len(terms)*3)
	for _, term := range terms {
		dictionaryPredicates = append(dictionaryPredicates,
			dictionaryEnt.NameContainsFold(term),
			dictionaryEnt.DescriptionContainsFold(term),
			dictionaryEnt.ValueContainsFold(term),
		)
	}
	dictionaryRows, err := r.data.GetSlaveEntClient().Dictionary.Query().
		Where(dictionaryEnt.And(
			dictionaryEnt.IDNEQ(dict.ID),
			dictionaryEnt.Or(dictionaryPredicates...),
		)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, dictionary := range dictionaryRows {
		count := countDictionaryTermMatches(terms, dictionary.Name, dictionary.Description, dictionary.Value)
		if count == 0 {
			continue
		}
		usage = append(usage, &structs.DictionaryUsage{
			Module:      "system.dictionaries",
			Location:    fmt.Sprintf("dictionary:%s", dictionary.Slug),
			Count:       count,
			ReferenceID: dictionary.ID,
		})
	}

	return usage, nil
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

func dictionaryUsageTerms(dict *ent.Dictionary) []string {
	rawTerms := []string{dict.ID, dict.Slug}
	if dict.ID == "" && dict.Slug == "" {
		rawTerms = append(rawTerms, dict.Name)
	}
	seen := make(map[string]struct{}, len(rawTerms))
	terms := make([]string, 0, len(rawTerms))
	for _, term := range rawTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		key := strings.ToLower(term)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		terms = append(terms, key)
	}
	return terms
}

func countDictionaryTermMatches(terms []string, values ...string) int {
	count := 0
	for _, value := range values {
		value = strings.ToLower(value)
		if value == "" {
			continue
		}
		for _, term := range terms {
			if strings.Contains(value, term) {
				count++
			}
		}
	}
	return count
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
