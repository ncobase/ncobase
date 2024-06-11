package repo

import (
	"context"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	resourceEnt "stocms/internal/data/ent/resource"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/meili"
	"stocms/pkg/types"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// Resource represents the resource repository interface.
type Resource interface {
	Create(ctx context.Context, body *structs.CreateResourceBody) (*ent.Resource, error)
	GetByID(ctx context.Context, slug string) (*ent.Resource, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Resource, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, p *structs.ListResourceParams) ([]*ent.Resource, error)
}

// resourceRepo implements the Resource interface.
type resourceRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Resource]
}

// NewResource creates a new resource repository.
func NewResource(d *data.Data) Resource {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &resourceRepo{ec, rc, ms, cache.NewCache[ent.Resource](rc, cache.Key("sc_resource"), true)}
}

// Create creates an resource.
func (r *resourceRepo) Create(ctx context.Context, body *structs.CreateResourceBody) (*ent.Resource, error) {

	// create builder.
	builder := r.ec.Resource.Create()
	// set values.

	builder.SetNillableName(&body.Name)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSize(body.Size)
	builder.SetNillableStorage(&body.Storage)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableObjectID(&body.ObjectID)
	builder.SetNillableDomainID(&body.DomainID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "resourceRepo.Create error: %v\n", err)
		return nil, err
	}

	// create the resource in Meilisearch index
	if err = r.ms.IndexDocuments("resources", row); err != nil {
		log.Errorf(ctx, "resourceRepo.Create index error: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets an resource by ID.
func (r *resourceRepo) GetByID(ctx context.Context, slug string) (*ent.Resource, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindResource(ctx, &structs.FindResource{ID: slug})
	if err != nil {
		log.Errorf(ctx, "resourceRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(ctx, "resourceRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// Update updates an resource by ID.
func (r *resourceRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Resource, error) {
	resource, err := r.FindResource(ctx, &structs.FindResource{ID: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := r.ec.Resource.UpdateOne(resource)

	// set values
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "path":
			builder.SetNillablePath(types.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(types.ToPointer(value.(string)))
		case "size":
			builder.SetNillableSize(types.ToPointer(value.(int64)))
		case "storage":
			builder.SetNillableStorage(types.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(types.ToPointer(value.(string)))
		case "object_id":
			builder.SetNillableObjectID(types.ToPointer(value.(string)))
		case "domain_id":
			builder.SetNillableDomainID(types.ToPointer(value.(string)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "resourceRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", resource.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "resourceRepo.Update cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("resources", resource.ID); err != nil {
		log.Errorf(nil, "resourceRepo.Update index error: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// Delete deletes an resource by ID.
func (r *resourceRepo) Delete(ctx context.Context, slug string) error {
	resource, err := r.FindResource(ctx, &structs.FindResource{ID: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Resource.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(resourceEnt.IDEQ(slug)).Exec(ctx); err != nil {
		log.Errorf(nil, "resourceRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", resource.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "resourceRepo.Delete cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("resources", resource.ID); err != nil {
		log.Errorf(nil, "resourceRepo.Delete index error: %v\n", err)
		// return nil, err
	}

	return err
}

// FindResource finds an resource.
func (r *resourceRepo) FindResource(ctx context.Context, p *structs.FindResource) (*ent.Resource, error) {
	// create builder.
	builder := r.ec.Resource.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(resourceEnt.Or(
			resourceEnt.IDEQ(p.ID),
			resourceEnt.NameEQ(p.ID),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets a list of resources.
func (r *resourceRepo) List(ctx context.Context, p *structs.ListResourceParams) ([]*ent.Resource, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(p.Limit))

	// belong domain
	if p.Domain != "" {
		builder.Where(resourceEnt.DomainIDEQ(p.Domain))
	}

	// belong user
	if p.User != "" {
		builder.Where(resourceEnt.CreatedByEQ(p.User))
	}

	// object id
	if p.Object != "" {
		builder.Where(resourceEnt.ObjectIDEQ(p.Object))
	}

	// resource type
	if p.Type != "" {
		builder.Where(resourceEnt.TypeEQ(p.Type))
	}

	// storage provider
	if p.Storage != "" {
		builder.Where(resourceEnt.StorageEQ(p.Storage))
	}

	// sort
	builder.Order(ent.Desc(resourceEnt.FieldCreatedAt))

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		log.Errorf(nil, "resourceRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder.
func (r *resourceRepo) ListBuilder(ctx context.Context, p *structs.ListResourceParams) (*ent.ResourceQuery, error) {
	var next *ent.Resource
	if validator.IsNotEmpty(p.Cursor) {
		row, err := r.FindResource(ctx, &structs.FindResource{ID: p.Cursor})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Resource.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(resourceEnt.CreatedAtLT(next.CreatedAt))
	}

	return builder, nil
}
