package repository

import (
	"context"
	"fmt"
	"ncobase/proxy/data"
	"ncobase/proxy/data/ent"
	routeEnt "ncobase/proxy/data/ent/route"
	"ncobase/proxy/structs"
	"strings"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/redis/go-redis/v9"
)

// RouteRepositoryInterface is the interface for the route repository.
type RouteRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateRouteBody) (*ent.Route, error)
	GetByID(ctx context.Context, id string) (*ent.Route, error)
	GetByName(ctx context.Context, name string) (*ent.Route, error)
	FindByPathPattern(ctx context.Context, path string) ([]*ent.Route, error)
	Update(ctx context.Context, id string, updates types.JSON) (*ent.Route, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListRouteParams) ([]*ent.Route, error)
	CountX(ctx context.Context, params *structs.ListRouteParams) int
}

// routeRepository implements the RouteRepositoryInterface.
type routeRepository struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Route]
}

// NewRouteRepository creates a new route repository.
func NewRouteRepository(d *data.Data) RouteRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &routeRepository{ec, rc, cache.NewCache[ent.Route](rc, "ncse_proxy_route")}
}

// Create creates a new routeEnt.
func (r *routeRepository) Create(ctx context.Context, body *structs.CreateRouteBody) (*ent.Route, error) {
	// Create builder
	builder := r.ec.Route.Create()

	// Set values
	builder.SetNillableName(&body.Name)
	builder.SetNillableDescription(&body.Description)
	builder.SetEndpointID(body.EndpointID)
	builder.SetPathPattern(body.PathPattern)
	builder.SetTargetPath(body.TargetPath)
	builder.SetNillableMethod(&body.Method)

	if body.InputTransformerID != nil {
		builder.SetInputTransformerID(*body.InputTransformerID)
	}

	if body.OutputTransformerID != nil {
		builder.SetOutputTransformerID(*body.OutputTransformerID)
	}

	builder.SetCacheEnabled(body.CacheEnabled)
	builder.SetCacheTTL(body.CacheTTL)

	if body.RateLimit != nil {
		builder.SetRateLimit(*body.RateLimit)
	}

	builder.SetStripAuthHeader(body.StripAuthHeader)
	builder.SetDisabled(body.Disabled)

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	builder.SetNillableCreatedBy(body.CreatedBy)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a route by ID.
func (r *routeRepository) GetByID(ctx context.Context, id string) (*ent.Route, error) {
	// Check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Route.Query()

	// Set conditions
	builder = builder.Where(routeEnt.IDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.GetByID cache error: %v", err)
	}

	return row, nil
}

// GetByName gets a route by name.
func (r *routeRepository) GetByName(ctx context.Context, name string) (*ent.Route, error) {
	// Check cache
	cacheKey := fmt.Sprintf("name:%s", name)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Create builder
	builder := r.ec.Route.Query()

	// Set conditions
	builder = builder.Where(routeEnt.NameEQ(name))

	// Execute the builder
	row, err := builder.Only(ctx)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.GetByName error: %v", err)
		return nil, err
	}

	// Cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.GetByName cache error: %v", err)
	}

	return row, nil
}

// FindByPathPattern finds routes by path pattern.
func (r *routeRepository) FindByPathPattern(ctx context.Context, path string) ([]*ent.Route, error) {
	// First try to find routes with exact match
	exactMatches, err := r.ec.Route.Query().
		Where(routeEnt.PathPatternEQ(path)).
		All(ctx)

	if err != nil && !ent.IsNotFound(err) {
		logger.Errorf(ctx, "routeRepo.FindByPathPattern error: %v", err)
		return nil, err
	}

	if len(exactMatches) > 0 {
		return exactMatches, nil
	}

	// If no exact matches, find routes with parameterized paths that could match
	// Note: This is a simplified pattern matching logic
	// In a real implementation, we would need more sophisticated path matching

	// Get all active routes
	routes, err := r.ec.Route.Query().
		Where(routeEnt.DisabledEQ(false)).
		All(ctx)

	if err != nil {
		logger.Errorf(ctx, "routeRepo.FindByPathPattern error: %v", err)
		return nil, err
	}

	var matches []*ent.Route

	// Check each route for parameterized pattern match
	for _, route := range routes {
		if pathMatches(path, route.PathPattern) {
			matches = append(matches, route)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no routes found matching path: %s", path)
	}

	return matches, nil
}

// Update updates a routeEnt.
func (r *routeRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.Route, error) {
	// Get the route
	route, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create builder
	builder := route.Update()

	// Apply updates
	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
		case "endpoint_id":
			builder.SetNillableEndpointID(types.ToPointer(value.(string)))
		case "path_pattern":
			builder.SetNillablePathPattern(types.ToPointer(value.(string)))
		case "target_path":
			builder.SetNillableTargetPath(types.ToPointer(value.(string)))
		case "method":
			builder.SetNillableMethod(types.ToPointer(value.(string)))
		case "input_transformer_id":
			if value == nil {
				builder.ClearInputTransformerID()
			} else {
				builder.SetInputTransformerID(value.(string))
			}
		case "output_transformer_id":
			if value == nil {
				builder.ClearOutputTransformerID()
			} else {
				builder.SetOutputTransformerID(value.(string))
			}
		case "cache_enabled":
			builder.SetCacheEnabled(value.(bool))
		case "cache_ttl":
			builder.SetCacheTTL(int(value.(float64)))
		case "rate_limit":
			if value == nil {
				builder.ClearRateLimit()
			} else {
				builder.SetRateLimit(value.(string))
			}
		case "strip_auth_header":
			builder.SetStripAuthHeader(value.(bool))
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
		logger.Errorf(ctx, "routeRepo.Update error: %v", err)
		return nil, err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", row.ID)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", row.Name))

	return row, nil
}

// Delete deletes a routeEnt.
func (r *routeRepository) Delete(ctx context.Context, id string) error {
	// Get the route first to clear cache later
	route, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Create builder
	builder := r.ec.Route.Delete()

	// Set conditions
	builder = builder.Where(routeEnt.IDEQ(route.ID))

	// Execute the builder
	_, err = builder.Exec(ctx)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.Delete error: %v", err)
		return err
	}

	// Clear cache
	cacheKey := fmt.Sprintf("%s", id)
	r.c.Delete(ctx, cacheKey)
	r.c.Delete(ctx, fmt.Sprintf("name:%s", routeEnt.Name))

	return nil
}

// List lists routes with pagination.
func (r *routeRepository) List(ctx context.Context, params *structs.ListRouteParams) ([]*ent.Route, error) {
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
				routeEnt.Or(
					routeEnt.CreatedAtGT(timestamp),
					routeEnt.And(
						routeEnt.CreatedAtEQ(timestamp),
						routeEnt.IDGT(id),
					),
				),
			)
		} else {
			builder.Where(
				routeEnt.Or(
					routeEnt.CreatedAtLT(timestamp),
					routeEnt.And(
						routeEnt.CreatedAtEQ(timestamp),
						routeEnt.IDLT(id),
					),
				),
			)
		}
	}

	// Set order
	if params.Direction == "backward" {
		builder.Order(ent.Asc(routeEnt.FieldCreatedAt), ent.Asc(routeEnt.FieldID))
	} else {
		builder.Order(ent.Desc(routeEnt.FieldCreatedAt), ent.Desc(routeEnt.FieldID))
	}

	// Set limit
	builder.Limit(params.Limit)

	// Execute the builder
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "routeRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// CountX counts routes.
func (r *routeRepository) CountX(ctx context.Context, params *structs.ListRouteParams) int {
	// Create builder for count
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder creates a builder for listing routes.
func (r *routeRepository) listBuilder(ctx context.Context, params *structs.ListRouteParams) (*ent.RouteQuery, error) {
	// Create builder
	builder := r.ec.Route.Query()

	// Apply filters
	if params.Name != "" {
		builder = builder.Where(routeEnt.NameContainsFold(params.Name))
	}

	if params.EndpointID != "" {
		builder = builder.Where(routeEnt.EndpointIDEQ(params.EndpointID))
	}

	if params.Method != "" {
		methods := strings.Split(params.Method, ",")
		for i := range methods {
			methods[i] = strings.TrimSpace(methods[i])
		}
		builder = builder.Where(routeEnt.MethodIn(methods...))
	}

	if params.Disabled != nil {
		builder = builder.Where(routeEnt.DisabledEQ(*params.Disabled))
	}

	return builder, nil
}

// pathMatches checks if a path matches a pattern.
// This is a simplified version - in a real implementation you would use a more
// sophisticated routing library or regular expressions.
func pathMatches(path, pattern string) bool {
	// Split path and pattern into segments
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")

	// If they have different number of segments, they can't match
	// Unless the pattern ends with a wildcard
	if len(pathSegments) != len(patternSegments) {
		if len(patternSegments) > 0 && patternSegments[len(patternSegments)-1] == "*" {
			// Wildcard at the end - path should have at least pattern-1 segments
			return len(pathSegments) >= len(patternSegments)-1
		}
		return false
	}

	// Check each segment
	for i, patternSeg := range patternSegments {
		// If pattern segment starts with :, it's a parameter and always matches
		if strings.HasPrefix(patternSeg, ":") {
			continue
		}

		// If pattern segment is *, it matches anything
		if patternSeg == "*" {
			continue
		}

		// Otherwise, segments must match exactly
		if patternSeg != pathSegments[i] {
			return false
		}
	}

	return true
}
