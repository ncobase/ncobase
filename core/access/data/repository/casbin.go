package repository

import (
	"context"
	"fmt"
	"ncobase/core/access/data"
	"ncobase/core/access/data/ent"
	casbinRuleEnt "ncobase/core/access/data/ent/casbinrule"
	"ncobase/core/access/structs"
	"time"
	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
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
	data            *data.Data
	ruleCache       cache.ICache[ent.CasbinRule]
	pTypeRulesCache cache.ICache[[]string] // Maps ptype to rule IDs
	ruleTTL         time.Duration
}

// NewCasbinRule creates a new Casbin rule repository.
func NewCasbinRule(d *data.Data) CasbinRuleRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &casbinRuleRepository{
		data:            d,
		ruleCache:       cache.NewCache[ent.CasbinRule](redisClient, "ncse_access:casbin_rules"),
		pTypeRulesCache: cache.NewCache[[]string](redisClient, "ncse_access:ptype_rules"),
		ruleTTL:         time.Hour * 1, // 1 hour cache TTL
	}
}

// Create creates a new Casbin rule.
func (r *casbinRuleRepository) Create(ctx context.Context, body *structs.CasbinRuleBody) (*ent.CasbinRule, error) {
	// Use master for writes
	builder := r.data.GetMasterEntClient().CasbinRule.Create()

	// Set values
	builder.SetNillablePType(&body.PType)
	builder.SetNillableV0(&body.V0)
	builder.SetNillableV1(&body.V1)
	builder.SetNillableV2(&body.V2)
	builder.SetNillableV3(body.V3)
	builder.SetNillableV4(body.V4)
	builder.SetNillableV5(body.V5)
	builder.SetNillableCreatedBy(body.CreatedBy)

	// Execute the builder
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the rule and invalidate ptype cache
	go func() {
		r.cacheRule(context.Background(), row)
		if body.PType != "" {
			r.invalidatePTypeRulesCache(context.Background(), body.PType)
		}
	}()

	return row, nil
}

// GetByID gets a Casbin rule by ID.
func (r *casbinRuleRepository) GetByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("id:%s", id)
	if cached, err := r.ruleCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// Fallback to database using slave
	row, err := r.FindByID(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "casbinRuleRepo.GetByID error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheRule(context.Background(), row)

	return row, nil
}

// Update updates a Casbin rule (full or partial).
func (r *casbinRuleRepository) Update(ctx context.Context, id string, updates types.JSON) (*ent.CasbinRule, error) {
	// Get original rule for cache invalidation
	originalRule, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use master for writes
	builder := originalRule.Update()

	// Update the Casbin rule fields based on updates map
	for field, value := range updates {
		switch field {
		case "p_type":
			builder = builder.SetNillablePType(convert.ToPointer(value.(string)))
		case "v0":
			builder = builder.SetNillableV0(convert.ToPointer(value.(string)))
		case "v1":
			builder = builder.SetNillableV1(convert.ToPointer(value.(string)))
		case "v2":
			builder = builder.SetNillableV2(convert.ToPointer(value.(string)))
		case "v3":
			builder = builder.SetNillableV3(convert.ToPointer(value.(string)))
		case "v4":
			builder = builder.SetNillableV4(convert.ToPointer(value.(string)))
		case "v5":
			builder = builder.SetNillableV5(convert.ToPointer(value.(string)))
		case "updated_by":
			builder = builder.SetNillableUpdatedBy(convert.ToPointer(value.(string)))
		}
	}

	// Save the updated Casbin rule
	updatedRule, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateRuleCache(context.Background(), originalRule)
		r.cacheRule(context.Background(), updatedRule)

		// Invalidate ptype caches for both old and new ptypes
		if originalRule.PType != "" {
			r.invalidatePTypeRulesCache(context.Background(), originalRule.PType)
		}
		if updatedRule.PType != "" && (originalRule.PType == "" || originalRule.PType != updatedRule.PType) {
			r.invalidatePTypeRulesCache(context.Background(), updatedRule.PType)
		}
	}()

	return updatedRule, nil
}

// Delete deletes a Casbin rule by ID.
func (r *casbinRuleRepository) Delete(ctx context.Context, id string) error {
	// Get rule first for cache invalidation
	rule, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Use master for writes
	builder := r.data.GetMasterEntClient().CasbinRule.Delete()

	// Execute the builder and verify the result
	if _, err = builder.Where(casbinRuleEnt.IDEQ(rule.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "casbinRuleRepo.Delete error: %v", err)
		return err
	}

	// Invalidate cache
	go func() {
		r.invalidateRuleCache(context.Background(), rule)
		if rule.PType != "" {
			r.invalidatePTypeRulesCache(context.Background(), rule.PType)
		}
	}()

	return nil
}

// FindByID finds a Casbin rule by ID.
func (r *casbinRuleRepository) FindByID(ctx context.Context, id string) (*ent.CasbinRule, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().CasbinRule.Query()

	// Add conditions to the query
	builder = builder.Where(casbinRuleEnt.IDEQ(id))

	// Execute the builder
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// Find finds Casbin rules based on query parameters.
func (r *casbinRuleRepository) Find(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// Create list builder using slave
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// Execute the query
	rows, err := builder.All(ctx)
	if err != nil {
		return nil, err
	}

	// Cache rules in background
	go func() {
		for _, rule := range rows {
			r.cacheRule(context.Background(), rule)
		}
	}()

	return rows, nil
}

// List gets a list of Casbin rules.
func (r *casbinRuleRepository) List(ctx context.Context, params *structs.ListCasbinRuleParams) ([]*ent.CasbinRule, error) {
	// Create list builder using slave
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

	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "casbinRuleRepo.List error: %v", err)
		return nil, err
	}

	// Cache rules in background
	go func() {
		for _, rule := range rows {
			r.cacheRule(context.Background(), rule)
		}
	}()

	return rows, nil
}

// CountX gets a count of Casbin rules.
func (r *casbinRuleRepository) CountX(ctx context.Context, params *structs.ListCasbinRuleParams) int {
	// Create list builder using slave
	builder, err := r.listBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// listBuilder builds the list query.
func (r *casbinRuleRepository) listBuilder(_ context.Context, params *structs.ListCasbinRuleParams) (*ent.CasbinRuleQuery, error) {
	// Use slave for reads
	builder := r.data.GetSlaveEntClient().CasbinRule.Query()

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

// cacheRule caches a Casbin rule
func (r *casbinRuleRepository) cacheRule(ctx context.Context, rule *ent.CasbinRule) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", rule.ID)
	if err := r.ruleCache.Set(ctx, idKey, rule, r.ruleTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache Casbin rule %s: %v", rule.ID, err)
	}
}

// invalidateRuleCache invalidates a Casbin rule
func (r *casbinRuleRepository) invalidateRuleCache(ctx context.Context, rule *ent.CasbinRule) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", rule.ID)
	if err := r.ruleCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate Casbin rule cache %s: %v", rule.ID, err)
	}
}

// invalidatePTypeRulesCache invalidates a Casbin rule
func (r *casbinRuleRepository) invalidatePTypeRulesCache(ctx context.Context, ptype string) {
	cacheKey := fmt.Sprintf("ptype:%s", ptype)
	if err := r.pTypeRulesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate ptype rules cache %s: %v", ptype, err)
	}
}
