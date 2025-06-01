package repository

import (
	"context"
	"fmt"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	groupEnt "ncobase/space/data/ent/group"
	"ncobase/space/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/validation/validator"
)

// GroupRepositoryInterface represents the group repository interface.
type GroupRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateGroupBody) (*ent.Group, error)
	Get(ctx context.Context, params *structs.FindGroup) (*ent.Group, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.Group, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Group, error)
	GetTree(ctx context.Context, params *structs.FindGroup) ([]*ent.Group, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Group, error)
	List(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, error)
	ListWithCount(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, int, error)
	Delete(ctx context.Context, slug string) error
	FindGroup(ctx context.Context, params *structs.FindGroup) (*ent.Group, error)
	ListBuilder(ctx context.Context, params *structs.ListGroupParams) (*ent.GroupQuery, error)
	CountX(ctx context.Context, params *structs.ListGroupParams) int
	GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error)
	IsGroupInTenant(ctx context.Context, groupID string, tenantID string) (bool, error)
}

// groupRepository implements the GroupRepositoryInterface.
type groupRepository struct {
	ec                *ent.Client
	groupCache        cache.ICache[ent.Group]
	slugMappingCache  cache.ICache[string]   // Maps slug to group ID
	tenantGroupsCache cache.ICache[[]string] // Maps tenant ID to group IDs
	groupTTL          time.Duration
}

// NewGroupRepository creates a new group repository.
func NewGroupRepository(d *data.Data) GroupRepositoryInterface {
	redisClient := d.GetRedis()

	return &groupRepository{
		ec:                d.GetMasterEntClient(),
		groupCache:        cache.NewCache[ent.Group](redisClient, "ncse_space:groups"),
		slugMappingCache:  cache.NewCache[string](redisClient, "ncse_space:group_mappings"),
		tenantGroupsCache: cache.NewCache[[]string](redisClient, "ncse_space:tenant_groups"),
		groupTTL:          time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new group
func (r *groupRepository) Create(ctx context.Context, body *structs.CreateGroupBody) (*ent.Group, error) {
	builder := r.ec.Group.Create()
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableParentID(body.ParentID)
	builder.SetNillableTenantID(body.TenantID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Leader) && !validator.IsEmpty(body.Leader) {
		builder.SetLeader(*body.Leader)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	group, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the group and invalidate tenant groups cache
	go func() {
		r.cacheGroup(context.Background(), group)
		if body.TenantID != nil {
			r.invalidateTenantGroupsCache(context.Background(), *body.TenantID)
		}
	}()

	return group, nil
}

// Get gets a group by ID or slug
func (r *groupRepository) Get(ctx context.Context, params *structs.FindGroup) (*ent.Group, error) {
	// Try to get group ID from slug mapping cache if searching by slug
	if params.Group != "" {
		if groupID, err := r.getGroupIDBySlug(ctx, params.Group); err == nil && groupID != "" {
			// Try to get from group cache
			cacheKey := fmt.Sprintf("id:%s", groupID)
			if cached, err := r.groupCache.Get(ctx, cacheKey); err == nil && cached != nil {
				return cached, nil
			}
		}
	}

	// Fallback to database
	row, err := r.FindGroup(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Get error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheGroup(context.Background(), row)

	return row, nil
}

// GetByIDs gets groups by IDs with batch caching
func (r *groupRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Group, error) {
	// Try to get from cache first
	cacheKeys := make([]string, len(ids))
	for i, id := range ids {
		cacheKeys[i] = fmt.Sprintf("id:%s", id)
	}

	cachedGroups, err := r.groupCache.GetMultiple(ctx, cacheKeys)
	if err == nil && len(cachedGroups) == len(ids) {
		// All groups found in cache
		groups := make([]*ent.Group, len(ids))
		for i, key := range cacheKeys {
			if group, exists := cachedGroups[key]; exists {
				groups[i] = group
			}
		}
		return groups, nil
	}

	// Fallback to database
	builder := r.ec.Group.Query()
	builder.Where(groupEnt.IDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetByIDs error: %v", err)
		return nil, err
	}

	// Cache groups in background
	go func() {
		for _, group := range rows {
			r.cacheGroup(context.Background(), group)
		}
	}()

	return rows, nil
}

// GetBySlug gets a group by slug
func (r *groupRepository) GetBySlug(ctx context.Context, slug string) (*ent.Group, error) {
	return r.Get(ctx, &structs.FindGroup{Group: slug})
}

// Update updates a group
func (r *groupRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Group, error) {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Group: slug})
	if err != nil {
		return nil, err
	}

	builder := group.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(convert.ToPointer(value.(string)))
		case "leader":
			builder.SetLeader(value.(types.JSON))
		case "disabled":
			builder.SetDisabled(value.(bool))
		case "extra_props":
			builder.SetExtras(value.(types.JSON))
		case "parent_id":
			builder.SetParentID(value.(string))
		case "tenant_id":
			builder.SetTenantID(value.(string))
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	updatedGroup, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Update error: %v", err)
		return nil, err
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateGroupCache(context.Background(), group)
		r.cacheGroup(context.Background(), updatedGroup)

		// If tenant changed, invalidate both old and new tenant caches
		if group.TenantID != "" {
			r.invalidateTenantGroupsCache(context.Background(), group.TenantID)
		}
		if updatedGroup.TenantID != "" && (group.TenantID == "" || group.TenantID != updatedGroup.TenantID) {
			r.invalidateTenantGroupsCache(context.Background(), updatedGroup.TenantID)
		}
	}()

	return updatedGroup, nil
}

// List gets a list of groups.
func (r *groupRepository) List(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithCount gets a list and counts of groups.
func (r *groupRepository) ListWithCount(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, int, error) {
	builder, err := r.ListBuilder(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	builder = applySorting(builder, params.SortBy)

	// Apply cursor-based pagination
	if params.Cursor != "" {
		id, timestamp, err := paging.DecodeCursor(params.Cursor)
		if err != nil {
			return nil, 0, err
		}
		builder = applyCursorCondition(builder, id, timestamp, params.Direction, params.SortBy)
	}

	// Execute count query
	total, err := builder.Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	// Execute main query
	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.ListWithCount error: %v", err)
		return nil, 0, err
	}

	return rows, total, nil
}

func applySorting(builder *ent.GroupQuery, sortBy string) *ent.GroupQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		return builder.Order(ent.Desc(groupEnt.FieldCreatedAt))
	default:
		return builder.Order(ent.Desc(groupEnt.FieldCreatedAt))
	}
}

func applyCursorCondition(builder *ent.GroupQuery, id string, value any, direction string, sortBy string) *ent.GroupQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				groupEnt.Or(
					groupEnt.CreatedAtGT(timestamp),
					groupEnt.And(
						groupEnt.CreatedAtEQ(timestamp),
						groupEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			groupEnt.Or(
				groupEnt.CreatedAtLT(timestamp),
				groupEnt.And(
					groupEnt.CreatedAtEQ(timestamp),
					groupEnt.IDLT(id),
				),
			),
		)
	default:
		return applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// Delete deletes a group
func (r *groupRepository) Delete(ctx context.Context, slug string) error {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Group: slug})
	if err != nil {
		return err
	}

	builder := r.ec.Group.Delete()
	if _, err = builder.Where(groupEnt.IDEQ(group.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "groupRepo.Delete error: %v", err)
		return err
	}

	// Invalidate cache
	go func() {
		r.invalidateGroupCache(context.Background(), group)
		if group.TenantID != "" {
			r.invalidateTenantGroupsCache(context.Background(), group.TenantID)
		}
	}()

	return nil
}

// FindGroup finds a group.
func (r *groupRepository) FindGroup(ctx context.Context, params *structs.FindGroup) (*ent.Group, error) {

	// create builder.
	builder := r.ec.Group.Query()

	// support slug or ID
	if validator.IsNotEmpty(params.Group) {
		builder = builder.Where(groupEnt.Or(
			groupEnt.ID(params.Group),
			groupEnt.SlugEQ(params.Group),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// GetTree retrieves the group tree.
func (r *groupRepository) GetTree(ctx context.Context, params *structs.FindGroup) ([]*ent.Group, error) {
	// create builder
	builder := r.ec.Group.Query()

	// set where conditions
	if validator.IsNotEmpty(params.Tenant) {
		builder.Where(groupEnt.TenantIDEQ(params.Tenant))
	}

	// handle sub groups
	if validator.IsNotEmpty(params.Group) && params.Group != "root" {
		return r.getSubGroup(ctx, params.Group, builder)
	}

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// ListBuilder - create list builder.
func (r *groupRepository) ListBuilder(_ context.Context, params *structs.ListGroupParams) (*ent.GroupQuery, error) {
	// create builder.
	builder := r.ec.Group.Query()

	// match tenant id.
	if params.Tenant != "" {
		builder.Where(groupEnt.TenantIDEQ(params.Tenant))
	}

	// match parent id.
	if params.Parent == "" {
		builder.Where(groupEnt.Or(
			groupEnt.ParentIDIsNil(),
			groupEnt.ParentIDEQ(""),
			groupEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(groupEnt.ParentIDEQ(params.Parent))
	}

	return builder, nil
}

// CountX gets a count of groups.
func (r *groupRepository) CountX(ctx context.Context, params *structs.ListGroupParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// GetGroupsByTenantID retrieves all groups under a tenant
func (r *groupRepository) GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error) {
	// Try to get group IDs from tenant cache
	cacheKey := fmt.Sprintf("tenant:%s", tenantID)
	var groupIDs []string
	if err := r.tenantGroupsCache.GetArray(ctx, cacheKey, &groupIDs); err == nil && len(groupIDs) > 0 {
		return r.GetByIDs(ctx, groupIDs)
	}

	// Fallback to database
	groups, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetGroupsByTenantID error: %v", err)
		return nil, err
	}

	// Cache groups and tenant mapping
	go func() {
		groupIDs := make([]string, len(groups))
		for i, group := range groups {
			r.cacheGroup(context.Background(), group)
			groupIDs[i] = group.ID
		}

		// Cache tenant to groups mapping
		if err := r.tenantGroupsCache.SetArray(context.Background(), cacheKey, groupIDs, r.groupTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache tenant groups mapping %s: %v", tenantID, err)
		}
	}()

	return groups, nil
}

// IsGroupInTenant verifies if a group belongs to a specific tenant.
func (r *groupRepository) IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error) {
	count, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID), groupEnt.IDEQ(groupID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.IsGroupInTenant error: %v", err)
		return false, err
	}
	return count > 0, nil
}

// getSubGroup - get sub groups.
func (r *groupRepository) getSubGroup(ctx context.Context, rootID string, builder *ent.GroupQuery) ([]*ent.Group, error) {
	// set where conditions
	builder.Where(
		groupEnt.Or(
			groupEnt.ID(rootID),
			groupEnt.ParentIDHasPrefix(rootID),
		),
	)

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// executeArrayQuery - execute the builder query and return results.
func (r *groupRepository) executeArrayQuery(ctx context.Context, builder *ent.GroupQuery) ([]*ent.Group, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}

func (r *groupRepository) cacheGroup(ctx context.Context, group *ent.Group) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", group.ID)
	if err := r.groupCache.Set(ctx, idKey, group, r.groupTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache group by ID %s: %v", group.ID, err)
	}

	// Cache slug to ID mapping
	if group.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", group.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &group.ID, r.groupTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", group.Slug, err)
		}
	}
}

func (r *groupRepository) invalidateGroupCache(ctx context.Context, group *ent.Group) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", group.ID)
	if err := r.groupCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group ID cache %s: %v", group.ID, err)
	}

	// Invalidate slug mapping
	if group.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", group.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", group.Slug, err)
		}
	}
}

func (r *groupRepository) invalidateTenantGroupsCache(ctx context.Context, tenantID string) {
	cacheKey := fmt.Sprintf("tenant:%s", tenantID)
	if err := r.tenantGroupsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate tenant groups cache %s: %v", tenantID, err)
	}
}

func (r *groupRepository) getGroupIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	groupID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || groupID == nil {
		return "", err
	}
	return *groupID, nil
}
