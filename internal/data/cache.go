package data

import (
	"context"
	"encoding/json"
	"fmt"
	"stocms/pkg/log"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// ICache defines a general caching interface
type ICache[T any] interface {
	// get data from cache using the specified field and return a pointer to type T and a possible error.
	get(context.Context, string) (*T, error)
	// set saves the specified data into cache using the specified string as key.
	set(context.Context, *T, string)
	// delete data from cache using the specified key.
	delete(context.Context, string)
	// reset data in cache using the specified pointer to type T as new value.
	reset(context.Context, *T, string)
}

// Cache implements the ICache interface
type Cache[T any] struct {
	rc  redis.Cmdable
	key string
}

// cacheKey defines the cache key for the user service
// @param key - format: prefix:%s, %s = table name or custom
func cacheKey(key string) string {
	return fmt.Sprintf("%s_%s", "sto", key)
}

// NewCache creates a new Cache instance
func NewCache[T any](rc redis.Cmdable, key string) *Cache[T] {
	return &Cache[T]{rc: rc, key: key}
}

// get retrieves data from cache
func (c *Cache[T]) get(ctx context.Context, field string) (*T, error) {
	result, err := c.rc.HGet(ctx, c.key, field).Result()
	if validator.IsNotNil(err) {
		return nil, err
	}
	var row T
	err = json.Unmarshal([]byte(result), &row)
	if validator.IsNotNil(err) {
		return nil, err
	}
	return &row, nil
}

// set saves data into cache
func (c *Cache[T]) set(ctx context.Context, data *T, field string) {
	bytes, err := json.Marshal(data)
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set cache: json.Marshal(%v) error(%v)", data, err)
		return
	}
	err = c.rc.HSet(ctx, c.key, field, string(bytes)).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set cache: redis.HSet(%v) error(%v)", data, err)
	}
}

// delete removes data from cache
func (c *Cache[T]) delete(ctx context.Context, field string) {
	err := c.rc.HDel(ctx, c.key, field).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to delete cache: redis.HDel(%v) field(%v) error(%v)", c.key, field, err)
	}
}

// reset resets data in cache
func (c *Cache[T]) reset(ctx context.Context, data *T, field string) {
	c.set(ctx, data, field)
}
