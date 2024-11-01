package repository

import (
	"context"
	"ncobase/core/auth/data"
	"time"

	"ncobase/common/cache"
	"ncobase/common/types"

	"github.com/redis/go-redis/v9"
)

// CaptchaRepositoryInterface represents the captcha repository interface.
type CaptchaRepositoryInterface interface {
	Set(ctx context.Context, id string, m *types.JSON) error
	Get(ctx context.Context, id string) (*types.JSON, error)
	Delete(ctx context.Context, id string) error
}

// captchaRepository implements the CaptchaRepositoryInterface.
type captchaRepository struct {
	rc *redis.Client
	c  *cache.Cache[types.JSON]
}

// NewCaptchaRepository creates a new captcha repository.
func NewCaptchaRepository(d *data.Data) CaptchaRepositoryInterface {
	rc := d.GetRedis()
	return &captchaRepository{rc, cache.NewCache[types.JSON](rc, "ncse_captcha", false)}
}

// Set sets the captcha in the cache.
func (r *captchaRepository) Set(ctx context.Context, id string, m *types.JSON) error {

	return r.c.Set(ctx, id, m, 5*time.Minute)
}

// Get gets the captcha from the cache.
func (r *captchaRepository) Get(ctx context.Context, id string) (*types.JSON, error) {
	return r.c.Get(ctx, id)
}

// Delete deletes the captcha from the cache.
func (r *captchaRepository) Delete(ctx context.Context, id string) error {
	return r.c.Delete(ctx, id)
}

// id is the cache key for the captcha.
var id = "sc_captcha"
