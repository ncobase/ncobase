package repo

import (
	"context"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	assetEnt "stocms/internal/data/ent/asset"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/meili"
	"stocms/pkg/types"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// Asset represents the asset repository interface.
type Asset interface {
	Create(ctx context.Context, body *structs.CreateAssetBody) (*ent.Asset, error)
	GetByID(ctx context.Context, slug string) (*ent.Asset, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Asset, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, p *structs.ListAssetParams) ([]*ent.Asset, error)
}

// assetRepo implements the Asset interface.
type assetRepo struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Asset]
}

// NewAsset creates a new asset repository.
func NewAsset(d *data.Data) Asset {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &assetRepo{ec, rc, ms, cache.NewCache[ent.Asset](rc, cache.Key("sc_asset"), true)}
}

// Create creates an asset.
func (r *assetRepo) Create(ctx context.Context, body *structs.CreateAssetBody) (*ent.Asset, error) {

	// create builder.
	builder := r.ec.Asset.Create()
	// set values.

	builder.SetNillableName(&body.Name)
	builder.SetNillablePath(&body.Path)
	builder.SetNillableType(&body.Type)
	builder.SetNillableSize(body.Size)
	builder.SetNillableStorage(&body.Storage)
	builder.SetNillableBucket(&body.Bucket)
	builder.SetNillableEndpoint(&body.Endpoint)
	builder.SetNillableObjectID(&body.ObjectID)
	builder.SetNillableDomainID(&body.DomainID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(ctx, "assetRepo.Create error: %v\n", err)
		return nil, err
	}

	// create the asset in Meilisearch index
	if err = r.ms.IndexDocuments("assets", row); err != nil {
		log.Errorf(ctx, "assetRepo.Create index error: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// GetByID gets an asset by ID.
func (r *assetRepo) GetByID(ctx context.Context, slug string) (*ent.Asset, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindAsset(ctx, &structs.FindAsset{ID: slug})
	if err != nil {
		log.Errorf(ctx, "assetRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(ctx, "assetRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// Update updates an asset by ID.
func (r *assetRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Asset, error) {
	asset, err := r.FindAsset(ctx, &structs.FindAsset{ID: slug})
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := r.ec.Asset.UpdateOne(asset)

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
		case "endpoint":
			builder.SetNillableEndpoint(types.ToPointer(value.(string)))
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
		log.Errorf(ctx, "assetRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", asset.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "assetRepo.Update cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("assets", asset.ID); err != nil {
		log.Errorf(nil, "assetRepo.Update index error: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// Delete deletes an asset by ID.
func (r *assetRepo) Delete(ctx context.Context, slug string) error {
	asset, err := r.FindAsset(ctx, &structs.FindAsset{ID: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Asset.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(assetEnt.IDEQ(slug)).Exec(ctx); err != nil {
		log.Errorf(nil, "assetRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", asset.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "assetRepo.Delete cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("assets", asset.ID); err != nil {
		log.Errorf(nil, "assetRepo.Delete index error: %v\n", err)
		// return nil, err
	}

	return err
}

// FindAsset finds an asset.
func (r *assetRepo) FindAsset(ctx context.Context, p *structs.FindAsset) (*ent.Asset, error) {
	// create builder.
	builder := r.ec.Asset.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(assetEnt.Or(
			assetEnt.IDEQ(p.ID),
			assetEnt.NameEQ(p.ID),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// List gets a list of assets.
func (r *assetRepo) List(ctx context.Context, p *structs.ListAssetParams) ([]*ent.Asset, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(p.Limit))

	// belong domain
	if p.DomainID != "" {
		builder.Where(assetEnt.DomainIDEQ(p.DomainID))
	}

	// belong user
	if p.UserID != "" {
		builder.Where(assetEnt.CreatedByEQ(p.UserID))
	}

	// object id
	if p.ObjectID != "" {
		builder.Where(assetEnt.ObjectIDEQ(p.ObjectID))
	}

	// asset type
	if p.Type != "" {
		builder.Where(assetEnt.TypeEQ(p.Type))
	}

	// storage provider
	if p.Storage != "" {
		builder.Where(assetEnt.StorageEQ(p.Storage))
	}

	// sort
	builder.Order(ent.Desc(assetEnt.FieldCreatedAt))

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		log.Errorf(nil, "assetRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder.
func (r *assetRepo) ListBuilder(ctx context.Context, p *structs.ListAssetParams) (*ent.AssetQuery, error) {
	var next *ent.Asset
	if validator.IsNotEmpty(p.Cursor) {
		row, err := r.FindAsset(ctx, &structs.FindAsset{ID: p.Cursor})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create builder.
	builder := r.ec.Asset.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(assetEnt.CreatedAtLT(next.CreatedAt))
	}

	return builder, nil
}
