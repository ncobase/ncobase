package repository

import (
	"context"
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/data/ent"
	ednpointEnt "ncobase/proxy/data/ent/endpoint"
	"ncobase/proxy/structs"
	"strings"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/redis/go-redis/v9"
)

// EndpointRepositoryInterface is the interface for the endpoint repository.
type EndpointRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateEndpointBody) (*ent.Endpoint, error)
	GetByID(ctx context.Context, id string) (*ent.Endpoint, error)
	GetByName(ctx context.Context, name string) (*ent.Endpoint, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Endpoint, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListEndpointParams) ([]*ent.Endpoint, error)
	CountX(ctx context.Context, params *structs.ListEndpointParams) int
}

// endpointRepository implements the EndpointRepositoryInterface.
type endpointRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Endpoint]
}

// NewEndpointRepository creates a new endpoint repository.
func NewEndpointRepository(d *data.Data) EndpointRepositoryInterface {
	ec := d.GetMasterEntClient()
	rc := d.GetRedis()
	return &endpointRepository{ec, rc, cache.NewCache[ent.Endpoint](rc, "ncse_proxy_endpoint")}
}

// Create creates a new endpoint.
func (r *endpointRepository) Create(ctx context.Context, body *structs.CreateEndpointBody) (*ent.Endpoint, error) {
	// Create builder
	builder := r.ec.Endpoint.Create()

	// Set values
	builder.SetNillableName(&body.Name)
	builder.SetNillableDescription(&body.Description)
	builder.SetBaseURL(body.BaseURL)
	builder.SetNillableProtocol(&body.Protocol)
	builder.SetNillableAuthType(&body.AuthType)

	if body.AuthConfig != nil {
		builder.SetAuthConfig(*body.AuthConfig)
	}

	builder.SetTimeout(body.Timeout)
	builder.SetUseCircuitBreaker(body.UseCircuitBreaker)
	builder.SetRetryCount(body.RetryCount)
	builder.SetValidateSsl(body.ValidateSSL)
	builder.SetLogRequests(body.LogRequests)
	builder.SetLogResponses(body.LogResponses)
	builder.SetDisabled(body.Disabled)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	builder.SetNillableCreatedBy(body.CreatedBy)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets an endpoint by ID.
func (r *endpointRepository) GetByID(ctx context.Context, id string) (*ent.Endpoint, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Endpoint.Query()

	// Set conditions
	builder = builder.Where(ednpointEnt.IDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByName gets an endpoint by name.
func (r *endpointRepository) GetByName(ctx context.Context, name string) (*ent.Endpoint, error) {
	// Check cache
	cacheKey := fmt.Sprintf("name:%s", name)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Endpoint.Query()

	// Set conditions
	builder = builder.Where(ednpointEnt.NameEQ(name))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.GetByName error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.GetByName cache error: %v", err)
	}

	return row, nil
}

// Update updates an endpoint.
func (r *endpointRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Endpoint, error) {
	// Get the endpoint
	endpoint, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create builder
	builder := endpoint.Update()

	// Apply updates
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "base_url":
			builder.SetNillableBaseURL(convert.ToPointer(value.(string)))
		case "protocol":
			builder.SetNillableProtocol(convert.ToPointer(value.(string)))
		case "auth_type":
			builder.SetNillableAuthType(convert.ToPointer(value.(string)))
		case "auth_config":
			builder.SetAuthConfig(value.(string))
		case "timeout":
			builder.SetTimeout(int(value.(float64)))
		case "use_circuit_breaker":
			builder.SetUseCircuitBreaker(value.(bool))
		case "retry_count":
			builder.SetRetryCount(int(value.(float64)))
		case "validate_ssl":
			builder.SetValidateSsl(value.(bool))
		case "log_requests":
			builder.SetLogRequests(value.(bool))
		case "log_responses":
			builder.SetLogResponses(value.(bool))
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
		logger.Errorf(ctx, "endpointRepo.Update error: %v", err)
		return nil, err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", row.Name))

	return row, nil
}

// Delete deletes an endpoint.
func (r *endpointRepository) Delete(ctx context.Context, id string) error {
	// Get the endpoint first to clear cache later
	endpoint, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Create builder
	builder := r.ec.Endpoint.Delete()

	// Set conditions
	builder = builder.Where(ednpointEnt.IDEQ(id))

	// Execute the builder
	_, err = builder.Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.Delete error: %v", err)
		return err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", id)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", endpoint.Name))

	return nil
}

// List lists endpoints with pagination.
func (r *endpointRepository) List(ctx context.Context, params *structs.ListEndpointParams) ([]*ent.Endpoint, error) {
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
				ednpointEnt.Or(
					ednpointEnt.CreatedAtGT(timestamp),
					ednpointEnt.And(
						ednpointEnt.CreatedAtEQ(timestamp),
						ednpointEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				ednpointEnt.Or(
					ednpointEnt.CreatedAtLT(timestamp),
					ednpointEnt.And(
						ednpointEnt.CreatedAtEQ(timestamp),
						ednpointEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Set order
	if params.Direction == "backward" {
		builder.Order(ent.Asc(ednpointEnt.FieldCreatedAt), ent.Asc(ednpointEnt.FieldID))
	} else {
		builder.Order(ent.Desc(ednpointEnt.FieldCreatedAt), ent.Desc(ednpointEnt.FieldID))
	}

	// Set limit
	builder.Limit(params.Limit)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "endpointRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX counts endpoints.
func (r *endpointRepository) CountX(ctx context.Context, params *structs.ListEndpointParams) int {
	// Create builder for count
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder creates a builder for listing endpoints.
func (r *endpointRepository) listBuilder(ctx context.Context, params *structs.ListEndpointParams) (*ent.EndpointQuery, error) {
	// Create builder
	builder := r.ec.Endpoint.Query()

	// Apply filters
	if params.Name != "" {
		builder = builder.Where(ednpointEnt.NameContainsFold(params.Name))
	}

	if params.Protocol != "" {
		protocols := strings.Split(params.Protocol, ",")
		for i := range protocols {
			protocols[i] = strings.TrimSpace(protocols[i])
		}
		builder = builder.Where(ednpointEnt.ProtocolIn(protocols...))
	}

	if params.Disabled != nil {
		builder = builder.Where(ednpointEnt.DisabledEQ(*params.Disabled))
	}

	return builder, nil
}
