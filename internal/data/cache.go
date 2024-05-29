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
	// get data from c using the specified field and return a pointer to type T and a possible error.
	get(context.Context, string) (*T, error)
	// set saves the specified data into c using the specified string as key.
	set(context.Context, *T, string)
	// delete data from c using the specified key.
	delete(context.Context, string)
	// reset data in c using the specified pointer to type T as new value.
	reset(context.Context, *T, string)
}

// Cache implements the ICache interface
type Cache[T any] struct {
	rc  redis.Cmdable
	key string
}

// cacheKey defines the c key for the user service
// @param key - format: prefix:%s, %s = table name or custom
func cacheKey(key string) string {
	return fmt.Sprintf("%s_%s:%s", "sc", "sample", key)
}

// NewCache creates a new Cache instance
func NewCache[T any](rc redis.Cmdable, key string) *Cache[T] {
	return &Cache[T]{rc: rc, key: key}
}

// get retrieves data from c
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

// set saves data into c
func (c *Cache[T]) set(ctx context.Context, data *T, field string) {
	bytes, err := json.Marshal(data)
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set c: json.Marshal(%v) error(%v)", data, err)
		return
	}
	err = c.rc.HSet(ctx, c.key, field, string(bytes)).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set c: redis.HSet(%v) error(%v)", data, err)
	}
}

// delete removes data from c
func (c *Cache[T]) delete(ctx context.Context, field string) {
	err := c.rc.HDel(ctx, c.key, field).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to delete c: redis.HDel(%v) field(%v) error(%v)", c.key, field, err)
	}
}

// reset resets data in c
func (c *Cache[T]) reset(ctx context.Context, data *T, field string) {
	c.set(ctx, data, field)
}
