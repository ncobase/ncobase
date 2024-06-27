package repo

import (
	"context"
	"ncobase/common/log"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	casbinRuleEnt "ncobase/internal/data/ent/casbinrule"
	"ncobase/internal/data/structs"
)

// CasbinRule represents the Casbin rule repository interface.
type CasbinRule interface {
	Create(ctx context.Context, body *structs.CasbinRuleBody) (*ent.CasbinRule, error)
	GetByID(ctx context.Context, id string) (*ent.CasbinRule, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.CasbinRule, error)
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*ent.CasbinRule, error)
	Find(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error)
	List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error)
	CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int
}

// casbinRuleRepo implements the Casbin rule interface.
type casbinRuleRepo struct {
	ec *ent.Client
}

// NewCasbinRule creates a new Casbin rule repository.
func NewCasbinRule(d *data.Data) CasbinRule {
	return &casbinRuleRepo{ec: d.GetEntClient()}
}

// Create creates a new Casbin rule.
func (r *casbinRuleRepo) Create(ctx context.Context, body *structs.CasbinRuleBody) (*ent.CasbinRule, error) {
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
func (r *casbinRuleRepo) GetByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
	row, err := r.FindByID(ctx, id)
	if err != nil {
		log.Errorf(ctx, "casbinRuleRepo.GetByID error: %v\n", err)
		return nil, err
	}
	return row, nil
}

// Update updates a Casbin rule (full or partial).
func (r *casbinRuleRepo) Update(ctx context.Context, id string, updates types.JSON) (*ent.CasbinRule, error) {
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
func (r *casbinRuleRepo) Delete(ctx context.Context, id string) error {
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
func (r *casbinRuleRepo) FindByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
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
func (r *casbinRuleRepo) Find(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	// Execute the query
	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// List gets a list of Casbin rules.
func (r *casbinRuleRepo) List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "casbinRuleRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// CountX gets a count of Casbin rules.
func (r *casbinRuleRepo) CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int {
	// create list builder
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder builds the list query.
func (r *casbinRuleRepo) listBuilder(ctx context.Context, params *structs.ListCasbinRuleParams) (*ent.CasbinRuleQuery, error) {

	var next *ent.CasbinRule
	if validator.IsNotEmpty(params.Cursor) {
		// query the role.
		row, err := r.FindByID(ctx, params.Cursor)
		if validator.IsNotNil(err) || validator.IsNil(row) {
			return nil, err
		}
		next = row
	}

	// create list builder
	builder := r.ec.CasbinRule.Query()

	// lt the cursor create time
	if next != nil {
		builder.Where(casbinRuleEnt.CreatedAtLT(next.CreatedAt))
	}

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
