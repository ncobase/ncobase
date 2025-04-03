package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	groupEnt "ncobase/core/space/data/ent/group"
	"ncobase/core/space/structs"
	"ncobase/ncore/data/cache"
	"ncobase/ncore/logger"
	"ncobase/ncore/paging"
	"ncobase/ncore/types"
	"ncobase/ncore/validator"

	"github.com/redis/go-redis/v9"
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
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Group]
}

// NewGroupRepository creates a new group repository.
func NewGroupRepository(d *data.Data) GroupRepositoryInterface {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &groupRepository{ec, rc, cache.NewCache[ent.Group](rc, "ncse_group")}
}

// Create creates a new group.
func (r *groupRepository) Create(ctx context.Context, body *structs.CreateGroupBody) (*ent.Group, error) {

	// create builder.
	builder := r.ec.Group.Create()
	// set values.
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

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Create error: %v", err)
		return nil, err
	}

	return row, nil
}

// Get gets a group by ID.
func (r *groupRepository) Get(ctx context.Context, params *structs.FindGroup) (*ent.Group, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", params.Group)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindGroup(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Get error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Get cache error: %v", err)
	}

	return row, nil
}

// GetByIDs gets groups by IDs.
func (r *groupRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Group, error) {
	// create builder.
	builder := r.ec.Group.Query()
	// set conditions.
	builder.Where(groupEnt.IDIn(ids...))
	// execute the builder.
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetByIDs error: %v", err)
		return nil, err
	}
	return rows, nil
}

// GetBySlug gets a group by slug.
func (r *groupRepository) GetBySlug(ctx context.Context, slug string) (*ent.Group, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindGroup(ctx, &structs.FindGroup{Group: slug})
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetBySlug error: %v", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetBySlug cache error: %v", err)
	}

	return row, nil
}

// Update updates a group (full or partial).
func (r *groupRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Group, error) {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Group: slug})
	if err != nil {
		return nil, err
	}

	builder := group.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(types.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(types.ToPointer(value.(string)))
		case "description":
			builder.SetNillableDescription(types.ToPointer(value.(string)))
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

	// execute the builder.
	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Update error: %v", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", group.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("group:slug:%s", group.Slug))
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Update cache error: %v", err)
	}

	return row, nil
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

// Delete deletes a group.
func (r *groupRepository) Delete(ctx context.Context, slug string) error {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Group: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Group.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(groupEnt.IDEQ(group.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "groupRepo.Delete error: %v", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", group.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("group:slug:%s", group.Slug))
	if err != nil {
		logger.Errorf(ctx, "groupRepo.Delete cache error: %v", err)
	}

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

// GetGroupsByTenantID retrieves all groups under a tenant.
func (r *groupRepository) GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error) {
	groups, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "groupRepo.GetGroupsByTenantID error: %v", err)
		return nil, err
	}
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
