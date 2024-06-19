package repo

import (
	"context"
	"fmt"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	taxonomyEnt "ncobase/internal/data/ent/taxonomy"
	"ncobase/internal/data/structs"
	"strings"

	"github.com/ncobase/common/cache"
	"github.com/ncobase/common/log"
	"github.com/ncobase/common/meili"
	"github.com/ncobase/common/types"
	"github.com/ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// Taxonomy represents the taxonomy repository interface.
type Taxonomy interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*ent.Taxonomy, error)
	GetByID(ctx context.Context, id string) (*ent.Taxonomy, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Taxonomy, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Taxonomy, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error)
	Delete(ctx context.Context, slug string) error
	FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error)
	ListBuilder(ctx context.Context, params *structs.ListTaxonomyParams) (*ent.TaxonomyQuery, error)
	CountX(ctx context.Context, params *structs.ListTaxonomyParams) int
}

// taxonomyRepo implements the Taxonomy interface.
type taxonomyRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Taxonomy]
}

// NewTaxonomy creates a new taxonomy repository.
func NewTaxonomy(d *data.Data) Taxonomy {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &taxonomyRepo{ec, rc, ms, cache.NewCache[ent.Taxonomy](rc, cache.Key("nb_taxonomy"), true)}
}

// Create create taxonomy
func (r *taxonomyRepo) Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*ent.Taxonomy, error) {

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
	builder.SetKeywords(strings.Join(body.Keywords, ","))
	builder.SetNillableDescription(&body.Description)
	builder.SetStatus(body.Status)
	builder.SetNillableParentID(&body.ParentID)
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
func (r *taxonomyRepo) GetByID(ctx context.Context, id string) (*ent.Taxonomy, error) {
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
	row, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{ID: id})

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
func (r *taxonomyRepo) GetBySlug(ctx context.Context, slug string) (*ent.Taxonomy, error) {
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
	row, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Slug: slug})

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

// Update update taxonomy
func (r *taxonomyRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Taxonomy, error) {
	taxonomy, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Slug: slug})
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
			builder.SetKeywords(strings.Join(value.([]string), ","))
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
func (r *taxonomyRepo) List(ctx context.Context, params *structs.ListTaxonomyParams) ([]*ent.Taxonomy, error) {
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
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	// belong tenant
	if params.TenantID != "" {
		builder.Where(taxonomyEnt.TenantIDEQ(params.TenantID))
	}

	// type
	if params.Type != "" {
		builder.Where(taxonomyEnt.TypeEQ(params.Type))
	}

	// sort
	builder.Order(ent.Desc(taxonomyEnt.FieldCreatedAt))

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
func (r *taxonomyRepo) Delete(ctx context.Context, slug string) error {
	taxonomy, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{Slug: slug})
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
func (r *taxonomyRepo) FindTaxonomy(ctx context.Context, params *structs.FindTaxonomy) (*ent.Taxonomy, error) {
	// create builder.
	builder := r.ec.Taxonomy.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(taxonomyEnt.IDEQ(params.ID))
	}
	// support slug or ID
	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(taxonomyEnt.Or(
			taxonomyEnt.ID(params.Slug),
			taxonomyEnt.SlugEQ(params.Slug),
		))
	}
	if validator.IsNotEmpty(params.TenantID) {
		builder = builder.Where(taxonomyEnt.TenantIDEQ(params.TenantID))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder create list builder
func (r *taxonomyRepo) ListBuilder(ctx context.Context, params *structs.ListTaxonomyParams) (*ent.TaxonomyQuery, error) {
	// verify query params.
	var next *ent.Taxonomy
	if validator.IsNotEmpty(params.Cursor) {
		// query the address.
		row, err := r.FindTaxonomy(ctx, &structs.FindTaxonomy{
			ID: params.Cursor,
		})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Taxonomy.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(taxonomyEnt.CreatedAtLT(next.CreatedAt))
	}

	// match parent id.
	// default is root.
	if validator.IsEmpty(params.ParentID) {
		builder.Where(taxonomyEnt.Or(
			taxonomyEnt.ParentIDIsNil(),
			taxonomyEnt.ParentIDEQ(""),
			taxonomyEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(taxonomyEnt.ParentIDEQ(params.ParentID))
	}

	return builder, nil
}

// CountX gets a count of taxonomies.
func (r *taxonomyRepo) CountX(ctx context.Context, params *structs.ListTaxonomyParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}
