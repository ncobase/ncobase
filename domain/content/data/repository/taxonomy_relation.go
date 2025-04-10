package repository

import (
	"context"
	"fmt"
	"ncobase/domain/content/data"
	"ncobase/domain/content/data/ent"
	taxonomyRelationEnt "ncobase/domain/content/data/ent/taxonomyrelation"
	"ncobase/domain/content/structs"

	"github.com/ncobase/ncore/pkg/data/cache"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/nanoid"
	"github.com/ncobase/ncore/pkg/paging"
	"github.com/ncobase/ncore/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// TaxonomyRelationsRepositoryInterface represents the taxonomy relations repository interface.
type TaxonomyRelationsRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*ent.TaxonomyRelation, error)
	GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelation, error)
	Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*ent.TaxonomyRelation, error)
	List(ctx context.Context, params *structs.ListTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error)
	CountX(ctx context.Context, params *structs.ListTaxonomyRelationParams) int
	Delete(ctx context.Context, object string) error
	BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationBody) ([]*ent.TaxonomyRelation, error)
	FindRelations(ctx context.Context, params *structs.FindTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error)
}

// taxonomyRelationsRepository implements the TaxonomyRelationsRepositoryInterface.
type taxonomyRelationsRepository struct {
	ec  *ent.Client
	ecr *ent.Client
	rc  *redis.Client
	c   *cache.Cache[ent.TaxonomyRelation]
}

// NewTaxonomyRelationsRepository creates a new taxonomy relations repository.
func NewTaxonomyRelationsRepository(d *data.Data) TaxonomyRelationsRepositoryInterface {
	ec := d.GetEntClient()
	ecr := d.GetEntClientRead()
	rc := d.GetRedis()
	return &taxonomyRelationsRepository{ec, ecr, rc, cache.NewCache[ent.TaxonomyRelation](rc, "ncse_taxonomy_relations")}
}

// Create creates a new taxonomy relation.
func (r *taxonomyRelationsRepository) Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*ent.TaxonomyRelation, error) {
	query := r.ec.TaxonomyRelation.
		Create().
		SetID(body.ObjectID).
		SetTaxonomyID(body.TaxonomyID).
		SetType(body.Type).
		SetNillableOrder(body.Order).
		SetNillableCreatedBy(body.CreatedBy)

	row, err := query.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRelationsRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByObject gets a taxonomy relation by Object.
func (r *taxonomyRelationsRepository) GetByObject(ctx context.Context, object string) (*ent.TaxonomyRelation, error) {
	cacheKey := fmt.Sprintf("%s", object)

	// check cache first
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindTaxonomyRelation(ctx, &structs.FindTaxonomyRelation{ObjectID: object})

	if err != nil {
		logger.Errorf(ctx, "taxonomyRelationsRepo.GetByObject error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRelationsRepo.GetByObject cache error: %v", err)
	}

	return row, nil
}

// Update updates a taxonomy relation.
func (r *taxonomyRelationsRepository) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*ent.TaxonomyRelation, error) {
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
		logger.Errorf(ctx, "taxonomyRelationsRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", body.ObjectID)
	err = r.c.Delete(ctx, cacheKey)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRelationsRepo.Update cache error: %v", err)
	}

	return row, nil
}

// List gets a list of taxonomy relations.
func (r *taxonomyRelationsRepository) List(ctx context.Context, params *structs.ListTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error) {
	// create builder.
	builder := r.ecr.TaxonomyRelation.
		Query().
		Limit(params.Limit)

	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}

		if !nanoid.IsPrimaryKey(id) {
			return nil, fmt.Errorf("invalid id in cursor: %s", id)
		}

		if params.Direction == "backward" {
			builder.Where(
				taxonomyRelationEnt.Or(
					taxonomyRelationEnt.CreatedAtGT(timestamp),
					taxonomyRelationEnt.And(
						taxonomyRelationEnt.CreatedAtEQ(timestamp),
						taxonomyRelationEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				taxonomyRelationEnt.Or(
					taxonomyRelationEnt.CreatedAtLT(timestamp),
					taxonomyRelationEnt.And(
						taxonomyRelationEnt.CreatedAtEQ(timestamp),
						taxonomyRelationEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(taxonomyRelationEnt.FieldCreatedAt), ent.Asc(taxonomyRelationEnt.FieldID))
	} else {
		builder.Order(ent.Desc(taxonomyRelationEnt.FieldCreatedAt), ent.Desc(taxonomyRelationEnt.FieldID))
	}

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "taxonomyRelationsRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX gets a count of taxonomy relations.
func (r *taxonomyRelationsRepository) CountX(ctx context.Context, _ *structs.ListTaxonomyRelationParams) int {
	return r.ecr.TaxonomyRelation.
		Query().
		CountX(ctx)
}

// Delete deletes a taxonomy relation.
func (r *taxonomyRelationsRepository) Delete(ctx context.Context, object string) error {
	_, err := r.ecr.TaxonomyRelation.
		Delete().
		Where(taxonomyRelationEnt.IDEQ(object)).
		Exec(ctx)

	if err == nil {
		// remove from cache
		cacheKey := fmt.Sprintf("%s", object)
		err := r.c.Delete(ctx, cacheKey)
		if err != nil {
			logger.Errorf(ctx, "taxonomyRelationsRepo.Delete cache error: %v", err)
		}
	}

	return err
}

// BatchCreate creates multiple taxonomy relations.
func (r *taxonomyRelationsRepository) BatchCreate(ctx context.Context, bodies []*structs.CreateTaxonomyRelationBody) ([]*ent.TaxonomyRelation, error) {
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
		logger.Errorf(ctx, "taxonomyRelationsRepo.BatchCreate error: %v", err)
		return nil, err
	}
	return rows, nil
}

// FindRelations finds taxonomy relations by various criteria.
func (r *taxonomyRelationsRepository) FindRelations(ctx context.Context, params *structs.FindTaxonomyRelationParams) ([]*ent.TaxonomyRelation, error) {

	// create builder.
	builder := r.ecr.TaxonomyRelation.Query()

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
		logger.Errorf(ctx, "taxonomyRelationsRepo.FindRelations error: %v", err)
		return nil, err
	}

	return rows, nil
}

// FindTaxonomyRelation gets a single taxonomy relation by criteria.
func (r *taxonomyRelationsRepository) FindTaxonomyRelation(ctx context.Context, params *structs.FindTaxonomyRelation) (*ent.TaxonomyRelation, error) {

	// create builder.
	builder := r.ecr.TaxonomyRelation.Query()

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
