package repository

import (
	"context"
	"fmt"
	"ncobase/domain/content/data"
	"ncobase/domain/content/data/ent"
	taxonomyEnt "ncobase/domain/content/data/ent/taxonomy"
	"ncobase/domain/content/structs"
	"ncore/pkg/data/cache"
	"ncore/pkg/data/meili"
	"ncore/pkg/logger"
	"ncore/pkg/paging"
	"ncore/pkg/types"
	"ncore/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// TaxonomyRepositoryInterface represents the taxonomy repository interface.
type TaxonomyRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*ent.Taxonomy, error)
	GetByID(ctx context.Context, id string) (*ent.Taxonomy, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Taxonomy, error)
	GetTree(ctx context.Context, params *structs.FindTaxonomy) ([]*ent.Taxonomy, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Taxonomy, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error)
	ListWithCount(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, int, error)
	Delete(ctx context.Context, slug string) error
	FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error)
	CountX(ctx context.Context, params *structs.ListTaxonomyParams) int
}

// taxonomyRepository implements the TaxonomyRepositoryInterface.
type taxonomyRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	ms  *meili.Client
	c   *cache.Cache[ent.Taxonomy]
}

// NewTaxonomyRepository creates a new taxonomy repository.
func NewTaxonomyRepository(d *data.Data) TaxonomyRepositoryInterface {
	ec := d.GetEntClient()
	ecr := d.GetEntClientRead()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &taxonomyRepository{ec, ecr, rc, ms, cache.NewCache[ent.Taxonomy](rc, "ncse_taxonomy")}
}

// Create create taxonomy
func (r *taxonomyRepository) Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*ent.Taxonomy, error) {

	// create builder.
	builder := r.ec.Taxonomy.Create()
	// set values.
	builder.SetNillableName(&body.Name)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableCover(&body.Cover)
	builder.SetNillableThumbnail(&body.Thumbnail)
	builder.SetNillableColor(&body.Color)
	builder.SetNillableIcon(&body.Icon)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableKeywords(&body.Keywords)
	builder.SetNillableDescription(&body.Description)
	builder.SetStatus(body.Status)
	builder.SetNillableParentID(body.ParentID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Create error: %v", err)
		return nil, err
	}

	// Create the taxonomy in Meilisearch index
	if err = r.ms.IndexDocuments("taxonomies", row); err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Create error creating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// GetByID get taxonomy by id
func (r *taxonomyRepository) GetByID(ctx context.Context, id string) (*ent.Taxonomy, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "taxonomies", id, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Taxonomy); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Taxonomy: id})

	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.GetByID error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetBySlug get taxonomy by slug
func (r *taxonomyRepository) GetBySlug(ctx context.Context, slug string) (*ent.Taxonomy, error) {
	// // Search in Meilisearch index
	// if res, _ := r.ms.Search(ctx, "taxonomies", slug, &meilisearch.SearchRequest{Limit: 1}); res != nil && res.Hits != nil && len(res.Hits) > 0 {
	// 	if hit, ok := res.Hits[0].(*ent.Taxonomy); ok {
	// 		return hit, nil
	// 	}
	// }
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Taxonomy: slug})

	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.GetBySlug cache error: %v", err)
	}

	return row, nil
}

// GetTree retrieves the taxonomy tree.
func (r *taxonomyRepository) GetTree(ctx context.Context, params *structs.FindTaxonomy) ([]*ent.Taxonomy, error) {
	// create builder
	builder := r.ecr.Taxonomy.Query()

	// set where conditions
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(taxonomyEnt.TenantIDEQ(params.Tenant))
	}

	// handle sub taxonomies
	if validator.IsNotEmpty(params.Taxonomy) && params.Taxonomy != "root" {
		return r.getSubTaxonomy(ctx, params.Taxonomy, builder)
	}

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// Update update taxonomy
func (r *taxonomyRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Taxonomy, error) {
	taxonomy, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Taxonomy: slug})
	if err != nil {
		return nil, err
	}

	builder := taxonomy.Update()

	// set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "cover":
			builder.SetNillableCover(types.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(types.ToPointer(value.(string)))
		case "color":
			builder.SetNillableColor(types.ToPointer(value.(string)))
		case "icon":
			builder.SetNillableIcon(types.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(types.ToPointer(value.(string)))
		case "keywords":
			builder.SetNillableKeywords(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "status":
			builder.SetStatus(value.(int))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "parent_id":
			builder.SetNillableParentID(types.ToPointer(value.(string)))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", taxonomy.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, taxonomy.Slug)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Update cache error: %v", err)
	}

	// Update the taxonomy in Meilisearch
	if err = r.ms.DeleteDocuments("taxonomies", slug); err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Update error deleting Meilisearch index: %v", err)
		// return nil, err
	}
	if err = r.ms.IndexDocuments("taxonomies", row, row.ID); err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Update error updating Meilisearch index: %v", err)
		// return nil, err
	}

	return row, nil
}

// // List get taxonomy list
// func (r *taxonomyRepository) List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error) {
// 	// Generate cache key based on query parameters
// 	// cacheKey := fmt.Sprintf("list_taxonomy_%s_%d_%s_%s", params.Cursor, params.Limit, params.TenantID, params.Type)
//
// 	// // check cache first
// 	// cachedResult, err := r.c.Get(ctx, cacheKey)
// 	// if err == nil {
// 	// 	// Convert cached JSON data back to []*ent.Taxonomy
// 	// 	var result []*ent.Taxonomy
// 	// 	if err := json.Unmarshal([]byte(cachedResult), &result); err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// 	return result, nil
// 	// } else if err != cache.ErrCacheMiss {
// 	// 	log.Errorf(ctx, "taxonomyRepo.List cache error: %v", err)
// 	// }
//
// 	// create list builder
// 	builder, err := r.listBuilder(ctx, params)
// 	if validator.IsNotNil(err) {
// 		return nil, err
// 	}
//
// 	if params.Cursor != "" {
// 		id, timestamp, err := paging.DecodeCursor(params.Cursor)
// 		if err != nil {
// 			return nil, fmt.Errorf("invalid cursor: %v", err)
// 		}
//
// 		if !nanoid.IsPrimaryKey(id) {
// 			return nil, fmt.Errorf("invalid id in cursor: %s", id)
// 		}
//
// 		if params.Direction == "backward" {
// 			builder.Where(
// 				taxonomyEnt.Or(
// 					taxonomyEnt.CreatedAtGT(timestamp),
// 					taxonomyEnt.And(
// 						taxonomyEnt.CreatedAtEQ(timestamp),
// 						taxonomyEnt.IDGT(id),
// 					),
// 				),
// 			)
// 		} else {
// 			builder.Where(
// 				taxonomyEnt.Or(
// 					taxonomyEnt.CreatedAtLT(timestamp),
// 					taxonomyEnt.And(
// 						taxonomyEnt.CreatedAtEQ(timestamp),
// 						taxonomyEnt.IDLT(id),
// 					),
// 				),
// 			)
// 		}
// 	}
//
// 	if params.Direction == "backward" {
// 		builder.Order(ent.Asc(taxonomyEnt.FieldCreatedAt), ent.Asc(taxonomyEnt.FieldID))
// 	} else {
// 		builder.Order(ent.Desc(taxonomyEnt.FieldCreatedAt), ent.Desc(taxonomyEnt.FieldID))
// 	}
//
// 	builder.Limit(params.Limit)
//
// 	// belong tenant
// 	if params.Tenant != "" {
// 		builder.Where(taxonomyEnt.TenantIDEQ(params.Tenant))
// 	}
//
// 	// type
// 	if params.Type == "all" {
// 		params.Type = ""
// 	}
// 	if params.Type != "" {
// 		builder.Where(taxonomyEnt.TypeEQ(params.Type))
// 	}
//
// 	rows, err := builder.All(ctx)
// 	if err != nil {
// 		log.Errorf(ctx, "taxonomyRepo.List error: %v", err)
// 		return nil, err
// 	}
//
// 	// // Convert []*ent.Taxonomy to JSON data
// 	// jsonData, err := json.Marshal(rows)
// 	// if err != nil {
// 	// 	log.Errorf(ctx, "taxonomyRepo.List cache error: %v", err)
// 	// } else {
// 	// 	// cache the result
// 	// 	if err := r.c.Set(ctx, cacheKey, jsonData); err != nil {
// 	// 		log.Errorf(ctx, "taxonomyRepo.List cache error: %v", err)
// 	// 	}
// 	// }
//
// 	return rows, nil
// }

// Delete delete taxonomy
func (r *taxonomyRepository) Delete(ctx context.Context, slug string) error {
	taxonomy, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Taxonomy: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Taxonomy.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(taxonomyEnt.IDEQ(taxonomy.ID)).Exec(ctx); err == nil {
		logger.Errorf(ctx, "taxonomyRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	err = r.c.Delete(ctx, taxonomy.ID)
	err = r.c.Delete(ctx, slug)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.Delete cache error: %v", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("taxonomies", taxonomy.ID); err != nil {
		logger.Errorf(ctx, "topicRepo.Delete index error: %v", err)
		// return nil, err
	}

	return err
}

// FindTaxonomy find taxonomy
func (r *taxonomyRepository) FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error) {
	// create builder.
	builder := r.ecr.Taxonomy.Query()

	// support slug or ID
	if validator.IsNotEmpty(params.Taxonomy) {
		builder = builder.Where(taxonomyEnt.Or(
			taxonomyEnt.ID(params.Taxonomy),
			taxonomyEnt.SlugEQ(params.Taxonomy),
		))
	}
	if validator.IsNotEmpty(params.Tenant) {
		builder = builder.Where(taxonomyEnt.TenantIDEQ(params.Tenant))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder create list builder
func (r *taxonomyRepository) listBuilder(_ context.Context, params *structs.ListTaxonomyParams) (*ent.TaxonomyQuery, error) {
	// create builder.
	builder := r.ecr.Taxonomy.Query()

	// match parent id.
	// default is root.
	if validator.IsEmpty(params.Parent) {
		builder.Where(taxonomyEnt.Or(
			taxonomyEnt.ParentIDIsNil(),
			taxonomyEnt.ParentIDEQ(""),
			taxonomyEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(taxonomyEnt.ParentIDEQ(params.Parent))
	}

	return builder, nil
}

// List returns a slice of taxonomies based on the provided parameters.
func (r *taxonomyRepository) List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("building list query: %w", err)
	}

	builder = applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	builder.Limit(params.Limit)

	return r.executeArrayQuery(ctx, builder)
}

// CountX returns the total count of taxonomies based on the provided parameters.
func (r *taxonomyRepository) CountX(ctx context.Context, params *structs.ListTaxonomyParams) int {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Error building count query: %v", err)
		return 0
	}
	return builder.CountX(ctx)
}

// ListWithCount returns both a slice of taxonomies and the total count based on the provided parameters.
func (r *taxonomyRepository) ListWithCount(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, int, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("building list query: %w", err)
	}

	builder = applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting taxonomies: %w", err)
	}

	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching taxonomies: %w", err)
	}

	return rows, total, nil
}

// applySorting applies the specified sorting to the query builder.
func applySorting(builder *ent.TaxonomyQuery, sortBy string) *ent.TaxonomyQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		return builder.Order(ent.Desc(taxonomyEnt.FieldCreatedAt), ent.Desc(taxonomyEnt.FieldID))
	default:
		return builder.Order(ent.Desc(taxonomyEnt.FieldCreatedAt), ent.Desc(taxonomyEnt.FieldID))
	}
}

// applyCursorCondition applies the cursor-based condition to the query builder.
func applyCursorCondition(builder *ent.TaxonomyQuery, id string, value any, direction string, sortBy string) *ent.TaxonomyQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				taxonomyEnt.Or(
					taxonomyEnt.CreatedAtGT(timestamp),
					taxonomyEnt.And(
						taxonomyEnt.CreatedAtEQ(timestamp),
						taxonomyEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			taxonomyEnt.Or(
				taxonomyEnt.CreatedAtLT(timestamp),
				taxonomyEnt.And(
					taxonomyEnt.CreatedAtEQ(timestamp),
					taxonomyEnt.IDLT(id),
				),
			),
		)
	default:
		return applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// getSubTaxonomy - get sub taxonomies.
func (r *taxonomyRepository) getSubTaxonomy(ctx context.Context, rootID string, builder *ent.TaxonomyQuery) ([]*ent.Taxonomy, error) {
	// set where conditions
	builder.Where(
		taxonomyEnt.Or(
			taxonomyEnt.ID(rootID),
			taxonomyEnt.ParentIDHasPrefix(rootID),
		),
	)

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// executeArrayQuery - execute the builder query and return results.
func (r *taxonomyRepository) executeArrayQuery(ctx context.Context, builder *ent.TaxonomyQuery) ([]*ent.Taxonomy, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}
