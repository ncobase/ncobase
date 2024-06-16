package repo

import (
	"context"
	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	userProfileEnt "ncobase/internal/data/ent/userprofile"
	"ncobase/internal/data/structs"

	"github.com/redis/go-redis/v9"
)

// UserProfile represents the user profile repository interface.
type UserProfile interface {
	Create(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error)
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
	return &userProfileRepo{ec, rc, cache.NewCache[ent.UserProfile](rc, cache.Key("nb_user_profile"), true)}
}

// Create create user profile
func (r *userProfileRepo) Create(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error) {

	// create builder.
	builder := r.ec.UserProfile.Create()
	// set values.
	builder.SetID(body.UserID)
	builder.SetDisplayName(body.DisplayName)
	builder.SetShortBio(body.ShortBio)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.CreateProfile error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// Get find profile by user id
func (r *userProfileRepo) Get(ctx context.Context, id string) (*ent.UserProfile, error) {
	row, err := r.ec.UserProfile.
		Query().
		Where(userProfileEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(context.Background(), "userProfileRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// Delete delete user profile
func (r *userProfileRepo) Delete(ctx context.Context, id string) error {
	if _, err := r.ec.UserProfile.Delete().Where(userProfileEnt.IDEQ(id)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userProfileRepo.Delete error: %v\n", err)
		return err
	}
	return nil
}
