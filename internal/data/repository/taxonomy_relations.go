package repo

import (
	"context"
	"errors"
	"fmt"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	taxonomyRelationEnt "ncobase/internal/data/ent/taxonomyrelation"
	"ncobase/internal/data/structs"

	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// TaxonomyRelation represents the taxonomy relations repository interface.
type TaxonomyRelation interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*ent.TaxonomyRelation, error)
	GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelation, error)
	Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*ent.TaxonomyRelation, error)
	List(ctx context.Context, params *structs.ListTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error)
	Delete(ctx context.Context, object string) error
	BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationBody) ([]*ent.TaxonomyRelation, error)
	FindRelations(ctx context.Context, params *structs.FindTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error)
}

// taxonomyRelationsRepo implements the TaxonomyRelation interface.
type taxonomyRelationsRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.TaxonomyRelation]
}

// NewTaxonomyRelation creates a new taxonomy relations repository.
func NewTaxonomyRelation(d *data.Data) TaxonomyRelation {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &taxonomyRelationsRepo{ec, rc, cache.NewCache[ent.TaxonomyRelation](rc, "nb_taxonomy_relations")}
}

// Create creates a new taxonomy relation.
func (r *taxonomyRelationsRepo) Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*ent.TaxonomyRelation, error) {
	query := r.ec.TaxonomyRelation.
		Create().
		SetID(body.ObjectID).
		SetTaxonomyID(body.TaxonomyID).
		SetType(body.Type).
		SetNillableOrder(body.Order).
		SetNillableCreatedBy(body.CreatedBy)

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByObject gets a taxonomy relation by Object.
func (r *taxonomyRelationsRepo) GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelation, error) {
	cacheKey := fmt.Sprintf("%s", object)

	// check cache first
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTaxonomyRelation(ctx, &structs.FindTaxonomyRelation{ObjectID: object})

	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.GetByObject error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.GetByObject cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a taxonomy relation.
func (r *taxonomyRelationsRepo) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*ent.TaxonomyRelation, error) {
	taxonomyRelations, err := r.GetByObject(ctx, body.ObjectID)
	if err != nil {
		return nil, err
	}

	query := taxonomyRelations.
		Update().
		SetTaxonomyID(body.TaxonomyID).
		SetType(body.Type).
		SetNillableOrder(body.Order)

	row, err := query.Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", body.ObjectID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of taxonomy relations.
func (r *taxonomyRelationsRepo) List(ctx context.Context, params *structs.ListTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error) {
	var next *ent.TaxonomyRelation
	if params.Cursor != "" {
		taxonomyRelations, err := r.ec.TaxonomyRelation.
			Query().
			Where(
				taxonomyRelationEnt.IDEQ(params.Cursor),
			).
			First(ctx)
		if err != nil || taxonomyRelations == nil {
			return nil, errors.New("invalid cursor")
		}
		next = taxonomyRelations
	}

	query := r.ec.TaxonomyRelation.
		Query().
		Limit(params.Limit)

	// lt the cursor create time
	if next != nil {
		query.Where(taxonomyRelationEnt.CreatedAtLT(next.CreatedAt))
	}

	// sort
	query.Order(ent.Desc(taxonomyRelationEnt.FieldCreatedAt))

	rows, err := query.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a taxonomy relation.
func (r *taxonomyRelationsRepo) Delete(ctx context.Context, object string) error {
	_, err := r.ec.TaxonomyRelation.
		Delete().
		Where(taxonomyRelationEnt.IDEQ(object)).
		Exec(ctx)

	if err == nil {
		// remove from cache
		cacheKey := fmt.Sprintf("%s", object)
		err := r.c.Delete(ctx, cacheKey)
		if err != nil {
			log.Errorf(context.Background(), "taxonomyRelationsRepo.Delete cache error: %v\n", err)
		}
	}

	return err
}

// BatchCreate creates multiple taxonomy relations.
func (r *taxonomyRelationsRepo) BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationBody) ([]*ent.TaxonomyRelation, error) {
	bulk := make([]*ent.TaxonomyRelationCreate, len(bodies))
	for i, body := range bodies {
		bulk[i] = r.ec.TaxonomyRelation.
			Create().
			SetID(body.ObjectID).
			SetTaxonomyID(body.TaxonomyID).
			SetType(body.Type).
			SetNillableOrder(body.Order).
			SetNillableCreatedBy(body.CreatedBy)
	}
	rows, err := r.ec.TaxonomyRelation.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.BatchCreate error: %v\n", err)
		return nil, err
	}
	return rows, nil
}

// FindRelations finds taxonomy relations by various criteria.
func (r *taxonomyRelationsRepo) FindRelations(ctx context.Context, params *structs.FindTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error) {

	// create builder.
	builder := r.ec.TaxonomyRelation.Query()

	if validator.IsNotEmpty(params.ObjectID) {
		builder = builder.Where(taxonomyRelationEnt.IDEQ(params.ObjectID))
	}
	if validator.IsNotEmpty(params.TaxonomyID) {
		builder = builder.Where(taxonomyRelationEnt.TaxonomyIDEQ(params.TaxonomyID))
	}
	if validator.IsNotEmpty(params.Type) {
		builder = builder.Where(taxonomyRelationEnt.TypeEQ(params.Type))
	}

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "taxonomyRelationsRepo.FindRelations error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// FindTaxonomyRelation gets a single taxonomy relation by criteria.
func (r *taxonomyRelationsRepo) FindTaxonomyRelation(ctx context.Context, params *structs.FindTaxonomyRelation) (*ent.TaxonomyRelation, error) {

	// create builder.
	builder := r.ec.TaxonomyRelation.Query()

	if validator.IsNotEmpty(params.ObjectID) {
		builder = builder.Where(taxonomyRelationEnt.IDEQ(params.ObjectID))
	}
	if validator.IsNotEmpty(params.TaxonomyID) {
		builder = builder.Where(taxonomyRelationEnt.TaxonomyIDEQ(params.TaxonomyID))
	}
	if validator.IsNotEmpty(params.Type) {
		builder = builder.Where(taxonomyRelationEnt.TypeEQ(params.Type))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}
