package repository

import (
	"context"
	"fmt"
	"ncobase/common/log"
	"ncobase/common/nanoid"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/access/data"
	"ncobase/feature/access/data/ent"
	casbinRuleEnt "ncobase/feature/access/data/ent/casbinrule"
	"ncobase/feature/access/structs"
)

// CasbinRuleRepositoryInterface represents the Casbin rule repository interface.
type CasbinRuleRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CasbinRuleBody) (*ent.CasbinRule, error)
	GetByID(ctx context.Context, id string) (*ent.CasbinRule, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.CasbinRule, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*ent.CasbinRule, error)
	Find(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error)
	List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error)
	CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int
}

// casbinRuleRepository implements the CasbinRuleRepositoryInterface.
type casbinRuleRepository struct {
	ec *ent.Client
}

// NewCasbinRule creates a new Casbin rule repository.
func NewCasbinRule(d *data.Data) CasbinRuleRepositoryInterface {
	return &casbinRuleRepository{ec: d.GetEntClient()}
}

// Create creates a new Casbin rule.
func (r *casbinRuleRepository) Create(ctx context.Context, body *structs.CasbinRuleBody) (*ent.CasbinRule, error) {
	// create builder.
	builder := r.ec.CasbinRule.Create()

	// set values.
	builder.SetNillablePType(&body.PType)
	builder.SetNillableV0(&body.V0)
	builder.SetNillableV1(&body.V1)
	builder.SetNillableV2(&body.V2)
	builder.SetNillableV3(body.V3)
	builder.SetNillableV4(body.V4)
	builder.SetNillableV5(body.V5)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// GetByID gets a Casbin rule by ID.
func (r *casbinRuleRepository) GetByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		log.Errorf(ctx, "casbinRuleRepo.GetByID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// Update updates a Casbin rule (full or partial).
func (r *casbinRuleRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.CasbinRule, error) {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// create builder.
	builder := row.Update()

	// Update the Casbin rule fields based on updates map
	for field, value := range updates {
		switch field {
		case "p_type":
			builder = builder.SetNillablePType(types.ToPointer(value.(string)))
		case "v0":
			builder = builder.SetNillableV0(types.ToPointer(value.(string)))
		case "v1":
			builder = builder.SetNillableV1(types.ToPointer(value.(string)))
		case "v2":
			builder = builder.SetNillableV2(types.ToPointer(value.(string)))
		case "v3":
			builder = builder.SetNillableV3(types.ToPointer(value.(string)))
		case "v4":
			builder = builder.SetNillableV4(types.ToPointer(value.(string)))
		case "v5":
			builder = builder.SetNillableV5(types.ToPointer(value.(string)))
		case "updated_by":
			builder = builder.SetNillableUpdatedBy(types.ToPointer(value.(string)))
		}
	}

	// Save the updated Casbin rule
	updatedRow, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return updatedRow, nil
}

// Delete deletes a Casbin rule by ID.
func (r *casbinRuleRepository) Delete(ctx context.Context, id string) error {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.CasbinRule.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(casbinRuleEnt.IDEQ(row.ID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "casbinRuleRepo.Delete error: %v\n", err)
		return err
	}

	return nil
}

// FindByID finds a Casbin rule by ID.
func (r *casbinRuleRepository) FindByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
	// create builder.
	builder := r.ec.CasbinRule.Query()

	// Add conditions to the query
	builder = builder.Where(casbinRuleEnt.IDEQ(id))

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// Find finds Casbin rules based on query parameters.
func (r *casbinRuleRepository) Find(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(params.Limit)

	// Execute the query
	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// List gets a list of Casbin rules.
func (r *casbinRuleRepository) List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}
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
				casbinRuleEnt.Or(
					casbinRuleEnt.CreatedAtGT(timestamp),
					casbinRuleEnt.And(
						casbinRuleEnt.CreatedAtEQ(timestamp),
						casbinRuleEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				casbinRuleEnt.Or(
					casbinRuleEnt.CreatedAtLT(timestamp),
					casbinRuleEnt.And(
						casbinRuleEnt.CreatedAtEQ(timestamp),
						casbinRuleEnt.IDLT(id),
					),
				),
			)
		}
	}

	if params.Direction == "backward" {
		builder.Order(ent.Asc(casbinRuleEnt.FieldCreatedAt), ent.Asc(casbinRuleEnt.FieldID))
	} else {
		builder.Order(ent.Desc(casbinRuleEnt.FieldCreatedAt), ent.Desc(casbinRuleEnt.FieldID))
	}

	builder.Offset(params.Offset)
	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "casbinRuleRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// CountX gets a count of Casbin rules.
func (r *casbinRuleRepository) CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder builds the list query.
func (r *casbinRuleRepository) listBuilder(_ context.Context, params *structs.ListCasbinRuleParams) (*ent.CasbinRuleQuery, error) {
	// create list builder
	builder := r.ec.CasbinRule.Query()

	// Add conditions to the query based on parameters
	if params.PType != nil && *params.PType != "" {
		builder = builder.Where(casbinRuleEnt.PTypeEQ(*params.PType))
	}
	if params.V0 != nil && *params.V0 != "" {
		builder = builder.Where(casbinRuleEnt.V0EQ(*params.V0))
	}
	if params.V1 != nil && *params.V1 != "" {
		builder = builder.Where(casbinRuleEnt.V1EQ(*params.V1))
	}
	if params.V2 != nil && *params.V2 != "" {
		builder = builder.Where(casbinRuleEnt.V2EQ(*params.V2))
	}
	if params.V3 != nil && *params.V3 != "" {
		builder = builder.Where(casbinRuleEnt.V3EQ(*params.V3))
	}
	if params.V4 != nil && *params.V4 != "" {
		builder = builder.Where(casbinRuleEnt.V4EQ(*params.V4))
	}
	if params.V5 != nil && *params.V5 != "" {
		builder = builder.Where(casbinRuleEnt.V5EQ(*params.V5))
	}

	return builder, nil
}
