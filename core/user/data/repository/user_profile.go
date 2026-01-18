package repository

import (
	"context"
	"fmt"
	"ncobase/core/user/data"
	"ncobase/core/user/data/ent"
	"ncobase/core/user/structs"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// UserProfileRepositoryInterface defines user profile repository operations
type UserProfileRepositoryInterface interface {
	Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error)
	Update(ctx context.Context, userID string, updates types.JSON) (*ent.UserProfile, error)
	Get(ctx context.Context, userID string) (*ent.UserProfile, error)
	Delete(ctx context.Context, userID string) error
}

// userProfileRepository implements UserProfileRepositoryInterface
type userProfileRepository struct {
	data         *data.Data
	profileCache cache.ICache[ent.UserProfile]
	profileTTL   time.Duration
}

// NewUserProfileRepository creates a new user profile repository
func NewUserProfileRepository(d *data.Data) UserProfileRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &userProfileRepository{
		data:         d,
		profileCache: cache.NewCache[ent.UserProfile](redisClient, "ncse_user_profiles"),
		profileTTL:   time.Hour * 1, // 1 hour cache TTL
	}
}

// Create creates a new user profile
func (r *userProfileRepository) Create(ctx context.Context, body *structs.UserProfileBody) (*ent.UserProfile, error) {
	client := r.data.GetMasterEntClient()
	builder := client.UserProfile.Create()

	builder.SetID(body.UserID)
	if body.DisplayName != "" {
		builder.SetDisplayName(body.DisplayName)
	}
	if body.FirstName != "" {
		builder.SetFirstName(body.FirstName)
	}
	if body.LastName != "" {
		builder.SetLastName(body.LastName)
	}
	if body.Title != "" {
		builder.SetTitle(body.Title)
	}
	if body.ShortBio != "" {
		builder.SetShortBio(body.ShortBio)
	}
	if body.About != nil {
		builder.SetAbout(*body.About)
	}
	if body.Thumbnail != nil {
		builder.SetThumbnail(*body.Thumbnail)
	}
	if body.Links != nil {
		builder.SetLinks(*body.Links)
	}
	if body.Extras != nil {
		builder.SetExtras(*body.Extras)
	}

	profile, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the profile
	go r.cacheProfile(context.Background(), profile)

	return profile, nil
}

// Get retrieves a user profile
func (r *userProfileRepository) Get(ctx context.Context, userID string) (*ent.UserProfile, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("profile:%s", userID)
	if cached, err := r.profileCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database
	client := r.data.GetSlaveEntClient()
	profile, err := client.UserProfile.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache for future use
	go r.cacheProfile(context.Background(), profile)

	return profile, nil
}

// Update updates a user profile
func (r *userProfileRepository) Update(ctx context.Context, userID string, updates types.JSON) (*ent.UserProfile, error) {
	client := r.data.GetMasterEntClient()
	builder := client.UserProfile.UpdateOneID(userID)

	if displayName, ok := updates["display_name"].(string); ok && displayName != "" {
		builder = builder.SetDisplayName(displayName)
	}
	if firstName, ok := updates["first_name"].(string); ok && firstName != "" {
		builder = builder.SetFirstName(firstName)
	}
	if lastName, ok := updates["last_name"].(string); ok && lastName != "" {
		builder = builder.SetLastName(lastName)
	}
	if title, ok := updates["title"].(string); ok && title != "" {
		builder = builder.SetTitle(title)
	}
	if shortBio, ok := updates["short_bio"].(string); ok && shortBio != "" {
		builder = builder.SetShortBio(shortBio)
	}
	if about, ok := updates["about"].(string); ok {
		builder = builder.SetAbout(about)
	}
	if thumbnail, ok := updates["thumbnail"].(string); ok {
		builder = builder.SetThumbnail(thumbnail)
	}
	if links, ok := updates["links"].([]types.JSON); ok {
		builder = builder.SetLinks(links)
	}
	if extras, ok := updates["extras"].(types.JSON); ok {
		builder = builder.SetExtras(extras)
	}

	profile, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateProfileCache(context.Background(), userID)
		r.cacheProfile(context.Background(), profile)
	}()

	return profile, nil
}

// Delete deletes a user profile
func (r *userProfileRepository) Delete(ctx context.Context, userID string) error {
	client := r.data.GetMasterEntClient()
	err := client.UserProfile.DeleteOneID(userID).Exec(ctx)
	if err != nil {
		return err
	}

	// Invalidate cache
	go r.invalidateProfileCache(context.Background(), userID)

	return nil
}

// cacheProfile caches a user profile
func (r *userProfileRepository) cacheProfile(ctx context.Context, profile *ent.UserProfile) {
	cacheKey := fmt.Sprintf("profile:%s", profile.ID)
	if err := r.profileCache.Set(ctx, cacheKey, profile, r.profileTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache user profile %s: %v", profile.ID, err)
	}
}

// invalidateProfileCache invalidates user profile cache
func (r *userProfileRepository) invalidateProfileCache(ctx context.Context, userID string) {
	cacheKey := fmt.Sprintf("profile:%s", userID)
	if err := r.profileCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate user profile cache %s: %v", userID, err)
	}
}
