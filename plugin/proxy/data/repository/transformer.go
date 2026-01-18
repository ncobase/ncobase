package repository

import (
	"context"
	"fmt"
	"ncobase/plugin/proxy/data"
	"ncobase/plugin/proxy/data/ent"
	transformerEnt "ncobase/plugin/proxy/data/ent/transformer"
	"ncobase/plugin/proxy/structs"
	"strings"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// TransformerRepositoryInterface is the interface for the transformer repository.
type TransformerRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateTransformerBody) (*ent.Transformer, error)
	GetByID(ctx context.Context, id string) (*ent.Transformer, error)
	GetByName(ctx context.Context, name string) (*ent.Transformer, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Transformer, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTransformerParams) ([]*ent.Transformer, error)
	CountX(ctx context.Context, params *structs.ListTransformerParams) int
}

// transformerRepository implements the TransformerRepositoryInterface.
type transformerRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Transformer]
}

// NewTransformerRepository creates a new transformer repository.
func NewTransformerRepository(d *data.Data) TransformerRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis().(*redis.Client)
	return &transformerRepository{ec, rc, cache.NewCache[ent.Transformer](rc, "ncse_proxy_transformer")}
}

// Create creates a new transformerEnt.
func (r *transformerRepository) Create(ctx context.Context, body *structs.CreateTransformerBody) (*ent.Transformer, error) {
	// Create builder
	builder := r.ec.Transformer.Create()

	// Set values
	builder.SetNillableName(&body.Name)
	builder.SetNillableDescription(&body.Description)
	builder.SetType(body.Type)
	builder.SetContent(body.Content)
	builder.SetNillableContentType(&body.ContentType)
	builder.SetDisabled(body.Disabled)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	builder.SetNillableCreatedBy(body.CreatedBy)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a transformer by ID.
func (r *transformerRepository) GetByID(ctx context.Context, id string) (*ent.Transformer, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Transformer.Query()

	// Set conditions
	builder = builder.Where(transformerEnt.IDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByName gets a transformer by name.
func (r *transformerRepository) GetByName(ctx context.Context, name string) (*ent.Transformer, error) {
	// Check cache
	cacheKey := fmt.Sprintf("name:%s", name)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Transformer.Query()

	// Set conditions
	builder = builder.Where(transformerEnt.NameEQ(name))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.GetByName error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.GetByName cache error: %v", err)
	}

	return row, nil
}

// Update updates a transformerEnt.
func (r *transformerRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Transformer, error) {
	// Get the transformer
	transformer, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create builder
	builder := transformer.Update()

	// Apply updates
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
		case "content":
			builder.SetNillableContent(convert.ToPointer(value.(string)))
		case "content_type":
			builder.SetNillableContentType(convert.ToPointer(value.(string)))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "extras":
			builder.SetExtras(value.(types.JSON))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.Update error: %v", err)
		return nil, err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", row.Name))

	return row, nil
}

// Delete deletes a transformerEnt.
func (r *transformerRepository) Delete(ctx context.Context, id string) error {
	// Get the transformer first to clear cache later
	transformer, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Create builder
	builder := r.ec.Transformer.Delete()

	// Set conditions
	builder = builder.Where(transformerEnt.IDEQ(transformer.ID))

	// Execute the builder
	_, err = builder.Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.Delete error: %v", err)
		return err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", id)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", transformerEnt.Name))

	return nil
}

// List lists transformers.
func (r *transformerRepository) List(ctx context.Context, params *structs.ListTransformerParams) ([]*ent.Transformer, error) {
	// Create builder for list
	builder, err := r.listBuilder(ctx, params)
	if err != nil {
		return nil, err
	}

	// Apply cursor pagination
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
				transformerEnt.Or(
					transformerEnt.CreatedAtGT(timestamp),
					transformerEnt.And(
						transformerEnt.CreatedAtEQ(timestamp),
						transformerEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				transformerEnt.Or(
					transformerEnt.CreatedAtLT(timestamp),
					transformerEnt.And(
						transformerEnt.CreatedAtEQ(timestamp),
						transformerEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Set order
	if params.Direction == "backward" {
		builder.Order(ent.Asc(transformerEnt.FieldCreatedAt), ent.Asc(transformerEnt.FieldID))
	} else {
		builder.Order(ent.Desc(transformerEnt.FieldCreatedAt), ent.Desc(transformerEnt.FieldID))
	}

	// Set limit
	builder.Limit(params.Limit)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "transformerRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX counts transformers.
func (r *transformerRepository) CountX(ctx context.Context, params *structs.ListTransformerParams) int {
	// Create builder for count
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder creates a builder for listing transformers.
func (r *transformerRepository) listBuilder(ctx context.Context, params *structs.ListTransformerParams) (*ent.TransformerQuery, error) {
	// Create builder
	builder := r.ec.Transformer.Query()

	// Apply filters
	if params.Name != "" {
		builder = builder.Where(transformerEnt.NameContainsFold(params.Name))
	}

	if params.Type != "" {
		types := strings.Split(params.Type, ",")
		for i := range types {
			types[i] = strings.TrimSpace(types[i])
		}
		builder = builder.Where(transformerEnt.TypeIn(types...))
	}

	if params.ContentType != "" {
		contentTypes := strings.Split(params.ContentType, ",")
		for i := range contentTypes {
			contentTypes[i] = strings.TrimSpace(contentTypes[i])
		}
		builder = builder.Where(transformerEnt.ContentTypeIn(contentTypes...))
	}

	if params.Disabled != nil {
		builder = builder.Where(transformerEnt.DisabledEQ(*params.Disabled))
	}

	return builder, nil
}
