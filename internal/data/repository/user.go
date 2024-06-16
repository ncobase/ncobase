package repo

import (
	"context"
	"fmt"
	"ncobase/common/cache"
	"ncobase/common/crypto"
	"ncobase/common/log"
	"ncobase/common/validator"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	userEnt "ncobase/internal/data/ent/user"
	"ncobase/internal/data/structs"
	"net/url"

	"github.com/redis/go-redis/v9"
)

// User represents the user repository interface.
type User interface {
	Create(ctx context.Context, body *structs.UserRequestBody) (*ent.User, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	Find(ctx context.Context, m *structs.FindUser) (*ent.User, error)
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
	return &userRepo{ec, rc, cache.NewCache[ent.User](rc, cache.Key("nb_user"), true)}
}

// Create create user
func (r *userRepo) Create(ctx context.Context, body *structs.UserRequestBody) (*ent.User, error) {
	countUser := r.ec.User.Query().CountX(ctx)

	// create builder.
	builder := r.ec.User.Create()
	// set values.
	builder.SetUsername(body.Username)
	builder.SetEmail(body.Email)
	builder.SetPhone(body.Phone)
	builder.SetIsCertified(true)
	builder.SetIsAdmin(countUser < 1)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "userRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID get user by id
func (r *userRepo) GetByID(ctx context.Context, id string) (*ent.User, error) {
	cacheKey := fmt.Sprintf("%s", id)

	// check cache first
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindUser(ctx, &structs.FindUser{ID: id})

	if err != nil {
		log.Errorf(context.Background(), "userRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "userRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// Find user by username, email, or phone
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

	// check cache first
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindUser(ctx, m)

	if err != nil {
		log.Errorf(context.Background(), "userRepo.Find error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "userRepo.Find cache error: %v\n", err)
	}

	return row, nil
}

// Existed verify user existed by username, email, or phone
func (r *userRepo) Existed(ctx context.Context, m *structs.FindUser) bool {
	return r.ec.User.Query().Where(userEnt.Or(userEnt.UsernameEQ(m.Username), userEnt.EmailEQ(m.Email), userEnt.PhoneEQ(m.Phone))).ExistX(ctx)
}

// Delete delete user
func (r *userRepo) Delete(ctx context.Context, id string) error {
	if err := r.ec.User.DeleteOneID(id).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "userRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", id)
	err := r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "userRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// UpdatePassword  update user password.
func (r *userRepo) UpdatePassword(ctx context.Context, p *structs.UserRequestBody) error {
	row, err := r.FindUser(ctx, &structs.FindUser{ID: p.UserID})
	if validator.IsNotNil(err) {
		return err
	}

	builder := row.Update()

	ph, err := crypto.HashPassword(ctx, p.NewPassword)
	if validator.IsNotNil(err) {
		return err
	}

	builder.SetPassword(ph)

	_, err = builder.Save(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	return nil
}

// FindUser find user by id, username, email, or phone
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

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ****** Internal methods of repository
