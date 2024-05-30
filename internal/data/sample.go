package data

import (
	"context"
	"stocms/internal/data/ent"
	"stocms/internal/data/ent/sample"
	"stocms/internal/data/structs"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// ISample represents the sample repository interface.
type ISample interface {
	Hello(ctx context.Context, body structs.Sample) (*ent.Sample, error)
}

// sampleRepo implements the ISample interface.
type sampleRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *Cache[ent.Sample]
}

// NewSample creates a new sample repository.
func NewSample(d *Data) ISample {
	return &sampleRepo{
		ec: d.ec,
		rc: d.rc,
		c:  NewCache[ent.Sample](d.rc, cacheKey("sample")),
	}
}

func (r *sampleRepo) Hello(ctx context.Context, p structs.Sample) (*ent.Sample, error) {
	// try to fetch from cache.
	cf := p.Name
	row, err := r.c.get(ctx, cf)
	if validator.IsNotNil(err) {
		// fetch from db when cache is empty.
		// use internal get method.
		row, err = r.getSample(ctx, &structs.FindSample{
			ID:   p.ID,
			Name: p.Name,
		})
		if validator.IsNotNil(err) {
			return nil, err
		}
		// set the cache.
		r.c.set(ctx, row, cf)
	}
	return row, err
}

// CountX - sample count.
func (r *sampleRepo) CountX(ctx context.Context, p *structs.ListSamples) int {
	// create list builder
	builder, err := r.listBuilder(ctx, p)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder - create list builder.
// internal method.
func (r *sampleRepo) listBuilder(ctx context.Context, p *structs.ListSamples) (*ent.SampleQuery, error) {
	// verify query params.
	var nextSample *ent.Sample
	if validator.IsNotEmpty(p.Cursor) {
		// query the sample.
		// use internal get method.
		row, err := r.getSample(ctx, &structs.FindSample{
			ID: p.Cursor,
		})
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		nextSample = row
	}

	// create builder.
	builder := r.ec.Sample.Query()

	// lt the cursor create time
	if nextSample != nil {
		builder.Where(sample.CreatedAtLT(nextSample.CreatedAt))
	}

	// set where conditions.
	// if validator.IsNotEmpty(p.User) {
	// 	builder.Where(sample.CreatedByEQ(p.User))
	// }
	// if validator.IsTrue(p.Private) || validator.IsFalse(p.Private) {
	// 	builder.Where(sample.PrivateEQ(p.Private))
	// }
	// if validator.IsTrue(p.Temp) || validator.IsFalse(p.Temp) {
	// 	builder.Where(sample.TempEQ(p.Temp))
	// }

	return builder, nil
}

// IsExist - check if sample exists
func (r *sampleRepo) IsExist(ctx context.Context, p *structs.FindSample) bool {
	// query the sample.
	// use internal get method.
	_, err := r.getSample(ctx, p)
	return validator.IsNotNil(err)
}

// getSample - get sample.
// internal method.
func (r *sampleRepo) getSample(ctx context.Context, p *structs.FindSample) (*ent.Sample, error) {
	// create builder.
	builder := r.ec.Sample.Query()

	// set where conditions.
	if validator.IsNotEmpty(p.ID) {
		builder.Where(sample.IDEQ(p.ID))
	}
	if validator.IsNotEmpty(p.Name) {
		builder.Where(sample.NameEQ(p.Name))
	}

	// execute the builder.
	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}
	return row, nil
}
