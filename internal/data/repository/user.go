package repo

import (
	"context"
	"fmt"
	"net/url"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	userEnt "stocms/internal/data/ent/user"
	userProfileEnt "stocms/internal/data/ent/userprofile"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/crypto"
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
	UpdatePassword(ctx context.Context, p *structs.UserRequestBody) error
	FindUser(ctx context.Context, p *structs.FindUser) (*ent.User, error) // not use cache
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
	return &userRepo{ec, rc, cache.NewCache[ent.User](rc, cache.Key("sc_user"), true)}
}

// Create - Create user
func (r *userRepo) Create(ctx context.Context, body *structs.UserRequestBody) (*ent.User, error) {
	countUser := r.ec.User.Query().CountX(ctx)

	// create builder.
	builder := r.ec.User.Create()
	// Set values
	builder.SetUsername(body.Username)
	builder.SetEmail(body.Email)
	builder.SetPhone(body.Phone)
	builder.SetIsCertified(true)
	builder.SetIsAdmin(countUser < 1)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(nil, "userRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// CreateProfile - Create user profile
func (r *userRepo) CreateProfile(ctx context.Context, body *structs.UserRequestBody) (*ent.UserProfile, error) {
	// create builder.
	builder := r.ec.UserProfile.Create()
	// Set values
	builder.SetID(body.UserID)
	builder.SetDisplayName(body.DisplayName)
	builder.SetShortBio(body.ShortBio)

	// execute the builder.
	row, err := builder.Save(ctx)
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
	row, err := r.FindUser(ctx, &structs.FindUser{ID: id})

	if err != nil {
		log.Errorf(nil, "userRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "userRepo.GetByID cache error: %v\n", err)
	}

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
	row, err := r.FindUser(ctx, m)

	if err != nil {
		log.Errorf(nil, "userRepo.Find error: %v\n", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "userRepo.Find cache error: %v\n", err)
	}

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
	return r.ec.User.Query().Where(userEnt.Or(userEnt.UsernameEQ(m.Username), userEnt.EmailEQ(m.Email), userEnt.PhoneEQ(m.Phone))).ExistX(ctx)
}

// Delete - Delete user
func (r *userRepo) Delete(ctx context.Context, id string) error {
	if err := r.ec.User.DeleteOneID(id).Exec(ctx); err != nil {
		log.Errorf(nil, "userRepo.Delete error: %v\n", err)
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err := r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "userRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// UpdatePassword - update user password.
func (r *userRepo) UpdatePassword(ctx context.Context, p *structs.UserRequestBody) error {
	row, err := r.FindUser(ctx, &structs.FindUser{ID: p.UserID})
	if validator.IsNotNil(err) {
		return err
	}

	// create builder.
	builder := row.Update()

	ph, err := crypto.HashPassword(ctx, p.NewPassword)
	if validator.IsNotNil(err) {
		return err
	}

	builder.SetPassword(ph)

	// execute the builder.
	_, err = builder.Save(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	return nil
}

// FindUser - Find user by id, username, email, or phone
func (r *userRepo) FindUser(ctx context.Context, p *structs.FindUser) (*ent.User, error) {
	// create builder.
	builder := r.ec.User.Query()

	if validator.IsNotEmpty(p.ID) {
		builder = builder.Where(userEnt.IDEQ(p.ID))
	}

	if validator.IsNotEmpty(p.Username) {
		// username value could be id, username, email, or phone
		builder = builder.Where(userEnt.Or(
			userEnt.IDEQ(p.Username),
			userEnt.UsernameEQ(p.Username),
			userEnt.EmailEQ(p.Username),
			userEnt.PhoneEQ(p.Username),
		))
	}
	if validator.IsNotEmpty(p.Email) {
		builder = builder.Where(userEnt.EmailEQ(p.Email))
	}
	if validator.IsNotEmpty(p.Phone) {
		builder = builder.Where(userEnt.PhoneEQ(p.Phone))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ****** Internal methods of repository
