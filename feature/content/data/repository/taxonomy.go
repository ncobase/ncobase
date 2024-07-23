package repository

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/meili"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	taxonomyEnt "ncobase/feature/content/data/ent/taxonomy"
	"ncobase/feature/content/structs"

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
	Delete(ctx context.Context, slug string) error
	FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error)
	CountX(ctx context.Context, params *structs.ListTaxonomyParams) int
}

// taxonomyRepository implements the TaxonomyRepositoryInterface.
type taxonomyRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Taxonomy]
}

// NewTaxonomyRepository creates a new taxonomy repository.
func NewTaxonomyRepository(d *data.Data) TaxonomyRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &taxonomyRepository{ec, rc, ms, cache.NewCache[ent.Taxonomy](rc, "ncse_taxonomy")}
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
		log.Errorf(context.Background(), "taxonomyRepo.Create error: %v\n", err)
		return nil, err
	}

	// Create the taxonomy in Meilisearch index
	if err = r.ms.IndexDocuments("taxonomies", row); err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.Create error creating Meilisearch index: %v\n", err)
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
		log.Errorf(context.Background(), "taxonomyRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.GetByID cache error: %v\n", err)
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
		log.Errorf(context.Background(), "taxonomyRepo.GetBySlug error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.GetBySlug cache error: %v\n", err)
	}

	return row, nil
}

// GetTree retrieves the taxonomy tree.
func (r *taxonomyRepository) GetTree(ctx context.Context, params *structs.FindTaxonomy) ([]*ent.Taxonomy, error) {
	// create builder
	builder := r.ec.Taxonomy.Query()

	// set where conditions
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(taxonomyEnt.TenantIDEQ(params.Tenant))
	}

	// handle sub taxonomys
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
		log.Errorf(context.Background(), "taxonomyRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", taxonomy.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, taxonomy.Slug)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.Update cache error: %v\n", err)
	}

	// Update the taxonomy in Meilisearch
	if err = r.ms.DeleteDocuments("taxonomies", slug); err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.Update error deleting Meilisearch index: %v\n", err)
		// return nil, err
	}
	if err = r.ms.IndexDocuments("taxonomies", row, row.ID); err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.Update error updating Meilisearch index: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// List get taxonomy list
func (r *taxonomyRepository) List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error) {
	// Generate cache key based on query parameters
	// cacheKey := fmt.Sprintf("list_taxonomy_%s_%d_%s_%s", params.Cursor, params.Limit, params.TenantID, params.Type)

	// // check cache first
	// cachedResult, err := r.c.Get(ctx, cacheKey)
	// if err == nil {
	// 	// Convert cached JSON data back to []*ent.Taxonomy
	// 	var result []*ent.Taxonomy
	// 	if err := json.Unmarshal([]byte(cachedResult), &result); err != nil {
	// 		return nil, err
	// 	}
	// 	return result, nil
	// } else if err != cache.ErrCacheMiss {
	// 	log.Errorf(context.Background(), "taxonomyRepo.List cache error: %v\n", err)
	// }

	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
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
				taxonomyEnt.Or(
					taxonomyEnt.CreatedAtGT(timestamp),
					taxonomyEnt.And(
						taxonomyEnt.CreatedAtEQ(timestamp),
						taxonomyEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				taxonomyEnt.Or(
					taxonomyEnt.CreatedAtLT(timestamp),
					taxonomyEnt.And(
						taxonomyEnt.CreatedAtEQ(timestamp),
						taxonomyEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(taxonomyEnt.FieldCreatedAt), ent.Asc(taxonomyEnt.FieldID))
	} else {
		builder.Order(ent.Desc(taxonomyEnt.FieldCreatedAt), ent.Desc(taxonomyEnt.FieldID))
	}

	builder.Limit(params.Limit)

	// belong tenant
	if params.Tenant != "" {
		builder.Where(taxonomyEnt.TenantIDEQ(params.Tenant))
	}

	// type
	if params.Type == "all" {
		params.Type = ""
	}
	if params.Type != "" {
		builder.Where(taxonomyEnt.TypeEQ(params.Type))
	}

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.List error: %v\n", err)
		return nil, err
	}

	// // Convert []*ent.Taxonomy to JSON data
	// jsonData, err := json.Marshal(rows)
	// if err != nil {
	// 	log.Errorf(context.Background(), "taxonomyRepo.List cache error: %v\n", err)
	// } else {
	// 	// cache the result
	// 	if err := r.c.Set(ctx, cacheKey, jsonData); err != nil {
	// 		log.Errorf(context.Background(), "taxonomyRepo.List cache error: %v\n", err)
	// 	}
	// }

	return rows, nil
}

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
		log.Errorf(context.Background(), "taxonomyRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	err = r.c.Delete(ctx, taxonomy.ID)
	err = r.c.Delete(ctx, slug)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRepo.Delete cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("taxonomies", taxonomy.ID); err != nil {
		log.Errorf(context.Background(), "topicRepo.Delete index error: %v\n", err)
		// return nil, err
	}

	return err
}

// FindTaxonomy find taxonomy
func (r *taxonomyRepository) FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error) {
	// create builder.
	builder := r.ec.Taxonomy.Query()

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
	builder := r.ec.Taxonomy.Query()

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

// CountX gets a count of taxonomies.
func (r *taxonomyRepository) CountX(ctx context.Context, params *structs.ListTaxonomyParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// getSubTaxonomy - get sub taxonomys.
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
		log.Errorf(ctx, "taxonomyRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}
