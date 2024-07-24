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
	"ncobase/feature/resource/data"
	"ncobase/feature/resource/data/ent"
	assetEnt "ncobase/feature/resource/data/ent/asset"
	"ncobase/feature/resource/structs"

	"github.com/redis/go-redis/v9"
)

// AssetRepositoryInterface represents the asset repository interface.
type AssetRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateAssetBody) (*ent.Asset, error)
	GetByID(ctx context.Context, slug string) (*ent.Asset, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Asset, error)
	Delete(ctx context.Context, slug string) error
	List(ctx context.Context, params *structs.ListAssetParams) ([]*ent.Asset, error)
	CountX(ctx context.Context, params *structs.ListAssetParams) int
}

// assetRepostory implements the AssetRepositoryInterface.
type assetRepostory struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
	c  *cache.Cache[ent.Asset]
}

// NewAssetRepository creates a new asset repository.
func NewAssetRepository(d *data.Data) AssetRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &assetRepostory{ec, rc, ms, cache.NewCache[ent.Asset](rc, "ncse_asset")}
}

// Create creates an asset.
func (r *assetRepostory) Create(ctx context.Context, body *structs.CreateAssetBody) (*ent.Asset, error) {

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
	builder.SetNillableTenantID(&body.TenantID)
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
func (r *assetRepostory) GetByID(ctx context.Context, slug string) (*ent.Asset, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindAsset(ctx, &structs.FindAsset{Asset: slug})
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
func (r *assetRepostory) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Asset, error) {
	asset, err := r.FindAsset(ctx, &structs.FindAsset{Asset: slug})
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
			builder.SetNillableSize(types.ToPointer(value.(int)))
		case "storage":
			builder.SetNillableStorage(types.ToPointer(value.(string)))
		case "endpoint":
			builder.SetNillableEndpoint(types.ToPointer(value.(string)))
		case "object_id":
			builder.SetNillableObjectID(types.ToPointer(value.(string)))
		case "tenant_id":
			builder.SetNillableTenantID(types.ToPointer(value.(string)))
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
		log.Errorf(context.Background(), "assetRepo.Update index error: %v\n", err)
		// return nil, err
	}

	return row, nil
}

// Delete deletes an asset by ID.
func (r *assetRepostory) Delete(ctx context.Context, slug string) error {
	asset, err := r.FindAsset(ctx, &structs.FindAsset{Asset: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Asset.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(assetEnt.IDEQ(slug)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "assetRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", asset.ID)
	if err = r.c.Delete(ctx, cacheKey); err != nil {
		log.Errorf(ctx, "assetRepo.Delete cache error: %v\n", err)
	}

	// delete from Meilisearch index
	if err = r.ms.DeleteDocuments("assets", asset.ID); err != nil {
		log.Errorf(context.Background(), "assetRepo.Delete index error: %v\n", err)
		// return nil, err
	}

	return nil
}

// FindAsset finds an asset.
func (r *assetRepostory) FindAsset(ctx context.Context, params *structs.FindAsset) (*ent.Asset, error) {
	// create builder.
	builder := r.ec.Asset.Query()

	if validator.IsNotEmpty(params.Asset) {
		builder = builder.Where(assetEnt.Or(
			assetEnt.IDEQ(params.Asset),
			assetEnt.NameEQ(params.Asset),
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
func (r *assetRepostory) List(ctx context.Context, params *structs.ListAssetParams) ([]*ent.Asset, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
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
				assetEnt.Or(
					assetEnt.CreatedAtGT(timestamp),
					assetEnt.And(
						assetEnt.CreatedAtEQ(timestamp),
						assetEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				assetEnt.Or(
					assetEnt.CreatedAtLT(timestamp),
					assetEnt.And(
						assetEnt.CreatedAtEQ(timestamp),
						assetEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(assetEnt.FieldCreatedAt), ent.Asc(assetEnt.FieldID))
	} else {
		builder.Order(ent.Desc(assetEnt.FieldCreatedAt), ent.Desc(assetEnt.FieldID))
	}

	builder.Limit(params.Limit)

	// execute the builder.
	rows, err := builder.All(ctx)
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "assetRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// ListBuilder creates list builder.
func (r *assetRepostory) ListBuilder(ctx context.Context, params *structs.ListAssetParams) (*ent.AssetQuery, error) {
	// create builder.
	builder := r.ec.Asset.Query()

	// belong tenant
	if params.Tenant != "" {
		builder = builder.Where(assetEnt.TenantIDEQ(params.Tenant))
	}

	// belong user
	if params.User != "" {
		builder = builder.Where(assetEnt.CreatedByEQ(params.User))
	}

	// object id
	if params.Object != "" {
		builder = builder.Where(assetEnt.ObjectIDEQ(params.Object))
	}

	// asset type
	if params.Type != "" {
		builder = builder.Where(assetEnt.TypeContains(params.Type))
	}

	// storage provider
	if params.Storage != "" {
		builder = builder.Where(assetEnt.StorageEQ(params.Storage))
	}

	return builder, nil
}

// CountX counts assets based on given parameters.
func (r *assetRepostory) CountX(ctx context.Context, params *structs.ListAssetParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}
