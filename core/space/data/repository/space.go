package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceEnt "ncobase/core/space/data/ent/space"
	"ncobase/core/space/structs"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	nd "github.com/ncobase/ncore/data"
	"github.com/ncobase/ncore/data/search"
)

// SpaceRepositoryInterface represents the space repository interface.
type SpaceRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceBody) (*ent.Space, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Space, error)
	GetByUser(ctx context.Context, user string) (*ent.Space, error)
	GetIDByUser(ctx context.Context, user string) (string, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.Space, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Space, error)
	List(ctx context.Context, params *structs.ListSpaceParams) ([]*ent.Space, error)
	Delete(ctx context.Context, id string) error
	DeleteByUser(ctx context.Context, id string) error
	CountX(ctx context.Context, params *structs.ListSpaceParams) int
}

// spaceRepository implements the SpaceRepositoryInterface.
type spaceRepository struct {
	data             *data.Data
	sc               *search.Client
	ec               *ent.Client
	spaceCache       cache.ICache[ent.Space]
	slugMappingCache cache.ICache[string] // Maps slug to space ID
	userMappingCache cache.ICache[string] // Maps user ID to space ID
	spaceTTL         time.Duration
}

// NewSpaceRepository creates a new space repository.
func NewSpaceRepository(d *data.Data) SpaceRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)
	sc := nd.NewSearchClient(d.Data)

	return &spaceRepository{
		data:             d,
		sc:               sc,
		ec:               d.GetMasterEntClient(),
		spaceCache:       cache.NewCache[ent.Space](redisClient, "ncse_space:spaces"),
		slugMappingCache: cache.NewCache[string](redisClient, "ncse_space:slug_mappings"),
		userMappingCache: cache.NewCache[string](redisClient, "ncse_space:user_mappings"),
		spaceTTL:         time.Hour * 4, // 4 hours cache TTL
	}
}

// Create create space
func (r *spaceRepository) Create(ctx context.Context, body *structs.CreateSpaceBody) (*ent.Space, error) {
	builder := r.ec.Space.Create()
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableType(&body.Type)
	builder.SetNillableTitle(&body.Title)
	builder.SetNillableURL(&body.URL)
	builder.SetNillableLogo(&body.Logo)
	builder.SetNillableLogoAlt(&body.LogoAlt)
	builder.SetNillableKeywords(&body.Keywords)
	builder.SetNillableCopyright(&body.Copyright)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableCreatedBy(body.CreatedBy)
	builder.SetNillableExpiredAt(body.ExpiredAt)

	if !validator.IsNil(body.Order) {
		builder.SetNillableOrder(body.Order)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	space, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceRepo.Create error: %v", err)
		return nil, err
	}

	// Create the space in Meilisearch index
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "spaces", Document: space}); err != nil {
			logger.Errorf(ctx, "spaceRepo.Create error creating Meilisearch index: %v", err)
		}
	}

	// Cache the space
	go r.cacheSpace(context.Background(), space)

	return space, nil
}

// GetBySlug get space by slug or id
func (r *spaceRepository) GetBySlug(ctx context.Context, slug string) (*ent.Space, error) {
	// Try to get space ID from slug mapping cache
	if spaceID, err := r.getSpaceIDBySlug(ctx, slug); err == nil && spaceID != "" {
		// Try to get from space cache
		cacheKey := fmt.Sprintf("id:%s", spaceID)
		if cached, err := r.spaceCache.Get(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to database
	row, err := r.FindSpace(ctx, &structs.FindSpace{Slug: slug})
	if err != nil {
		logger.Errorf(ctx, "spaceRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSpace(context.Background(), row)

	return row, nil
}

// GetByUser get space by user
func (r *spaceRepository) GetByUser(ctx context.Context, userID string) (*ent.Space, error) {
	// Try to get space ID from user mapping cache
	if spaceID, err := r.getSpaceIDByUser(ctx, userID); err == nil && spaceID != "" {
		// Try to get from space cache
		cacheKey := fmt.Sprintf("id:%s", spaceID)
		if cached, err := r.spaceCache.Get(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}
	}

	// Fallback to database
	row, err := r.FindSpace(ctx, &structs.FindSpace{User: userID})
	if err != nil {
		logger.Errorf(ctx, "spaceRepo.GetByUser error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheSpace(context.Background(), row)

	return row, nil
}

// GetIDByUser get space id by user id
func (r *spaceRepository) GetIDByUser(ctx context.Context, userID string) (string, error) {
	// Try cache first
	if spaceID, err := r.getSpaceIDByUser(ctx, userID); err == nil && spaceID != "" {
		return spaceID, nil
	}

	// Fallback to database
	id, err := r.ec.Space.
		Query().
		Where(spaceEnt.CreatedByEQ(userID)).
		OnlyID(ctx)

	if err != nil {
		logger.Errorf(ctx, "spaceRepo.GetIDByUser error: %v", err)
		return "", err
	}

	// Cache user to space mapping
	go func() {
		userKey := fmt.Sprintf("user:%s", userID)
		if err := r.userMappingCache.Set(context.Background(), userKey, &id, r.spaceTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache user mapping %s: %v", userID, err)
		}
	}()

	return id, nil
}

// GetByIDs retrieves multiple spaces by their IDs in a single query
func (r *spaceRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Space, error) {
	if len(ids) == 0 {
		return []*ent.Space{}, nil
	}

	// Remove duplicates
	uniqueIDs := make(map[string]bool)
	cleanIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" && !uniqueIDs[id] {
			uniqueIDs[id] = true
			cleanIDs = append(cleanIDs, id)
		}
	}

	if len(cleanIDs) == 0 {
		return []*ent.Space{}, nil
	}

	// Query all spaces in a single database call
	spaces, err := r.ec.Space.
		Query().
		Where(spaceEnt.IDIn(cleanIDs...)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "spaceRepo.GetByIDs error: %v", err)
		return nil, err
	}

	// Cache each space
	go func() {
		for _, space := range spaces {
			r.cacheSpace(context.Background(), space)
		}
	}()

	return spaces, nil
}

// Update update space
func (r *spaceRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Space, error) {
	space, err := r.FindSpace(ctx, &structs.FindSpace{Slug: slug})
	if err != nil {
		return nil, err
	}

	builder := space.Update()

	// Set values as in original implementation
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "url":
			builder.SetNillableURL(convert.ToPointer(value.(string)))
		case "logo":
			builder.SetNillableLogo(convert.ToPointer(value.(string)))
		case "logo_alt":
			builder.SetNillableLogoAlt(convert.ToPointer(value.(string)))
		case "keywords":
			builder.SetNillableKeywords(convert.ToPointer(value.(string)))
		case "copyright":
			builder.SetNillableCopyright(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "order":
			builder.SetOrder(int(value.(float64)))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		case "expired_at":
			adjustedTime, _ := convert.AdjustToEndOfDay(value)
			builder.SetNillableExpiredAt(&adjustedTime)
		}
	}

	updatedSpace, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if r.sc != nil {
		if err = r.sc.Index(ctx, &search.IndexRequest{Index: "spaces", Document: updatedSpace, DocumentID: updatedSpace.ID}); err != nil {
			logger.Errorf(ctx, "spaceRepo.Update error updating Meilisearch index: %v", err)
		}
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateSpaceCache(context.Background(), space)
		r.cacheSpace(context.Background(), updatedSpace)
	}()

	return updatedSpace, nil
}

// List get space list
func (r *spaceRepository) List(ctx context.Context, params *structs.ListSpaceParams) ([]*ent.Space, error) {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Is disabled
	builder.Where(spaceEnt.DisabledEQ(false))

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
				spaceEnt.Or(
					spaceEnt.CreatedAtGT(timestamp),
					spaceEnt.And(
						spaceEnt.CreatedAtEQ(timestamp),
						spaceEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				spaceEnt.Or(
					spaceEnt.CreatedAtLT(timestamp),
					spaceEnt.And(
						spaceEnt.CreatedAtEQ(timestamp),
						spaceEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(spaceEnt.FieldCreatedAt), ent.Asc(spaceEnt.FieldID))
	} else {
		builder.Order(ent.Desc(spaceEnt.FieldCreatedAt), ent.Desc(spaceEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceRepo.List error: %v", err)
		return nil, err
	}

	// Cache spaces in background
	go func() {
		for _, space := range rows {
			r.cacheSpace(context.Background(), space)
		}
	}()

	return rows, nil
}

// Delete delete space
func (r *spaceRepository) Delete(ctx context.Context, id string) error {
	space, err := r.FindSpace(ctx, &structs.FindSpace{Slug: id})
	if err != nil {
		return err
	}

	builder := r.ec.Space.Delete()
	if _, err = builder.Where(spaceEnt.IDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceRepo.Delete error: %v", err)
		return err
	}

	// Delete from Meilisearch index
	if r.sc != nil {
		if err = r.sc.Delete(ctx, "spaces", space.ID); err != nil {
			logger.Errorf(ctx, "spaceRepo.Delete index error: %v", err)
		}
	}

	// Invalidate cache
	go r.invalidateSpaceCache(context.Background(), space)

	return nil
}

// DeleteByUser delete space by user ID
func (r *spaceRepository) DeleteByUser(ctx context.Context, userID string) error {
	// Get space first for cache invalidation
	space, err := r.GetByUser(ctx, userID)
	if err != nil {
		return err
	}

	builder := r.ec.Space.Delete()
	if _, err := builder.Where(spaceEnt.CreatedByEQ(userID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceRepo.DeleteByUser error: %v", err)
		return err
	}

	// Delete from Meilisearch index
	if r.sc != nil {
		if err = r.sc.Delete(ctx, "spaces", space.ID); err != nil {
			logger.Errorf(ctx, "spaceRepo.DeleteByUser index error: %v", err)
		}
	}

	// Invalidate cache
	go r.invalidateSpaceCache(context.Background(), space)

	return nil
}

// CountX gets a count of spaces
func (r *spaceRepository) CountX(ctx context.Context, params *structs.ListSpaceParams) int {
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// FindSpace retrieves a space
func (r *spaceRepository) FindSpace(ctx context.Context, params *structs.FindSpace) (*ent.Space, error) {
	builder := r.ec.Space.Query()

	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(spaceEnt.Or(
			spaceEnt.IDEQ(params.Slug),
			spaceEnt.SlugEQ(params.Slug),
		))
	}
	if validator.IsNotEmpty(params.User) {
		builder = builder.Where(spaceEnt.CreatedByEQ(params.User))
	}

	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// listBuilder - create list builder
func (r *spaceRepository) listBuilder(_ context.Context, params *structs.ListSpaceParams) (*ent.SpaceQuery, error) {
	builder := r.ec.Space.Query()

	// Match belong user
	if validator.IsNotEmpty(params.User) {
		builder.Where(spaceEnt.CreatedByEQ(params.User))
	}

	return builder, nil
}

func (r *spaceRepository) cacheSpace(ctx context.Context, space *ent.Space) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", space.ID)
	if err := r.spaceCache.Set(ctx, idKey, space, r.spaceTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space by ID %s: %v", space.ID, err)
	}

	// Cache slug to ID mapping
	if space.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", space.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &space.ID, r.spaceTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", space.Slug, err)
		}
	}

	// Cache user to space ID mapping
	if space.CreatedBy != "" {
		userKey := fmt.Sprintf("user:%s", space.CreatedBy)
		if err := r.userMappingCache.Set(ctx, userKey, &space.ID, r.spaceTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache user mapping %s: %v", space.CreatedBy, err)
		}
	}
}

func (r *spaceRepository) invalidateSpaceCache(ctx context.Context, space *ent.Space) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", space.ID)
	if err := r.spaceCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space ID cache %s: %v", space.ID, err)
	}

	// Invalidate slug mapping
	if space.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", space.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", space.Slug, err)
		}
	}

	// Invalidate user mapping
	if space.CreatedBy != "" {
		userKey := fmt.Sprintf("user:%s", space.CreatedBy)
		if err := r.userMappingCache.Delete(ctx, userKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate user mapping cache %s: %v", space.CreatedBy, err)
		}
	}
}

func (r *spaceRepository) getSpaceIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	spaceID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || spaceID == nil {
		return "", err
	}
	return *spaceID, nil
}

func (r *spaceRepository) getSpaceIDByUser(ctx context.Context, userID string) (string, error) {
	cacheKey := fmt.Sprintf("user:%s", userID)
	spaceID, err := r.userMappingCache.Get(ctx, cacheKey)
	if err != nil || spaceID == nil {
		return "", err
	}
	return *spaceID, nil
}
