package service

import (
	"context"
	"errors"
	"ncobase/plugin/proxy/data"
	"ncobase/plugin/proxy/data/ent"
	"ncobase/plugin/proxy/data/repository"
	"ncobase/plugin/proxy/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// EndpointServiceInterface is the interface for the endpoint service.
type EndpointServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateEndpointBody) (*structs.ReadEndpoint, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadEndpoint, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*structs.ReadEndpoint, error)
	GetByName(ctx context.Context, name string) (*structs.ReadEndpoint, error)
	List(ctx context.Context, params *structs.ListEndpointParams) (paging.Result[*structs.ReadEndpoint], error)
	Serialize(row *ent.Endpoint) *structs.ReadEndpoint
	Serializes(rows []*ent.Endpoint) []*structs.ReadEndpoint
}

// endpointService is the struct for the endpoint service.
type endpointService struct {
	endpoint repository.EndpointRepositoryInterface
}

// NewEndpointService creates a new endpoint service.
func NewEndpointService(d *data.Data) EndpointServiceInterface {
	return &endpointService{
		endpoint: repository.NewEndpointRepository(d),
	}
}

// Create creates a new endpoint.
func (s *endpointService) Create(ctx context.Context, body *structs.CreateEndpointBody) (*structs.ReadEndpoint, error) {
	if body.Name == "" {
		return nil, errors.New("endpoint name is required")
	}

	row, err := s.endpoint.Create(ctx, body)
	if err := handleEntError(ctx, "Endpoint", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing endpoint.
func (s *endpointService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadEndpoint, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.endpoint.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Endpoint", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes an endpoint by ID.
func (s *endpointService) Delete(ctx context.Context, id string) error {
	err := s.endpoint.Delete(ctx, id)
	if err := handleEntError(ctx, "Endpoint", err); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves an endpoint by ID.
func (s *endpointService) GetByID(ctx context.Context, id string) (*structs.ReadEndpoint, error) {
	row, err := s.endpoint.GetByID(ctx, id)
	if err := handleEntError(ctx, "Endpoint", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByName retrieves an endpoint by name.
func (s *endpointService) GetByName(ctx context.Context, name string) (*structs.ReadEndpoint, error) {
	row, err := s.endpoint.GetByName(ctx, name)
	if err := handleEntError(ctx, "Endpoint", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// List lists all endpoints.
func (s *endpointService) List(ctx context.Context, params *structs.ListEndpointParams) (paging.Result[*structs.ReadEndpoint], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadEndpoint, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.endpoint.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing endpoints: %v", err)
			return nil, 0, err
		}

		total := s.endpoint.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes a list of endpoint entities to a response format.
func (s *endpointService) Serializes(rows []*ent.Endpoint) []*structs.ReadEndpoint {
	rs := make([]*structs.ReadEndpoint, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes an endpoint entity to a response format.
func (s *endpointService) Serialize(row *ent.Endpoint) *structs.ReadEndpoint {
	return &structs.ReadEndpoint{
		ID:                row.ID,
		Name:              row.Name,
		Description:       row.Description,
		BaseURL:           row.BaseURL,
		Protocol:          row.Protocol,
		AuthType:          row.AuthType,
		AuthConfig:        &row.AuthConfig,
		Timeout:           row.Timeout,
		UseCircuitBreaker: row.UseCircuitBreaker,
		RetryCount:        row.RetryCount,
		ValidateSSL:       row.ValidateSsl,
		LogRequests:       row.LogRequests,
		LogResponses:      row.LogResponses,
		Disabled:          row.Disabled,
		Extras:            &row.Extras,
		CreatedBy:         &row.CreatedBy,
		CreatedAt:         &row.CreatedAt,
		UpdatedBy:         &row.UpdatedBy,
		UpdatedAt:         &row.UpdatedAt,
	}
}
