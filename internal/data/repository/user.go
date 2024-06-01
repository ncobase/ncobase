package repo

import (
	"context"
	"fmt"
	"net/url"
	"stocms/internal/data"
	"stocms/internal/data/cache"
	"stocms/internal/data/ent"
	"stocms/internal/data/ent/user"
	userProfileEnt "stocms/internal/data/ent/userprofile"
	"stocms/internal/data/structs"
	"stocms/pkg/log"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// User represents the user repository interface.
type User interface {
	Create(ctx context.Context, body *structs.UserRequestBody) (*ent.User, error)
	CreateProfile(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	Find(ctx context.Context, m *structs.FindUser) (*ent.User, error)
	GetProfile(ctx context.Context, id string) (*ent.UserProfile, error)
	Existed(ctx context.Context, m *structs.FindUser) bool
	Delete(ctx context.Context, id string) error
}

// userRepo implements the User interface.
type userRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.User]
}

// NewUser creates a new user repository.
func NewUser(d *data.Data) User {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userRepo{ec, rc, cache.NewCache[ent.User](rc, cache.Key("user"))}
}

// Create - Create user
func (r *userRepo) Create(ctx context.Context, body *structs.UserRequestBody) (*ent.User, error) {
	countUser := r.ec.User.Query().CountX(ctx)

	row, err := r.ec.User.
		Create().
		SetUsername(body.Username).
		SetEmail(body.Email).
		SetPhone(body.Phone).
		SetIsCertified(true).
		SetIsAdmin(countUser < 1).
		Save(ctx)

	if err != nil {
		log.Errorf(nil, "userRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// CreateProfile - Create user profile
func (r *userRepo) CreateProfile(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error) {
	row, err := r.ec.UserProfile.
		Create().
		SetID(body.UserID).
		SetDisplayName(body.DisplayName).
		SetShortBio(body.ShortBio).
		Save(ctx)

	if err != nil {
		log.Errorf(nil, "userRepo.CreateProfile error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID - Get user by id
func (r *userRepo) GetByID(ctx context.Context, id string) (*ent.User, error) {
	cacheKey := fmt.Sprintf("%s", id)

	// Check cache first
	if cachedUser, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedUser, nil
	}

	// If not found in cache, query the database
	row, err := r.ec.User.
		Query().
		Where(user.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// Cache the result
	r.c.Set(ctx, cacheKey, row)

	return row, nil
}

// Find - Find user by username, email, or phone
func (r *userRepo) Find(ctx context.Context, m *structs.FindUser) (*ent.User, error) {
	params := url.Values{}
	if validator.IsNotEmpty(m.Username) {
		params.Set("username", m.Username)
	}
	if validator.IsNotEmpty(m.Email) {
		params.Set("email", m.Email)
	}
	if validator.IsNotEmpty(m.Phone) {
		params.Set("phone", m.Phone)
	}
	cacheKey := params.Encode()

	// Check cache first
	if cachedUser, err := r.c.Get(ctx, cacheKey); err == nil {
		return cachedUser, nil
	}

	// If not found in cache, query the database
	row, err := r.ec.User.
		Query().
		Where(user.Or(
			user.UsernameEQ(m.Username),
			user.EmailEQ(m.Email),
			user.PhoneEQ(m.Phone),
		)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userRepo.Find error: %v\n", err)
		return nil, err
	}

	// Cache the result
	r.c.Set(ctx, cacheKey, row)

	return row, nil
}

// GetProfile - Find profile by user id
func (r *userRepo) GetProfile(ctx context.Context, id string) (*ent.UserProfile, error) {
	row, err := r.ec.UserProfile.
		Query().
		Where(userProfileEnt.IDEQ(id)).
		Only(ctx)

	if err != nil {
		log.Errorf(nil, "userRepo.GetProfile error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// Existed - Verify user existed by username, email, or phone
func (r *userRepo) Existed(ctx context.Context, m *structs.FindUser) bool {
	return r.ec.User.Query().Where(user.Or(user.UsernameEQ(m.Username), user.EmailEQ(m.Email), user.PhoneEQ(m.Phone))).ExistX(ctx)
}

// Delete - Delete user
func (r *userRepo) Delete(ctx context.Context, id string) error {
	if err := r.ec.User.DeleteOneID(id).Exec(ctx); err != nil {
		log.Errorf(nil, "userRepo.Delete error: %v\n", err)
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	r.c.Delete(ctx, cacheKey)

	return nil
}
