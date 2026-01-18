package repository

import (
	"context"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	spaceOrgEnt "ncobase/core/space/data/ent/spaceorganization"
	"ncobase/core/space/structs"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ncobase/ncore/data/cache"
	"github.com/ncobase/ncore/logging/logger"
)

// SpaceOrganizationRepositoryInterface represents the space group repository interface.
type SpaceOrganizationRepositoryInterface interface {
	Create(ctx context.Context, body *structs.SpaceOrganization) (*ent.SpaceOrganization, error)
	GetBySpaceID(ctx context.Context, id string) ([]*ent.SpaceOrganization, error)
	GetByOrgID(ctx context.Context, id string) ([]*ent.SpaceOrganization, error)
	GetBySpaceIDs(ctx context.Context, ids []string) ([]*ent.SpaceOrganization, error)
	GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.SpaceOrganization, error)
	Delete(ctx context.Context, tid, gid string) error
	DeleteAllBySpaceID(ctx context.Context, id string) error
	DeleteAllByOrgID(ctx context.Context, id string) error
	GetOrgsBySpaceID(ctx context.Context, spaceID string) ([]string, error)
	GetSpacesByOrgID(ctx context.Context, orgID string) ([]string, error)
	IsGroupInSpace(ctx context.Context, spaceID string, orgID string) (bool, error)
}

// spaceGroupRepository implements the SpaceOrganizationRepositoryInterface.
type spaceGroupRepository struct {
	data             *data.Data
	spaceGroupCache  cache.ICache[ent.SpaceOrganization]
	spaceGroupsCache cache.ICache[[]string] // Maps space ID to group IDs
	groupSpacesCache cache.ICache[[]string] // Maps group ID to space IDs
	relationshipTTL  time.Duration
}

// NewSpaceOrganizationRepository creates a new space group repository.
func NewSpaceOrganizationRepository(d *data.Data) SpaceOrganizationRepositoryInterface {
	redisClient := d.GetRedis().(*redis.Client)

	return &spaceGroupRepository{
		data:             d,
		spaceGroupCache:  cache.NewCache[ent.SpaceOrganization](redisClient, "ncse_space:space_orgs"),
		spaceGroupsCache: cache.NewCache[[]string](redisClient, "ncse_space:space_group_mappings"),
		groupSpacesCache: cache.NewCache[[]string](redisClient, "ncse_space:group_space_mappings"),
		relationshipTTL:  time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates space group relationship
func (r *spaceGroupRepository) Create(ctx context.Context, body *structs.SpaceOrganization) (*ent.SpaceOrganization, error) {
	builder := r.data.GetMasterEntClient().SpaceOrganization.Create()

	builder.SetSpaceID(body.SpaceID)
	builder.SetOrgID(body.OrgID)

	row, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.Create error: %v", err)
		return nil, err
	}

	// Cache the relationship and invalidate related caches
	go func() {
		r.cacheSpaceOrganization(context.Background(), row)
		r.invalidateSpaceOrganizationsCache(context.Background(), body.SpaceID)
		r.invalidateGroupSpacesCache(context.Background(), body.OrgID)
	}()

	return row, nil
}

// GetBySpaceID finds orgs by space id
func (r *spaceGroupRepository) GetBySpaceID(ctx context.Context, id string) ([]*ent.SpaceOrganization, error) {
	builder := r.data.GetSlaveEntClient().SpaceOrganization.Query()
	builder.Where(spaceOrgEnt.SpaceIDEQ(id))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetBySpaceID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheSpaceOrganization(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetByOrgID finds spaces by group id
func (r *spaceGroupRepository) GetByOrgID(ctx context.Context, id string) ([]*ent.SpaceOrganization, error) {
	builder := r.data.GetSlaveEntClient().SpaceOrganization.Query()
	builder.Where(spaceOrgEnt.OrgIDEQ(id))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetByOrgID error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheSpaceOrganization(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetBySpaceIDs finds orgs by space ids
func (r *spaceGroupRepository) GetBySpaceIDs(ctx context.Context, ids []string) ([]*ent.SpaceOrganization, error) {
	builder := r.data.GetSlaveEntClient().SpaceOrganization.Query()
	builder.Where(spaceOrgEnt.SpaceIDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetBySpaceIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheSpaceOrganization(context.Background(), tg)
		}
	}()

	return rows, nil
}

// GetByOrgIDs finds spaces by group ids
func (r *spaceGroupRepository) GetByOrgIDs(ctx context.Context, ids []string) ([]*ent.SpaceOrganization, error) {
	builder := r.data.GetSlaveEntClient().SpaceOrganization.Query()
	builder.Where(spaceOrgEnt.OrgIDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetByOrgIDs error: %v", err)
		return nil, err
	}

	// Cache relationships in background
	go func() {
		for _, tg := range rows {
			r.cacheSpaceOrganization(context.Background(), tg)
		}
	}()

	return rows, nil
}

// Delete deletes space group relationship
func (r *spaceGroupRepository) Delete(ctx context.Context, tid, gid string) error {
	if _, err := r.data.GetMasterEntClient().SpaceOrganization.Delete().
		Where(spaceOrgEnt.SpaceIDEQ(tid), spaceOrgEnt.OrgIDEQ(gid)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.Delete error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceOrganizationCache(context.Background(), tid, gid)
		r.invalidateSpaceOrganizationsCache(context.Background(), tid)
		r.invalidateGroupSpacesCache(context.Background(), gid)
	}()

	return nil
}

// DeleteAllBySpaceID deletes all space group by space id
func (r *spaceGroupRepository) DeleteAllBySpaceID(ctx context.Context, id string) error {
	relationships, err := r.GetBySpaceID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceOrganization.Delete().
		Where(spaceOrgEnt.SpaceIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.DeleteAllBySpaceID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateSpaceOrganizationsCache(context.Background(), id)
		for _, tg := range relationships {
			r.invalidateSpaceOrganizationCache(context.Background(), tg.SpaceID, tg.OrgID)
			r.invalidateGroupSpacesCache(context.Background(), tg.OrgID)
		}
	}()

	return nil
}

// DeleteAllByOrgID deletes all space group by group id
func (r *spaceGroupRepository) DeleteAllByOrgID(ctx context.Context, id string) error {
	relationships, err := r.GetByOrgID(ctx, id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get relationships for cache invalidation: %v", err)
	}

	if _, err := r.data.GetMasterEntClient().SpaceOrganization.Delete().
		Where(spaceOrgEnt.OrgIDEQ(id)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.DeleteAllByOrgID error: %v", err)
		return err
	}

	// Invalidate caches
	go func() {
		r.invalidateGroupSpacesCache(context.Background(), id)
		for _, tg := range relationships {
			r.invalidateSpaceOrganizationCache(context.Background(), tg.SpaceID, tg.OrgID)
			r.invalidateSpaceOrganizationsCache(context.Background(), tg.SpaceID)
		}
	}()

	return nil
}

// GetOrgsBySpaceID retrieves all orgs under a space
func (r *spaceGroupRepository) GetOrgsBySpaceID(ctx context.Context, spaceID string) ([]string, error) {
	cacheKey := fmt.Sprintf("space_orgs:%s", spaceID)
	var orgIDs []string
	if err := r.spaceGroupsCache.GetArray(ctx, cacheKey, &orgIDs); err == nil && len(orgIDs) > 0 {
		return orgIDs, nil
	}

	// Fallback to database
	spaceGroups, err := r.data.GetSlaveEntClient().SpaceOrganization.Query().
		Where(spaceOrgEnt.SpaceIDEQ(spaceID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetOrgsBySpaceID error: %v", err)
		return nil, err
	}

	orgIDs = make([]string, len(spaceGroups))
	for i, tg := range spaceGroups {
		orgIDs[i] = tg.OrgID
	}

	// Cache for future use
	go func() {
		if err := r.spaceGroupsCache.SetArray(context.Background(), cacheKey, orgIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache space orgs %s: %v", spaceID, err)
		}
	}()

	return orgIDs, nil
}

// GetSpacesByOrgID retrieves all spaces that have a organization
func (r *spaceGroupRepository) GetSpacesByOrgID(ctx context.Context, orgID string) ([]string, error) {
	cacheKey := fmt.Sprintf("group_spaces:%s", orgID)
	var spaceIDs []string
	if err := r.groupSpacesCache.GetArray(ctx, cacheKey, &spaceIDs); err == nil && len(spaceIDs) > 0 {
		return spaceIDs, nil
	}

	// Fallback to database
	spaceGroups, err := r.data.GetSlaveEntClient().SpaceOrganization.Query().
		Where(spaceOrgEnt.OrgIDEQ(orgID)).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.GetSpacesByOrgID error: %v", err)
		return nil, err
	}

	spaceIDs = make([]string, len(spaceGroups))
	for i, tg := range spaceGroups {
		spaceIDs[i] = tg.SpaceID
	}

	// Cache for future use
	go func() {
		if err := r.groupSpacesCache.SetArray(context.Background(), cacheKey, spaceIDs, r.relationshipTTL); err != nil {
			logger.Debugf(context.Background(), "Failed to cache group spaces %s: %v", orgID, err)
		}
	}()

	return spaceIDs, nil
}

// IsGroupInSpace verifies if a organization belongs to a specific space
func (r *spaceGroupRepository) IsGroupInSpace(ctx context.Context, spaceID string, orgID string) (bool, error) {
	cacheKey := fmt.Sprintf("relationship:%s:%s", spaceID, orgID)
	if cached, err := r.spaceGroupCache.Get(ctx, cacheKey); err == nil && cached != nil {
		return true, nil
	}

	count, err := r.data.GetSlaveEntClient().SpaceOrganization.Query().
		Where(spaceOrgEnt.SpaceIDEQ(spaceID), spaceOrgEnt.OrgIDEQ(orgID)).Count(ctx)
	if err != nil {
		logger.Errorf(ctx, "spaceGroupRepo.IsGroupInSpace error: %v", err)
		return false, err
	}

	exists := count > 0

	if exists {
		go func() {
			// Create dummy relationship for caching
			relationship := &ent.SpaceOrganization{
				SpaceID: spaceID,
				OrgID:   orgID,
			}
			r.cacheSpaceOrganization(context.Background(), relationship)
		}()
	}

	return exists, nil
}

// Cache management methods
func (r *spaceGroupRepository) cacheSpaceOrganization(ctx context.Context, tg *ent.SpaceOrganization) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", tg.SpaceID, tg.OrgID)
	if err := r.spaceGroupCache.Set(ctx, relationshipKey, tg, r.relationshipTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache space group relationship %s:%s: %v", tg.SpaceID, tg.OrgID, err)
	}
}

func (r *spaceGroupRepository) invalidateSpaceOrganizationCache(ctx context.Context, spaceID, orgID string) {
	relationshipKey := fmt.Sprintf("relationship:%s:%s", spaceID, orgID)
	if err := r.spaceGroupCache.Delete(ctx, relationshipKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space group relationship cache %s:%s: %v", spaceID, orgID, err)
	}
}

func (r *spaceGroupRepository) invalidateSpaceOrganizationsCache(ctx context.Context, spaceID string) {
	cacheKey := fmt.Sprintf("space_orgs:%s", spaceID)
	if err := r.spaceGroupsCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate space orgs cache %s: %v", spaceID, err)
	}
}

func (r *spaceGroupRepository) invalidateGroupSpacesCache(ctx context.Context, orgID string) {
	cacheKey := fmt.Sprintf("group_spaces:%s", orgID)
	if err := r.groupSpacesCache.Delete(ctx, cacheKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate group spaces cache %s: %v", orgID, err)
	}
}
