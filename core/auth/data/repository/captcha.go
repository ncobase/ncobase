package repository

import (
	"github.com/redis/go-redis/v9"
	"context"
	"fmt"
	"ncobase/core/auth/data"
	"time"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// CaptchaRepositoryInterface represents the captcha repository interface.
type CaptchaRepositoryInterface interface {
	Set(ctx context.Context, id string, m *types.JSON) error
	Get(ctx context.Context, id string) (*types.JSON, error)
	Delete(ctx context.Context, id string) error
	Verify(ctx context.Context, id string, answer string) (bool, error)
	IsExpired(ctx context.Context, id string) (bool, error)
	GetAndDelete(ctx context.Context, id string) (*types.JSON, error)
}

// captchaRepository implements the CaptchaRepositoryInterface.
type captchaRepository struct {
	captchaCache cache.ICache[types.JSON]
	attemptCache cache.ICache[int] // Track verification attempts
	captchaTTL   time.Duration
	maxAttempts  int
}

// NewCaptchaRepository creates a new captcha repository.
func NewCaptchaRepository(d *data.Data) CaptchaRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &captchaRepository{
		captchaCache: cache.NewCache[types.JSON](redisClient, "ncse_auth:captchas"),
		attemptCache: cache.NewCache[int](redisClient, "ncse_auth:captcha_attempts"),
		captchaTTL:   5 * time.Minute, // 5 minutes cache TTL
		maxAttempts:  3,               // Maximum verification attempts
	}
}

// Set sets the captcha in the cache.
func (r *captchaRepository) Set(ctx context.Context, id string, m *types.JSON) error {
	cacheKey := fmt.Sprintf("captcha:%s", id)
	if err := r.captchaCache.Set(ctx, cacheKey, m, r.captchaTTL); err != nil {
		logger.Errorf(ctx, "captchaRepo.Set error: %v", err)
		return err
	}

	// Initialize attempt counter
	attemptKey := fmt.Sprintf("attempts:%s", id)
	if err := r.attemptCache.Set(ctx, attemptKey, &[]int{0}[0], r.captchaTTL); err != nil {
		logger.Debugf(ctx, "Failed to initialize attempt counter for captcha %s: %v", id, err)
	}

	return nil
}

// Get gets the captcha from the cache.
func (r *captchaRepository) Get(ctx context.Context, id string) (*types.JSON, error) {
	cacheKey := fmt.Sprintf("captcha:%s", id)
	captcha, err := r.captchaCache.Get(ctx, cacheKey)
	if err != nil {
		logger.Debugf(ctx, "captchaRepo.Get error for ID %s: %v", id, err)
		return nil, err
	}

	return captcha, nil
}

// Delete deletes the captcha from the cache.
func (r *captchaRepository) Delete(ctx context.Context, id string) error {
	cacheKey := fmt.Sprintf("captcha:%s", id)
	attemptKey := fmt.Sprintf("attempts:%s", id)

	// Delete captcha
	if err := r.captchaCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to delete captcha %s: %v", id, err)
	}

	// Delete attempt counter
	if err := r.attemptCache.Delete(ctx, attemptKey); err != nil {
		logger.Debugf(ctx, "Failed to delete attempt counter for captcha %s: %v", id, err)
	}

	return nil
}

// Verify verifies the captcha answer and tracks attempts.
func (r *captchaRepository) Verify(ctx context.Context, id string, answer string) (bool, error) {
	// Check if captcha exists
	captcha, err := r.Get(ctx, id)
	if err != nil {
		return false, err
	}

	// Check attempt count
	attemptKey := fmt.Sprintf("attempts:%s", id)
	attempts, err := r.attemptCache.Get(ctx, attemptKey)
	if err != nil {
		// If no attempt record, initialize to 0
		attempts = &[]int{0}[0]
	}

	// Check if max attempts exceeded
	if *attempts >= r.maxAttempts {
		logger.Warnf(ctx, "Max verification attempts exceeded for captcha %s", id)
		r.Delete(ctx, id) // Clean up expired captcha
		return false, fmt.Errorf("maximum verification attempts exceeded")
	}

	// Increment attempt counter
	newAttempts := *attempts + 1
	if err := r.attemptCache.Set(ctx, attemptKey, &newAttempts, r.captchaTTL); err != nil {
		logger.Debugf(ctx, "Failed to update attempt counter for captcha %s: %v", id, err)
	}

	// Verify answer (assuming the answer is stored in captcha data)
	if captchaData, ok := (*captcha)["answer"]; ok {
		if expectedAnswer, ok := captchaData.(string); ok {
			isValid := expectedAnswer == answer

			// If verification successful or max attempts reached, delete captcha
			if isValid || newAttempts >= r.maxAttempts {
				go r.Delete(context.Background(), id)
			}

			return isValid, nil
		}
	}

	return false, fmt.Errorf("invalid captcha data format")
}

// IsExpired checks if a captcha has expired.
func (r *captchaRepository) IsExpired(ctx context.Context, id string) (bool, error) {
	_, err := r.Get(ctx, id)
	if err != nil {
		// If we can't get it, consider it expired
		return true, nil
	}
	return false, nil
}

// GetAndDelete retrieves and immediately deletes the captcha (one-time use).
func (r *captchaRepository) GetAndDelete(ctx context.Context, id string) (*types.JSON, error) {
	captcha, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Delete the captcha after retrieving it
	go r.Delete(context.Background(), id)

	return captcha, nil
}
