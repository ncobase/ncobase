package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/data/ent"
	"ncobase/proxy/data/repository"
	"ncobase/proxy/structs"
	"path"
	"regexp"
	"strings"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"
)

// RouteServiceInterface is the interface for the route service.
type RouteServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateRouteBody) (*structs.ReadRoute, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadRoute, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*structs.ReadRoute, error)
	GetByName(ctx context.Context, name string) (*structs.ReadRoute, error)
	FindByPathAndMethod(ctx context.Context, path, method string) (*structs.ReadRoute, error)
	List(ctx context.Context, params *structs.ListRouteParams) (paging.Result[*structs.ReadRoute], error)
	Serialize(row *ent.Route) *structs.ReadRoute
	Serializes(rows []*ent.Route) []*structs.ReadRoute
}

// routeService is the struct for the route service.
type routeService struct {
	route    repository.RouteRepositoryInterface
	endpoint repository.EndpointRepositoryInterface
}

// NewRouteService creates a new route service.
func NewRouteService(d *data.Data) RouteServiceInterface {
	return &routeService{
		route:    repository.NewRouteRepository(d),
		endpoint: repository.NewEndpointRepository(d),
	}
}

// Create creates a new route.
func (s *routeService) Create(ctx context.Context, body *structs.CreateRouteBody) (*structs.ReadRoute, error) {
	if body.Name == "" {
		return nil, errors.New("route name is required")
	}

	// Validate endpoint exists
	_, err := s.endpoint.GetByID(ctx, body.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("endpoint not found: %w", err)
	}

	// Validate path pattern format
	if err := validatePathPattern(body.PathPattern); err != nil {
		return nil, err
	}

	// Normalize route method to uppercase
	body.Method = strings.ToUpper(body.Method)

	// Check if input transformer exists if specified
	if body.InputTransformerID != nil && *body.InputTransformerID != "" {
		// This would be handled by a foreign key constraint in the database
	}

	// Check if output transformer exists if specified
	if body.OutputTransformerID != nil && *body.OutputTransformerID != "" {
		// This would be handled by a foreign key constraint in the database
	}

	row, err := s.route.Create(ctx, body)
	if err := handleEntError(ctx, "Route", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing route.
func (s *routeService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadRoute, error) {
	if validator.IsEmpty(id) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	// Validate path pattern if it's being updated
	if pathPattern, ok := updates["path_pattern"].(string); ok {
		if err := validatePathPattern(pathPattern); err != nil {
			return nil, err
		}
	}

	// Normalize route method to uppercase if it's being updated
	if method, ok := updates["method"].(string); ok {
		updates["method"] = strings.ToUpper(method)
	}

	row, err := s.route.Update(ctx, id, updates)
	if err := handleEntError(ctx, "Route", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a route by ID.
func (s *routeService) Delete(ctx context.Context, id string) error {
	err := s.route.Delete(ctx, id)
	if err := handleEntError(ctx, "Route", err); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a route by ID.
func (s *routeService) GetByID(ctx context.Context, id string) (*structs.ReadRoute, error) {
	row, err := s.route.GetByID(ctx, id)
	if err := handleEntError(ctx, "Route", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByName retrieves a route by name.
func (s *routeService) GetByName(ctx context.Context, name string) (*structs.ReadRoute, error) {
	row, err := s.route.GetByName(ctx, name)
	if err := handleEntError(ctx, "Route", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// FindByPathAndMethod finds a route by path and method.
func (s *routeService) FindByPathAndMethod(ctx context.Context, path, method string) (*structs.ReadRoute, error) {
	// Normalize method to uppercase
	method = strings.ToUpper(method)

	// First try to find an exact match for the method
	routes, err := s.route.FindByPathPattern(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("no route found for path %s: %w", path, err)
	}

	// Check for method match
	for _, route := range routes {
		routeMethod := strings.ToUpper(route.Method)
		if routeMethod == method || routeMethod == "ANY" || routeMethod == "*" {
			return s.Serialize(route), nil
		}
	}

	// If we get here, no route matched the method
	return nil, fmt.Errorf("no route found for path %s and method %s", path, method)
}

// List lists all routes.
func (s *routeService) List(ctx context.Context, params *structs.ListRouteParams) (paging.Result[*structs.ReadRoute], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadRoute, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.route.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing routes: %v", err)
			return nil, 0, err
		}

		total := s.route.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes a list of route entities to a response format.
func (s *routeService) Serializes(rows []*ent.Route) []*structs.ReadRoute {
	rs := make([]*structs.ReadRoute, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes a route entity to a response format.
func (s *routeService) Serialize(row *ent.Route) *structs.ReadRoute {
	return &structs.ReadRoute{
		ID:                  row.ID,
		Name:                row.Name,
		Description:         row.Description,
		EndpointID:          row.EndpointID,
		PathPattern:         row.PathPattern,
		TargetPath:          row.TargetPath,
		Method:              row.Method,
		InputTransformerID:  &row.InputTransformerID,
		OutputTransformerID: &row.OutputTransformerID,
		CacheEnabled:        row.CacheEnabled,
		CacheTTL:            row.CacheTTL,
		RateLimit:           &row.RateLimit,
		StripAuthHeader:     row.StripAuthHeader,
		Disabled:            row.Disabled,
		Extras:              &row.Extras,
		CreatedBy:           &row.CreatedBy,
		CreatedAt:           &row.CreatedAt,
		UpdatedBy:           &row.UpdatedBy,
		UpdatedAt:           &row.UpdatedAt,
	}
}

// validatePathPattern validates the format of the path pattern.
func validatePathPattern(pathPattern string) error {
	// Check if the path starts with a slash
	if !strings.HasPrefix(pathPattern, "/") {
		return fmt.Errorf("path pattern must start with a slash: %s", pathPattern)
	}

	// Check for invalid characters
	invalidCharsRegex := regexp.MustCompile(`[^a-zA-Z0-9/_\-\.:\*]`)
	if invalidCharsRegex.MatchString(pathPattern) {
		return fmt.Errorf("path pattern contains invalid characters: %s", pathPattern)
	}

	// Validate parameter format
	parts := strings.Split(pathPattern, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			if len(paramName) == 0 {
				return fmt.Errorf("empty parameter name in path pattern: %s", pathPattern)
			}
			// Check that parameter name only contains valid characters
			paramNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			if !paramNameRegex.MatchString(paramName) {
				return fmt.Errorf("invalid parameter name in path pattern: %s", paramName)
			}
		}
	}

	// Clean the path to handle multiple slashes and trailing slashes
	cleanPath := path.Clean(pathPattern)
	if pathPattern != cleanPath && pathPattern != cleanPath+"/" {
		return fmt.Errorf("invalid path pattern format (should be %s): %s", cleanPath, pathPattern)
	}

	return nil
}
