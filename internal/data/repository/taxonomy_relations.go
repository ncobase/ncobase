package repo

import (
	"context"
	"errors"
	"fmt"
	"stocms/internal/data"
	"stocms/internal/data/ent"
	taxonomyRelationsEnt "stocms/internal/data/ent/taxonomyrelations"
	"stocms/internal/data/structs"
	"stocms/pkg/cache"
	"stocms/pkg/log"
	"stocms/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// TaxonomyRelations represents the taxonomy relations repository interface.
type TaxonomyRelations interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyRelationsBody) (*ent.TaxonomyRelations, error)
	GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelations, error)
	Update(ctx context.Context, body *structs.UpdateTaxonomyRelationsBody) (*ent.TaxonomyRelations, error)
	List(ctx context.Context, params *structs.ListTaxonomyRelationsParams) ([]*ent.TaxonomyRelations, error)
	Delete(ctx context.Context, object string) error
	BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationsBody) ([]*ent.TaxonomyRelations, error)
	FindRelations(ctx context.Context, params *structs.FindTaxonomyRelationsParams) ([]*ent.TaxonomyRelations, error)
}

// taxonomyRelationsRepo implements the TaxonomyRelations interface.
type taxonomyRelationsRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.TaxonomyRelations]
}

// NewTaxonomyRelations creates a new taxonomy relations repository.
func NewTaxonomyRelations(d *data.Data) TaxonomyRelations {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &taxonomyRelationsRepo{ec, rc, cache.NewCache[ent.TaxonomyRelations](rc, cache.Key("sc_taxonomy_relations"), true)}
}

// Create creates a new taxonomy relation.
func (r *taxonomyRelationsRepo) Create(ctx context.Context, body *structs.CreateTaxonomyRelationsBody) (*ent.TaxonomyRelations, error) {
	query := r.ec.TaxonomyRelations.
		Create().
		SetID(body.ObjectID).
		SetTaxonomyID(body.TaxonomyID).
		SetType(body.Type).
		SetOrder(body.Order).
		SetCreatedBy(body.CreatedBy).
		SetCreatedAt(body.CreatedAt)

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByObject gets a taxonomy relation by Object.
func (r *taxonomyRelationsRepo) GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelations, error) {
	cacheKey := fmt.Sprintf("%s", object)

	// check cache first
	if cached, err := r.c.Get(ctx, cacheKey); err == nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTaxonomyRelations(ctx, &structs.FindTaxonomyRelations{Object: object})

	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.GetByObject error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.GetByObject cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a taxonomy relation.
func (r *taxonomyRelationsRepo) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationsBody) (*ent.TaxonomyRelations, error) {
	taxonomyRelations, err := r.GetByObject(ctx, body.ObjectID)
	if err != nil {
		return nil, err
	}

	query := taxonomyRelations.
		Update().
		SetTaxonomyID(body.TaxonomyID).
		SetType(body.Type).
		SetOrder(body.Order)

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", body.ObjectID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of taxonomy relations.
func (r *taxonomyRelationsRepo) List(ctx context.Context, p *structs.ListTaxonomyRelationsParams) ([]*ent.TaxonomyRelations, error) {
	var nextTaxonomyRelations *ent.TaxonomyRelations
	if p.Cursor != "" {
		taxonomyRelations, err := r.ec.TaxonomyRelations.
			Query().
			Where(
				taxonomyRelationsEnt.IDEQ(p.Cursor),
			).
			First(ctx)
		if err != nil || taxonomyRelations == nil {
			return nil, errors.New("invalid cursor")
		}
		nextTaxonomyRelations = taxonomyRelations
	}

	query := r.ec.TaxonomyRelations.
		Query().
		Limit(int(p.Limit))

	// lt the cursor create time
	if nextTaxonomyRelations != nil {
		query.Where(taxonomyRelationsEnt.CreatedAtLT(nextTaxonomyRelations.CreatedAt))
	}

	// sort
	query.Order(ent.Desc(taxonomyRelationsEnt.FieldCreatedAt))

	rows, err := query.All(ctx)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a taxonomy relation.
func (r *taxonomyRelationsRepo) Delete(ctx context.Context, object string) error {
	_, err := r.ec.TaxonomyRelations.
		Delete().
		Where(taxonomyRelationsEnt.IDEQ(object)).
		Exec(ctx)

	if err == nil {
		// remove from cache
		cacheKey := fmt.Sprintf("%s", object)
		err := r.c.Delete(ctx, cacheKey)
		if err != nil {
			log.Errorf(nil, "taxonomyRelationsRepo.Delete cache error: %v\n", err)
		}
	}

	return err
}

// BatchCreate creates multiple taxonomy relations.
func (r *taxonomyRelationsRepo) BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationsBody) ([]*ent.TaxonomyRelations, error) {
	bulk := make([]*ent.TaxonomyRelationsCreate, len(bodies))
	for i, body := range bodies {
		bulk[i] = r.ec.TaxonomyRelations.
			Create().
			SetID(body.ObjectID).
			SetTaxonomyID(body.TaxonomyID).
			SetType(body.Type).
			SetOrder(body.Order).
			SetCreatedBy(body.CreatedBy).
			SetCreatedAt(body.CreatedAt)
	}
	rows, err := r.ec.TaxonomyRelations.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.BatchCreate error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// FindRelations finds taxonomy relations by various criteria.
func (r *taxonomyRelationsRepo) FindRelations(ctx context.Context, p *structs.FindTaxonomyRelationsParams) ([]*ent.TaxonomyRelations, error) {

	// create builder.
	builder := r.ec.TaxonomyRelations.Query()

	if validator.IsNotEmpty(p.Object) {
		builder = builder.Where(taxonomyRelationsEnt.IDEQ(p.Object))
	}
	if validator.IsNotEmpty(p.Taxonomy) {
		builder = builder.Where(taxonomyRelationsEnt.TaxonomyIDEQ(p.Taxonomy))
	}
	if validator.IsNotEmpty(p.Type) {
		builder = builder.Where(taxonomyRelationsEnt.TypeEQ(p.Type))
	}

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(nil, "taxonomyRelationsRepo.FindRelations error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// FindTaxonomyRelations gets a single taxonomy relation by criteria.
func (r *taxonomyRelationsRepo) FindTaxonomyRelations(ctx context.Context, p *structs.FindTaxonomyRelations) (*ent.TaxonomyRelations, error) {

	// create builder.
	builder := r.ec.TaxonomyRelations.Query()

	if validator.IsNotEmpty(p.Object) {
		builder = builder.Where(taxonomyRelationsEnt.IDEQ(p.Object))
	}
	if validator.IsNotEmpty(p.Taxonomy) {
		builder = builder.Where(taxonomyRelationsEnt.TaxonomyIDEQ(p.Taxonomy))
	}
	if validator.IsNotEmpty(p.Type) {
		builder = builder.Where(taxonomyRelationsEnt.TypeEQ(p.Type))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}
