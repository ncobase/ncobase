package repository

import (
	"context"
	"fmt"
	"ncobase/user/data"
	"ncobase/user/data/ent"
	userProfileEnt "ncobase/user/data/ent/userprofile"
	"ncobase/user/structs"

	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/logging/logger"

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
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &userProfileRepository{ec, rc, cache.NewCache[ent.UserProfile](rc, "ncse_user_profile")}
}

// Create create user profile
func (r *userProfileRepository) Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error) {

	// create builder.
	builder := r.ec.UserProfile.Create()
	// set values.
	builder.SetID(body.UserID)
	builder.SetDisplayName(body.DisplayName)
	builder.SetNillableFirstName(&body.FirstName)
	builder.SetNillableLastName(&body.LastName)
	builder.SetNillableTitle(&body.Title)
	builder.SetShortBio(body.ShortBio)
	builder.SetNillableAbout(body.About)
	builder.SetNillableThumbnail(body.Thumbnail)

	if body.Links != nil {
		builder.SetLinks(convert.ToValue(body.Links))
	}

	if body.Extras != nil {
		builder.SetExtras(convert.ToValue(body.Extras))
	}

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "userProfileRepo.CreateProfile error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, fmt.Sprintf("%s", body.UserID), row)
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
			builder.SetNillableDisplayName(convert.ToPointer(value.(string)))
		case "first_name":
			builder.SetNillableFirstName(convert.ToPointer(value.(string)))
		case "last_name":
			builder.SetNillableLastName(convert.ToPointer(value.(string)))
		case "title":
			builder.SetNillableTitle(convert.ToPointer(value.(string)))
		case "short_bio":
			builder.SetNillableShortBio(convert.ToPointer(value.(string)))
		case "about":
			builder.SetNillableAbout(convert.ToPointer(value.(string)))
		case "thumbnail":
			builder.SetNillableThumbnail(convert.ToPointer(value.(string)))
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
