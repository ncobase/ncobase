package repository

import (
	"context"
	"fmt"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	optionsEnt "ncobase/core/system/data/ent/options"
	"ncobase/core/system/structs"
	"ncore/pkg/data/meili"
	"ncore/pkg/logger"
	"ncore/pkg/paging"
	"ncore/pkg/validator"

	"github.com/redis/go-redis/v9"
)

// OptionsRepositoryInterface represents the options repository interface.
type OptionsRepositoryInterface interface {
	Create(context.Context, *structs.OptionsBody) (*ent.Options, error)
	Get(context.Context, *structs.FindOptions) (*ent.Options, error)
	Update(context.Context, *structs.UpdateOptionsBody) (*ent.Options, error)
	Delete(context.Context, *structs.FindOptions) error
	List(context.Context, *structs.ListOptionsParams) ([]*ent.Options, error)
	ListWithCount(ctx context.Context, params *structs.ListOptionsParams) ([]*ent.Options, int, error)
	CountX(context.Context, *structs.ListOptionsParams) int
}

// optionsRepository implements the OptionsRepositoryInterface.
type optionsRepository struct {
	ec *ent.Client
	rc *redis.Client
	ms *meili.Client
}

// NewOptionsRepository creates a new options repository.
func NewOptionsRepository(d *data.Data) OptionsRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	ms := d.GetMeilisearch()
	return &optionsRepository{
		ec: ec,
		rc: rc,
		ms: ms,
	}
}

// Create creates a new option.
func (r *optionsRepository) Create(ctx context.Context, body *structs.OptionsBody) (*ent.Options, error) {
	builder := r.ec.Options.Create()

	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotNil(body.Autoload) {
		builder.SetAutoload(body.Autoload)
	}
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Create error: %v", err)
		return nil, err
	}

	if err = r.ms.IndexDocuments("options", row); err != nil {
		logger.Errorf(ctx, "optionsRepo.Create error creating Meilisearch index: %v", err)
	}

	return row, nil
}

// Get retrieves a specific option.
func (r *optionsRepository) Get(ctx context.Context, params *structs.FindOptions) (*ent.Options, error) {
	row, err := r.getOption(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Get error: %v", err)
		return nil, err
	}
	return row, nil
}

// Update updates an existing option.
func (r *optionsRepository) Update(ctx context.Context, body *structs.UpdateOptionsBody) (*ent.Options, error) {
	row, err := r.getOption(ctx, &structs.FindOptions{
		Option: body.ID,
	})
	if validator.IsNotNil(err) {
		return nil, err
	}

	builder := row.Update()

	if validator.IsNotEmpty(body.Name) {
		builder.SetNillableName(&body.Name)
	}
	if validator.IsNotEmpty(body.Type) {
		builder.SetNillableType(&body.Type)
	}
	if validator.IsNotEmpty(body.Value) {
		builder.SetNillableValue(&body.Value)
	}
	if validator.IsNotNil(body.Autoload) {
		builder.SetAutoload(body.Autoload)
	}
	if validator.IsNotEmpty(body.TenantID) {
		builder.SetNillableTenantID(&body.TenantID)
	}

	row, err = builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.Update error: %v", err)
		return nil, err
	}

	return row, nil
}

// Delete deletes an option.
func (r *optionsRepository) Delete(ctx context.Context, params *structs.FindOptions) error {
	builder := r.ec.Options.Delete()

	builder.Where(optionsEnt.Or(
		optionsEnt.IDEQ(params.Option),
		optionsEnt.NameEQ(params.Option),
	))

	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(optionsEnt.TenantIDEQ(params.Tenant))
	}

	_, err := builder.Exec(ctx)
	if validator.IsNotNil(err) {
		return err
	}

	return nil
}

// List returns a slice of options based on the provided parameters.
func (r *optionsRepository) List(ctx context.Context, params *structs.ListOptionsParams) ([]*ent.Options, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("building list query: %w", err)
	}

	builder = applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	builder.Limit(params.Limit)

	return r.executeArrayQuery(ctx, builder)
}

// CountX returns the total count of options based on the provided parameters.
func (r *optionsRepository) CountX(ctx context.Context, params *structs.ListOptionsParams) int {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "Error building count query: %v", err)
		return 0
	}
	return builder.CountX(ctx)
}

// ListWithCount returns both a slice of options and the total count based on the provided parameters.
func (r *optionsRepository) ListWithCount(ctx context.Context, params *structs.ListOptionsParams) ([]*ent.Options, int, error) {
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("building list query: %w", err)
	}

	builder = applySorting(builder, params.SortBy)

	if params.Cursor != "" {
		id, value, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("decoding cursor: %w", err)
		}
		builder = applyCursorCondition(builder, id, value, params.Direction, params.SortBy)
	}

	total, err := builder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting options: %w", err)
	}

	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("fetching options: %w", err)
	}

	return rows, total, nil
}

// applySorting applies the specified sorting to the query builder.
func applySorting(builder *ent.OptionsQuery, sortBy string) *ent.OptionsQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		return builder.Order(ent.Desc(optionsEnt.FieldCreatedAt), ent.Desc(optionsEnt.FieldID))
	case structs.SortByName:
		return builder.Order(ent.Asc(optionsEnt.FieldName), ent.Desc(optionsEnt.FieldID))
	default:
		return builder.Order(ent.Desc(optionsEnt.FieldCreatedAt), ent.Desc(optionsEnt.FieldID))
	}
}

// applyCursorCondition applies the cursor-based condition to the query builder.
func applyCursorCondition(builder *ent.OptionsQuery, id string, value any, direction string, sortBy string) *ent.OptionsQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				optionsEnt.Or(
					optionsEnt.CreatedAtGT(timestamp),
					optionsEnt.And(
						optionsEnt.CreatedAtEQ(timestamp),
						optionsEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			optionsEnt.Or(
				optionsEnt.CreatedAtLT(timestamp),
				optionsEnt.And(
					optionsEnt.CreatedAtEQ(timestamp),
					optionsEnt.IDLT(id),
				),
			),
		)
	case structs.SortByName:
		name, ok := value.(string)
		if !ok {
			logger.Errorf(context.Background(), "Invalid name value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				optionsEnt.Or(
					optionsEnt.NameGT(name),
					optionsEnt.And(
						optionsEnt.NameEQ(name),
						optionsEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			optionsEnt.Or(
				optionsEnt.NameLT(name),
				optionsEnt.And(
					optionsEnt.NameEQ(name),
					optionsEnt.IDLT(id),
				),
			),
		)
	default:
		return applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// listBuilder - create list builder.
func (r *optionsRepository) listBuilder(_ context.Context, params *structs.ListOptionsParams) (*ent.OptionsQuery, error) {
	builder := r.ec.Options.Query()

	if params.Tenant != "" {
		builder.Where(optionsEnt.TenantIDEQ(params.Tenant))
	}

	if params.Type != "" {
		builder.Where(optionsEnt.TypeEQ(params.Type))
	}

	if params.Autoload != nil {
		builder.Where(optionsEnt.AutoloadEQ(*params.Autoload))
	}

	return builder, nil
}

// getOption - get option.
// internal method.
func (r *optionsRepository) getOption(ctx context.Context, params *structs.FindOptions) (*ent.Options, error) {
	builder := r.ec.Options.Query()

	if validator.IsNotEmpty(params.Option) {
		builder.Where(optionsEnt.Or(
			optionsEnt.IDEQ(params.Option),
			optionsEnt.NameEQ(params.Option),
		))
	}
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(optionsEnt.TenantIDEQ(params.Tenant))
	}

	row, err := builder.First(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// executeArrayQuery - execute the builder query and return results.
func (r *optionsRepository) executeArrayQuery(ctx context.Context, builder *ent.OptionsQuery) ([]*ent.Options, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "optionsRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}
