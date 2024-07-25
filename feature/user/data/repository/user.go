package repository

import (
	"context"
	"fmt"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/feature/user/data"
	"ncobase/feature/user/data/ent"
	userEnt "ncobase/feature/user/data/ent/user"
	"ncobase/feature/user/structs"
	"net/url"

	"ncobase/common/cache"
	"ncobase/common/crypto"
	"ncobase/common/log"
	"ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// UserRepositoryInterface represents the user repository interface.
type UserRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserBody) (*ent.User, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	Find(ctx context.Context, m *structs.FindUser) (*ent.User, error)
	Existed(ctx context.Context, m *structs.FindUser) bool
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListUserParams) ([]*ent.User, error)
	UpdatePassword(ctx context.Context, params *structs.UserPassword) error
	FindUser(ctx context.Context, params *structs.FindUser) (*ent.User, error) // not use cache
	CountX(ctx context.Context, params *structs.ListUserParams) int
}

// userRepository implements the UserRepositoryInterface.
type userRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.User]
}

// NewUserRepository creates a new user repository.
func NewUserRepository(d *data.Data) UserRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &userRepository{ec, rc, cache.NewCache[ent.User](rc, "ncse_user")}
}

// Create create user
func (r *userRepository) Create(ctx context.Context, body *structs.UserBody) (*ent.User, error) {
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
func (r *userRepository) GetByID(ctx context.Context, id string) (*ent.User, error) {
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
func (r *userRepository) Find(ctx context.Context, m *structs.FindUser) (*ent.User, error) {
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
func (r *userRepository) Existed(ctx context.Context, m *structs.FindUser) bool {
	return r.ec.User.Query().Where(userEnt.Or(userEnt.UsernameEQ(m.Username), userEnt.EmailEQ(m.Email), userEnt.PhoneEQ(m.Phone))).ExistX(ctx)
}

// Delete delete user
func (r *userRepository) Delete(ctx context.Context, id string) error {
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
func (r *userRepository) UpdatePassword(ctx context.Context, params *structs.UserPassword) error {
	row, err := r.FindUser(ctx, &structs.FindUser{Username: params.User})
	if validator.IsNotNil(err) {
		return err
	}

	builder := row.Update()

	ph, err := crypto.HashPassword(ctx, params.NewPassword)
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
func (r *userRepository) FindUser(ctx context.Context, params *structs.FindUser) (*ent.User, error) {

	// create builder.
	builder := r.ec.User.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(userEnt.IDEQ(params.ID))
	}

	if validator.IsNotEmpty(params.Username) {
		// username value could be id, username, email, or phone
		builder = builder.Where(userEnt.Or(
			userEnt.IDEQ(params.Username),
			userEnt.UsernameEQ(params.Username),
			userEnt.EmailEQ(params.Username),
			userEnt.PhoneEQ(params.Username),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// CountX gets a count of users.
func (r *userRepository) CountX(ctx context.Context, params *structs.ListUserParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// List gets a list of users.
func (r *userRepository) List(ctx context.Context, params *structs.ListUserParams) ([]*ent.User, error) {
	builder := r.ec.User.Query()

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
				userEnt.Or(
					userEnt.CreatedAtGT(timestamp),
					userEnt.And(
						userEnt.CreatedAtEQ(timestamp),
						userEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				userEnt.Or(
					userEnt.CreatedAtLT(timestamp),
					userEnt.And(
						userEnt.CreatedAtEQ(timestamp),
						userEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(userEnt.FieldCreatedAt), ent.Asc(userEnt.FieldID))
	} else {
		builder.Order(ent.Desc(userEnt.FieldCreatedAt), ent.Desc(userEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ****** Internal methods of repository
// listBuilder creates list builder.
func (r *userRepository) listBuilder(_ context.Context, _ *structs.ListUserParams) (*ent.UserQuery, error) {
	// create builder.
	builder := r.ec.User.Query()

	return builder, nil
}
