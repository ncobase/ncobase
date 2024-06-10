package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	userProfileEnt "stocms/internal/data/ent/userprofile"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"

	"github.com/redis/go-redis/v9"
)

// UserProfile represents the user profile repository interface.
type UserProfile interface {
	Create(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error)
	GetByUserID(ctx context.Context, id string) (*ent.UserProfile, error)
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
	return &userProfileRepo{ec, rc, cache.NewCache[ent.UserProfile](rc, cache.Key("sc_user_profile"), true)}
}

// Create - Create user profile
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
		log.Errorf(nil, "userProfileRepo.CreateProfile error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByUserID - Find profile by user id
func (r *userProfileRepo) GetByUserID(ctx context.Context, id string) (*ent.UserProfile, error) {
	row, err := r.ec.UserProfile.
		Query().
		Where(userProfileEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userProfileRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}
