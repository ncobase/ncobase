package repo

import (
	"context"
	"stocms/internal/data"
	"stocms/pkg/cache"
	"stocms/pkg/types"
	"time"

	"github.com/redis/go-redis/v9"
)

// Captcha represents the captcha repository interface.
type Captcha interface {
	Set(ctx context.Context, id string, m *types.JSON) error
	Get(ctx context.Context, id string) (*types.JSON, error)
	Delete(ctx context.Context, id string) error
}

// captchaRepo implements the Captcha interface.
type captchaRepo struct {
	rc *redis.Client
	c  *cache.Cache[types.JSON]
}

// NewCaptcha creates a new captcha repository.
func NewCaptcha(d *data.Data) Captcha {
	rc := d.GetRedis()
	return &captchaRepo{rc, cache.NewCache[types.JSON](rc, cache.Key("sc_captcha"), false)}
}

// Set sets the captcha in the cache.
func (r *captchaRepo) Set(ctx context.Context, id string, m *types.JSON) error {

	return r.c.Set(ctx, id, m, 5*time.Minute)
}

// Get gets the captcha from the cache.
func (r *captchaRepo) Get(ctx context.Context, id string) (*types.JSON, error) {
	return r.c.Get(ctx, id)
}

// Delete deletes the captcha from the cache.
func (r *captchaRepo) Delete(ctx context.Context, id string) error {
	return r.c.Delete(ctx, id)
}

// id is the cache key for the captcha.
var id = "sc_captcha"
