package repo

import (
	"context"
	"fmt"
	"ncobase/internal/data"
	"ncobase/internal/data/ent"
	groupEnt "ncobase/internal/data/ent/group"
	"ncobase/internal/data/structs"

	"ncobase/common/cache"
	"ncobase/common/log"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/redis/go-redis/v9"
)

// Group represents the group repository interface.
type Group interface {
	Create(ctx context.Context, body *structs.CreateGroupBody) (*ent.Group, error)
	GetByID(ctx context.Context, id string) (*ent.Group, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Group, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Group, error)
	List(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, error)
	Delete(ctx context.Context, slug string) error
	FindGroup(ctx context.Context, params *structs.FindGroup) (*ent.Group, error)
	ListBuilder(ctx context.Context, params *structs.ListGroupParams) (*ent.GroupQuery, error)
	CountX(ctx context.Context, params *structs.ListGroupParams) int
	GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error)
	IsGroupInTenant(ctx context.Context, groupID string, tenantID string) (bool, error)
}

// groupRepo implements the Group interface.
type groupRepo struct {
	ec *ent.Client
	rc *redis.Client
	c  *cache.Cache[ent.Group]
}

// NewGroup creates a new group repository.
func NewGroup(d *data.Data) Group {
	ec := d.GetEntClient()
	rc := d.GetRedis()
	return &groupRepo{ec, rc, cache.NewCache[ent.Group](rc, cache.Key("nb_group"))}
}

// Create creates a new group.
func (r *groupRepo) Create(ctx context.Context, body *structs.CreateGroupBody) (*ent.Group, error) {

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
		log.Errorf(context.Background(), "groupRepo.Create error: %v\n", err)
		return nil, err
	}

	return row, nil
}

// GetByID gets a group by ID.
func (r *groupRepo) GetByID(ctx context.Context, id string) (*ent.Group, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", id)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindGroup(ctx, &structs.FindGroup{ID: id})
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.GetByID error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.GetByID cache error: %v\n", err)
	}

	return row, nil
}

// GetBySlug gets a group by slug.
func (r *groupRepo) GetBySlug(ctx context.Context, slug string) (*ent.Group, error) {
	// check cache
	cacheKey := fmt.Sprintf("%s", slug)
	if cached, err := r.c.Get(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	// If not found in cache, query the database
	row, err := r.FindGroup(ctx, &structs.FindGroup{Slug: slug})
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.GetBySlug error: %v\n", err)
		return nil, err
	}

	// cache the result
	err = r.c.Set(ctx, cacheKey, row)
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.GetBySlug cache error: %v\n", err)
	}

	return row, nil
}

// Update updates a group (full or partial).
func (r *groupRepo) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Group, error) {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Slug: slug})
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
		log.Errorf(context.Background(), "groupRepo.Update error: %v\n", err)
		return nil, err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", group.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("group:slug:%s", group.Slug))
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.Update cache error: %v\n", err)
	}

	return row, nil
}

// List gets a list of groups.
func (r *groupRepo) List(ctx context.Context, params *structs.ListGroupParams) ([]*ent.Group, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(int(params.Limit))

	rows, err := builder.All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.List error: %v\n", err)
		return nil, err
	}

	return rows, nil
}

// Delete deletes a group.
func (r *groupRepo) Delete(ctx context.Context, slug string) error {
	group, err := r.FindGroup(ctx, &structs.FindGroup{Slug: slug})
	if err != nil {
		return err
	}

	// create builder.
	builder := r.ec.Group.Delete()

	// execute the builder and verify the result.
	if _, err = builder.Where(groupEnt.IDEQ(group.ID)).Exec(ctx); err != nil {
		log.Errorf(context.Background(), "groupRepo.Delete error: %v\n", err)
		return err
	}

	// remove from cache
	cacheKey := fmt.Sprintf("%s", group.ID)
	err = r.c.Delete(ctx, cacheKey)
	err = r.c.Delete(ctx, fmt.Sprintf("group:slug:%s", group.Slug))
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.Delete cache error: %v\n", err)
	}

	return nil
}

// FindGroup finds a group.
func (r *groupRepo) FindGroup(ctx context.Context, params *structs.FindGroup) (*ent.Group, error) {

	// create builder.
	builder := r.ec.Group.Query()

	if validator.IsNotEmpty(params.ID) {
		builder = builder.Where(groupEnt.IDEQ(params.ID))
	}
	// support slug or ID
	if validator.IsNotEmpty(params.Slug) {
		builder = builder.Where(groupEnt.Or(
			groupEnt.ID(params.Slug),
			groupEnt.SlugEQ(params.Slug),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// ListBuilder creates list builder.
func (r *groupRepo) ListBuilder(ctx context.Context, params *structs.ListGroupParams) (*ent.GroupQuery, error) {
	// Here you can construct and return a builder for listing groups based on the provided parameters.
	// Similar to the ListBuilder method in the topicRepo.
	return nil, nil
}

// CountX gets a count of groups.
func (r *groupRepo) CountX(ctx context.Context, params *structs.ListGroupParams) int {
	// Here you can implement the logic to count the number of groups based on the provided parameters.
	return 0
}

// GetGroupsByTenantID retrieves all groups under a tenant.
func (r *groupRepo) GetGroupsByTenantID(ctx context.Context, tenantID string) ([]*ent.Group, error) {
	groups, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID)).All(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.GetGroupsByTenantID error: %v\n", err)
		return nil, err
	}
	return groups, nil
}

// IsGroupInTenant verifies if a group belongs to a specific tenant.
func (r *groupRepo) IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error) {
	count, err := r.ec.Group.Query().Where(groupEnt.TenantIDEQ(tenantID), groupEnt.IDEQ(groupID)).Count(ctx)
	if err != nil {
		log.Errorf(context.Background(), "groupRepo.IsGroupInTenant error: %v\n", err)
		return false, err
	}
	return count > 0, nil
}
