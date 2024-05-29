package data

import (
	"context"
	"encoding/json"
	"stocms/internal/data/ent"
	"stocms/pkg/log"
	"stocms/pkg/validator"
)

var sampleCacheKey = getCacheKey("sample")

func (r *sampleRepo) getCache(ctx context.Context, field string) (*ent.Sample, error) {
	result, err := r.rc.HGet(ctx, sampleCacheKey, field).Result()
	if validator.IsNotNil(err) {
		return nil, err
	}
	row := &ent.Sample{}
	err = json.Unmarshal([]byte(result), row)
	if validator.IsNotNil(err) {
		return nil, err
	}
	return row, nil
}

func (r *sampleRepo) setCache(ctx context.Context, data *ent.Sample, key string) {
	bytes, err := json.Marshal(data)
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set token cache:json.Marshal(%v) error(%v)", data, err)
	}
	err = r.rc.HSet(ctx, sampleCacheKey, key, string(bytes)).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to set token cache:redis.HSet(%v) error(%v)", data, err)
	}
}

func (r *sampleRepo) deleteCache(ctx context.Context, field string) {
	err := r.rc.HDel(ctx, sampleCacheKey, field).Err()
	if validator.IsNotNil(err) {
		log.Errorf(context.Background(), "failed to delete token cache:redis.HDel(%v) field(%v) error(%v)", sampleCacheKey, field, err)
	}
}

func (r *sampleRepo) resetCache(ctx context.Context, data *ent.Sample, key string) {
	r.setCache(ctx, data, key)
}
