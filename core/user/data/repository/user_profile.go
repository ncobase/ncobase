package repository

import (
	"context"
	"fmt"
	"ncobase/common/types"
	"ncobase/core/user/data"
	"ncobase/core/user/data/ent"
	userProfileEnt "ncobase/core/user/data/ent/userprofile"
	"ncobase/core/user/structs"

	"ncobase/common/data/cache"
	"ncobase/common/logger"

	"github.com/redis/go-redis/v9"
)

// UserProfileRepositoryInterface represents the user profile repository interface.
type UserProfileRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.UserProfile, error)
	Get(ctx context.Context, id string) (*ent.UserProfile, error)
	Delete(ctx context.Context, id string) error
}

// userProfileRepository implements the UserProfileRepositoryInterface.
type userProfileRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserProfile]
}

// NewUserProfileRepository creates a new user profile repository.
func NewUserProfileRepository(d *data.Data) UserProfileRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userProfileRepository{ec, rc, cache.NewCache[ent.UserProfile](rc, "ncse_user_profile")}
}

// Create create user profile
func (r *userProfileRepository) Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error) {

	// create builder.
	builder := r.ec.UserProfile.Create()
	// set values.
	builder.SetID(body.ID)
	builder.SetDisplayName(body.DisplayName)
	builder.SetShortBio(body.ShortBio)
	builder.SetNillableAbout(body.About)
	builder.SetNillableThumbnail(body.Thumbnail)

	if body.Links != nil {
		builder.SetLinks(types.ToValue(body.Links))
	}

	if body.Extras != nil {
		builder.SetExtras(types.ToValue(body.Extras))
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.CreateProfile error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, fmt.Sprintf("%s", body.ID), row)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.CreateProfile cache error: %v", err)
	}

	return row, nil
}

// Update update user profile
func (r *userProfileRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.UserProfile, error) {

	// create builder
	builder := r.ec.UserProfile.UpdateOneID(id)

	for field, value := range updates {
		switch field {
		case "display_name":
			builder.SetNillableDisplayName(types.ToPointer(value.(string)))
		case "short_bio":
			builder.SetNillableShortBio(types.ToPointer(value.(string)))
		case "about":
			builder.SetNillableAbout(types.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(types.ToPointer(value.(string)))
		case "links":
			builder.SetLinks(value.([]types.JSON))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		}
	}

	// execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.UpdateProfile error: %v", err)
		return nil, err
	}

	return row, nil
}

// Get find profile by user id
func (r *userProfileRepository) Get(ctx context.Context, id string) (*ent.UserProfile, error) {
	// try get from cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// create builder.
	builder := r.ec.UserProfile.Query()
	// set conditions.
	builder.Where(userProfileEnt.IDEQ(id))
	// execute the builder.
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.GetProfile error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.GetProfile cache error: %v", err)
	}

	return row, nil
}

// Delete delete user profile
func (r *userProfileRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.ec.UserProfile.Delete().Where(userProfileEnt.IDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "userProfileRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err := r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.Delete cache error: %v", err)
	}

	return nil
}
