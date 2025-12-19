package repository

import (
	"context"
	"fmt"
	"ncobase/user/data"
	"ncobase/user/data/ent"
	userEnt "ncobase/user/data/ent/user"
	"ncobase/user/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/crypto"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"

	"github.com/ncobase/ncore/data/search"
)

// UserRepositoryInterface defines user repository operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserBody) (*ent.User, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.User, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListUserParams) ([]*ent.User, error)
	Find(ctx context.Context, filter *structs.FindUser) (*ent.User, error)
	FindUser(ctx context.Context, filter *structs.FindUser) (*ent.User, error)
	UpdatePassword(ctx context.Context, body *structs.UserPassword) error
	UpdatePasswordByID(ctx context.Context, userID, hashedPassword string) error
	CountX(ctx context.Context, params *structs.ListUserParams) int
}

// userRepository implements UserRepositoryInterface
type userRepository struct {
	data                 *data.Data
	userCache            cache.ICache[ent.User]
	usernameMappingCache cache.ICache[string] // Maps username to user ID
	emailMappingCache    cache.ICache[string] // Maps email to user ID
	userTTL              time.Duration
}

// NewUserRepository creates a new user repository
func NewUserRepository(d *data.Data) UserRepositoryInterface {
	redisClient := d.GetRedis()

	return &userRepository{
		data:                 d,
		userCache:            cache.NewCache[ent.User](redisClient, "ncse_users"),
		usernameMappingCache: cache.NewCache[string](redisClient, "ncse_user_mappings:username"),
		emailMappingCache:    cache.NewCache[string](redisClient, "ncse_user_mappings:email"),
		userTTL:              time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, body *structs.UserBody) (*ent.User, error) {
	id := nanoid.PrimaryKey()()
	now := time.Now().UnixMilli()

	client := r.data.GetMasterEntClient()
	builder := client.User.Create()
	builder.SetID(id)
	builder.SetUsername(body.Username)
	builder.SetCreatedAt(now)
	builder.SetUpdatedAt(now)

	if body.Email != "" {
		builder.SetEmail(body.Email)
	}
	if body.Phone != "" {
		builder.SetPhone(body.Phone)
	}
	if body.Extras != nil {
		builder.SetExtras(*body.Extras)
	}

	user, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Index in Meilisearch
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "users", Document: user}); err != nil {
		logger.Errorf(ctx, "userRepo.Create error creating Meilisearch index: %v", err)
	}

	// Cache the user
	go r.cacheUser(context.Background(), user)

	return user, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*ent.User, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.userCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	user, err := client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheUser(context.Background(), user)

	return user, nil
}

// Find retrieves a user by various filters
func (r *userRepository) Find(ctx context.Context, filter *structs.FindUser) (*ent.User, error) {
	// Try to find user ID from cache mappings first
	var userID string
	var err error

	if filter.Username != "" {
		userID, err = r.getUserIDByUsername(ctx, filter.Username)
		if err == nil && userID != "" {
			return r.GetByID(ctx, userID)
		}
	}

	if filter.Email != "" {
		userID, err = r.getUserIDByEmail(ctx, filter.Email)
		if err == nil && userID != "" {
			return r.GetByID(ctx, userID)
		}
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	builder := client.User.Query()

	if filter.ID != "" {
		builder = builder.Where(userEnt.IDEQ(filter.ID))
	}
	if filter.Username != "" {
		builder = builder.Where(userEnt.Or(
			userEnt.IDEQ(filter.Username),
			userEnt.UsernameEQ(filter.Username),
			userEnt.EmailEQ(filter.Username),
			userEnt.PhoneEQ(filter.Username),
		))
	}

	user, err := builder.Only(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheUser(context.Background(), user)

	return user, nil
}

// FindUser is an alias for Find
func (r *userRepository) FindUser(ctx context.Context, filter *structs.FindUser) (*ent.User, error) {
	return r.Find(ctx, filter)
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.User, error) {
	// Get original user for cache invalidation
	originalUser, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	client := r.data.GetMasterEntClient()
	builder := client.User.UpdateOneID(id)

	if email, ok := updates["email"].(string); ok && email != "" {
		builder = builder.SetEmail(email)
	}
	if phone, ok := updates["phone"].(string); ok && phone != "" {
		builder = builder.SetPhone(phone)
	}
	if status, ok := updates["status"].(int); ok {
		builder = builder.SetStatus(status)
	}
	if isCertified, ok := updates["is_certified"].(bool); ok {
		builder = builder.SetIsCertified(isCertified)
	}
	if isAdmin, ok := updates["is_admin"].(bool); ok {
		builder = builder.SetIsAdmin(isAdmin)
	}
	if extras, ok := updates["extras"].(types.JSON); ok {
		builder = builder.SetExtras(extras)
	}

	builder = builder.SetUpdatedAt(time.Now().UnixMilli())

	user, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Update Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "users", Document: user, DocumentID: user.ID}); err != nil {
		logger.Errorf(ctx, "userRepo.Update error updating Meilisearch index: %v", err)
	}

	// Invalidate old cache and cache new data
	go func() {
		r.invalidateUserCache(context.Background(), originalUser)
		r.cacheUser(context.Background(), user)
	}()

	return user, nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	// Get user first for cache invalidation
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	client := r.data.GetMasterEntClient()
	err = client.User.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return err
	}

	// Delete from Meilisearch
	if err = r.data.DeleteDocument(ctx, "users", id); err != nil {
		logger.Errorf(ctx, "userRepo.Delete error deleting Meilisearch index: %v", err)
	}

	// Invalidate cache
	go r.invalidateUserCache(context.Background(), user)

	return nil
}

// List lists users
func (r *userRepository) List(ctx context.Context, params *structs.ListUserParams) ([]*ent.User, error) {
	client := r.data.GetSlaveEntClient()
	builder := client.User.Query()

	// Apply filters
	if params.SearchQuery != "" {
		builder = builder.Where(
			userEnt.Or(
				userEnt.UsernameContains(params.SearchQuery),
				userEnt.EmailContains(params.SearchQuery),
			),
		)
	}
	if params.Status != 0 {
		builder = builder.Where(userEnt.StatusEQ(params.Status))
	}

	// Apply cursor pagination
	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if params.Direction == "backward" {
			builder = builder.Where(
				userEnt.Or(
					userEnt.CreatedAtGT(timestamp),
					userEnt.And(
						userEnt.CreatedAtEQ(timestamp),
						userEnt.IDGT(id),
					),
				),
			)
		} else {
			builder = builder.Where(
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

	// Apply ordering
	if params.Direction == "backward" {
		builder = builder.Order(ent.Asc(userEnt.FieldCreatedAt), ent.Asc(userEnt.FieldID))
	} else {
		builder = builder.Order(ent.Desc(userEnt.FieldCreatedAt), ent.Desc(userEnt.FieldID))
	}

	// Apply limit
	if params.Limit > 0 {
		builder = builder.Limit(params.Limit)
	}

	users, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	// Cache users in background for frequently accessed ones
	go func() {
		for _, user := range users {
			r.cacheUser(context.Background(), user)
		}
	}()

	return users, nil
}

// UpdatePassword updates user password
func (r *userRepository) UpdatePassword(ctx context.Context, body *structs.UserPassword) error {
	hashedPassword, err := crypto.HashPassword(ctx, body.NewPassword)
	if err != nil {
		return err
	}

	// Find user by username to get ID
	user, err := r.Find(ctx, &structs.FindUser{Username: body.User})
	if err != nil {
		return err
	}

	return r.UpdatePasswordByID(ctx, user.ID, hashedPassword)
}

// UpdatePasswordByID updates user password by ID
func (r *userRepository) UpdatePasswordByID(ctx context.Context, userID, hashedPassword string) error {
	client := r.data.GetMasterEntClient()
	_, err := client.User.UpdateOneID(userID).
		SetPassword(hashedPassword).
		SetUpdatedAt(time.Now().UnixMilli()).
		Save(ctx)

	if err != nil {
		return err
	}

	// Invalidate user cache (password change should invalidate cache)
	go func() {
		cacheKey := fmt.Sprintf("id:%s", userID)
		if err := r.userCache.Delete(context.Background(), cacheKey); err != nil {
			logger.Debugf(context.Background(), "Failed to invalidate user cache %s: %v", userID, err)
		}
	}()

	return nil
}

// CountX counts users
func (r *userRepository) CountX(ctx context.Context, params *structs.ListUserParams) int {
	client := r.data.GetSlaveEntClient()
	builder := client.User.Query()

	if params.SearchQuery != "" {
		builder = builder.Where(
			userEnt.Or(
				userEnt.UsernameContains(params.SearchQuery),
				userEnt.EmailContains(params.SearchQuery),
			),
		)
	}
	if params.Status != 0 {
		builder = builder.Where(userEnt.StatusEQ(params.Status))
	}

	return builder.CountX(ctx)
}

// cacheUser caches a user
func (r *userRepository) cacheUser(ctx context.Context, user *ent.User) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", user.ID)
	if err := r.userCache.Set(ctx, idKey, user, r.userTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user by ID %s: %v", user.ID, err)
	}

	// Cache username to ID mapping
	usernameKey := fmt.Sprintf("username:%s", user.Username)
	if err := r.usernameMappingCache.Set(ctx, usernameKey, &user.ID, r.userTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache username mapping %s: %v", user.Username, err)
	}

	// Cache email to ID mapping if email exists
	if user.Email != "" {
		emailKey := fmt.Sprintf("email:%s", user.Email)
		if err := r.emailMappingCache.Set(ctx, emailKey, &user.ID, r.userTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache email mapping %s: %v", user.Email, err)
		}
	}
}

// invalidateUserCache invalidates user cache
func (r *userRepository) invalidateUserCache(ctx context.Context, user *ent.User) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", user.ID)
	if err := r.userCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user ID cache %s: %v", user.ID, err)
	}

	// Invalidate username mapping
	usernameKey := fmt.Sprintf("username:%s", user.Username)
	if err := r.usernameMappingCache.Delete(ctx, usernameKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate username mapping cache %s: %v", user.Username, err)
	}

	// Invalidate email mapping if email exists
	if user.Email != "" {
		emailKey := fmt.Sprintf("email:%s", user.Email)
		if err := r.emailMappingCache.Delete(ctx, emailKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate email mapping cache %s: %v", user.Email, err)
		}
	}
}

// getUserIDByUsername gets user ID by username
func (r *userRepository) getUserIDByUsername(ctx context.Context, username string) (string, error) {
	cacheKey := fmt.Sprintf("username:%s", username)
	userID, err := r.usernameMappingCache.Get(ctx, cacheKey)
	if err != nil || userID == nil {
		return "", err
	}
	return *userID, nil
}

// getUserIDByEmail gets user ID by email
func (r *userRepository) getUserIDByEmail(ctx context.Context, email string) (string, error) {
	cacheKey := fmt.Sprintf("email:%s", email)
	userID, err := r.emailMappingCache.Get(ctx, cacheKey)
	if err != nil || userID == nil {
		return "", err
	}
	return *userID, nil
}
