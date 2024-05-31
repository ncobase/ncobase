package cache

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
	// Get data from cache using the specified field and return a pointer to type T and a possible error.
	Get(context.Context, string) (*T, error)
	// Set saves the specified data into cache using the specified string as key.
	Set(context.Context, *T, string)
	// Delete data from cache using the specified key.
	Delete(context.Context, string)
	// Reset data in cache using the specified pointer to type T as new value.
	Reset(context.Context, *T, string)
}

// Cache implements the ICache interface
type Cache[T any] struct {
	rc  *redis.Client
	key string
}

// Key defines the cache key for the user service
func Key(key string) string {
	return fmt.Sprintf("%s_%s", "sto", key)
}

// NewCache creates a new Cache instance
func NewCache[T any](rc *redis.Client, key string) *Cache[T] {
	return &Cache[T]{rc: rc, key: key}
}

// Get retrieves data from cache
func (c *Cache[T]) Get(ctx context.Context, field string) (*T, error) {
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

// Set saves data into cache
func (c *Cache[T]) Set(ctx context.Context, field string, data *T) {
	bytes, err := json.Marshal(data)
	if validator.IsNotNil(err) {
		log.Errorf(nil, "failed to Set cache: json.Marshal(%v) error(%v)", data, err)
		return
	}
	err = c.rc.HSet(ctx, c.key, field, string(bytes)).Err()
	if validator.IsNotNil(err) {
		log.Errorf(nil, "failed to Set cache: redis.HSet(%v) error(%v)", data, err)
	}
}

// Delete removes data from cache
func (c *Cache[T]) Delete(ctx context.Context, field string) {
	err := c.rc.HDel(ctx, c.key, field).Err()
	if validator.IsNotNil(err) {
		log.Errorf(nil, "failed to Delete cache: redis.HDel(%v) field(%v) error(%v)", c.key, field, err)
	}
}

// Reset resets data in cache
func (c *Cache[T]) Reset(ctx context.Context, field string, data *T) {
	c.Set(ctx, field, data)
}
