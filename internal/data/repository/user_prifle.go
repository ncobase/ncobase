package repo

import (
	"context"
	"fmt"
	"ncobase/common/types"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	userProfileEnt "ncobase/internal/data/ent/userprofile"
	"ncobase/internal/data/structs"

	"ncobase/common/cache"
	"ncobase/common/log"

	"github.com/redis/go-redis/v9"
)

// UserProfile represents the user profile repository interface.
type UserProfile interface {
	Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.UserProfile, error)
	Get(ctx context.Context, id string) (*ent.UserProfile, error)
	Delete(ctx context.Context, id string) error
}

// userProfileRepo implements the User interface.
type userProfileRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.UserProfile]
}

// NewUserProfile creates a new user profile repository.
func NewUserProfile(d *data.Data) UserProfile {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userProfileRepo{ec, rc, cache.NewCache[ent.UserProfile](rc, cache.Key("nb_user_profile"))}
}

// Create create user profile
func (r *userProfileRepo) Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error) {

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
		log.Errorf(context.Background(), "userProfileRepo.CreateProfile error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, fmt.Sprintf("user_profile=%s", body.ID), row)
	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.CreateProfile cache error: %v\n", err)
	}

	return row, nil
}

// Update update user profile
func (r *userProfileRepo) Update(ctx context.Context, id string, updates types.JSON) (*ent.UserProfile, error) {

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
		log.Errorf(context.Background(), "userProfileRepo.UpdateProfile error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// Get find profile by user id
func (r *userProfileRepo) Get(ctx context.Context, id string) (*ent.UserProfile, error) {
	// try get from cache
	cacheKey := fmt.Sprintf("user_profile=%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	row, err := r.ec.UserProfile.
		Query().
		Where(userProfileEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.GetProfile error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.GetProfile cache error: %v\n", err)
	}

	return row, nil
}

// Delete delete user profile
func (r *userProfileRepo) Delete(ctx context.Context, id string) error {
	if _, err := r.ec.UserProfile.Delete().Where(userProfileEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userProfileRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("user_profile=%s", id)
	err := r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.Delete cache error: %v\n", err)
	}

	return nil
}
