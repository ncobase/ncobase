package repository

import (
	"context"
	"fmt"
	"ncobase/organization/data"
	"ncobase/organization/data/ent"
	organizationEnt "ncobase/organization/data/ent/organization"
	"ncobase/organization/structs"
	"time"

	"github.com/ncobase/ncore/data/databases/cache"
	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/convert"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/ncobase/ncore/data/search"
)

// OrganizationRepositoryInterface represents the organization repository interface.
type OrganizationRepositoryInterface interface {
	Create(ctx context.Context, body *structs.CreateOrganizationBody) (*ent.Organization, error)
	Get(ctx context.Context, params *structs.FindOrganization) (*ent.Organization, error)
	GetByIDs(ctx context.Context, ids []string) ([]*ent.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*ent.Organization, error)
	GetTree(ctx context.Context, params *structs.FindOrganization) ([]*ent.Organization, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*ent.Organization, error)
	List(ctx context.Context, params *structs.ListOrganizationParams) ([]*ent.Organization, error)
	ListWithCount(ctx context.Context, params *structs.ListOrganizationParams) ([]*ent.Organization, int, error)
	Delete(ctx context.Context, slug string) error
	FindOrganization(ctx context.Context, params *structs.FindOrganization) (*ent.Organization, error)
	ListBuilder(ctx context.Context, params *structs.ListOrganizationParams) (*ent.OrganizationQuery, error)
	CountX(ctx context.Context, params *structs.ListOrganizationParams) int
}

// organizationRepository implements the OrganizationRepositoryInterface.
type organizationRepository struct {
	data              *data.Data
	ec                *ent.Client
	organizationCache cache.ICache[ent.Organization]
	slugMappingCache  cache.ICache[string] // Maps slug to organization ID
	organizationTTL   time.Duration
}

// NewOrganizationRepository creates a new organization repository.
func NewOrganizationRepository(d *data.Data) OrganizationRepositoryInterface {
	redisClient := d.GetRedis()

	return &organizationRepository{
		data:              d,
		ec:                d.GetMasterEntClient(),
		organizationCache: cache.NewCache[ent.Organization](redisClient, "ncse_organization:organizations"),
		slugMappingCache:  cache.NewCache[string](redisClient, "ncse_organization:organization_mappings"),
		organizationTTL:   time.Hour * 2, // 2 hours cache TTL
	}
}

// Create creates a new organization
func (r *organizationRepository) Create(ctx context.Context, body *structs.CreateOrganizationBody) (*ent.Organization, error) {
	builder := r.ec.Organization.Create()
	builder.SetNillableName(&body.Name)
	builder.SetNillableSlug(&body.Slug)
	builder.SetNillableType(&body.Type)
	builder.SetNillableDescription(&body.Description)
	builder.SetDisabled(body.Disabled)
	builder.SetNillableParentID(body.ParentID)
	builder.SetNillableCreatedBy(body.CreatedBy)

	if !validator.IsNil(body.Leader) && !validator.IsEmpty(body.Leader) {
		builder.SetLeader(*body.Leader)
	}

	if !validator.IsNil(body.Extras) && !validator.IsEmpty(body.Extras) {
		builder.SetExtras(*body.Extras)
	}

	organization, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.Create error: %v", err)
		return nil, err
	}

	// Index in Meilisearch
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "organizations", Document: organization}); err != nil {
		logger.Errorf(ctx, "organizationRepo.Create error creating Meilisearch index: %v", err)
	}

	// Cache the organization
	go r.cacheOrganization(context.Background(), organization)

	return organization, nil
}

// Get gets an organization by ID or slug
func (r *organizationRepository) Get(ctx context.Context, params *structs.FindOrganization) (*ent.Organization, error) {
	// Try to get organization ID from slug mapping cache if searching by slug
	if params.Organization != "" {
		if organizationID, err := r.getOrgIDBySlug(ctx, params.Organization); err == nil && organizationID != "" {
			// Try to get from organization cache
			cacheKey := fmt.Sprintf("id:%s", organizationID)
			if cached, err := r.organizationCache.Get(ctx, cacheKey); err == nil && cached != nil {
				return cached, nil
			}
		}
	}

	// Fallback to database
	row, err := r.FindOrganization(ctx, params)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.Get error: %v", err)
		return nil, err
	}

	// Cache for future use
	go r.cacheOrganization(context.Background(), row)

	return row, nil
}

// GetByIDs gets organizations by IDs with batch caching
func (r *organizationRepository) GetByIDs(ctx context.Context, ids []string) ([]*ent.Organization, error) {
	// Try to get from cache first
	cacheKeys := make([]string, len(ids))
	for i, id := range ids {
		cacheKeys[i] = fmt.Sprintf("id:%s", id)
	}

	cachedOrganizations, err := r.organizationCache.GetMultiple(ctx, cacheKeys)
	if err == nil && len(cachedOrganizations) == len(ids) {
		// All organizations found in cache
		organizations := make([]*ent.Organization, len(ids))
		for i, key := range cacheKeys {
			if organization, exists := cachedOrganizations[key]; exists {
				organizations[i] = organization
			}
		}
		return organizations, nil
	}

	// Fallback to database
	builder := r.ec.Organization.Query()
	builder.Where(organizationEnt.IDIn(ids...))

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.GetByIDs error: %v", err)
		return nil, err
	}

	// Cache organizations in background
	go func() {
		for _, organization := range rows {
			r.cacheOrganization(context.Background(), organization)
		}
	}()

	return rows, nil
}

// GetBySlug gets an organization by slug
func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*ent.Organization, error) {
	return r.Get(ctx, &structs.FindOrganization{Organization: slug})
}

// Update updates an organization
func (r *organizationRepository) Update(ctx context.Context, slug string, updates types.JSON) (*ent.Organization, error) {
	organization, err := r.FindOrganization(ctx, &structs.FindOrganization{Organization: slug})
	if err != nil {
		return nil, err
	}

	builder := organization.Update()

	for field, value := range updates {
		switch field {
		case "name":
			builder.SetNillableName(convert.ToPointer(value.(string)))
		case "slug":
			builder.SetNillableSlug(convert.ToPointer(value.(string)))
		case "type":
			builder.SetNillableType(convert.ToPointer(value.(string)))
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
		case "updated_by":
			builder.SetUpdatedBy(value.(string))
		}
	}

	updatedOrganization, err := builder.Save(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.Update error: %v", err)
		return nil, err
	}

	// Update Meilisearch index
	if err = r.data.IndexDocument(ctx, &search.IndexRequest{Index: "organizations", Document: updatedOrganization, DocumentID: updatedOrganization.ID}); err != nil {
		logger.Errorf(ctx, "organizationRepo.Update error updating Meilisearch index: %v", err)
	}

	// Invalidate and re-cache
	go func() {
		r.invalidateOrganizationCache(context.Background(), organization)
		r.cacheOrganization(context.Background(), updatedOrganization)
	}()

	return updatedOrganization, nil
}

// List gets a list of organizations.
func (r *organizationRepository) List(ctx context.Context, params *structs.ListOrganizationParams) ([]*ent.Organization, error) {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return nil, err
	}

	// limit the result
	builder.Limit(params.Limit)

	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.List error: %v", err)
		return nil, err
	}

	return rows, nil
}

// ListWithCount gets a list and counts of organizations.
func (r *organizationRepository) ListWithCount(ctx context.Context, params *structs.ListOrganizationParams) ([]*ent.Organization, int, error) {
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
		logger.Errorf(ctx, "organizationRepo.ListWithCount count error: %v", err)
		return nil, 0, err
	}

	// Execute main query
	rows, err := builder.Limit(params.Limit).All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.ListWithCount error: %v", err)
		return nil, 0, err
	}

	return rows, total, nil
}

func applySorting(builder *ent.OrganizationQuery, sortBy string) *ent.OrganizationQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		return builder.Order(ent.Desc(organizationEnt.FieldCreatedAt))
	default:
		return builder.Order(ent.Desc(organizationEnt.FieldCreatedAt))
	}
}

func applyCursorCondition(builder *ent.OrganizationQuery, id string, value any, direction string, sortBy string) *ent.OrganizationQuery {
	switch sortBy {
	case structs.SortByCreatedAt:
		timestamp, ok := value.(int64)
		if !ok {
			logger.Errorf(context.Background(), "Invalid timestamp value for cursor")
			return builder
		}
		if direction == "backward" {
			return builder.Where(
				organizationEnt.Or(
					organizationEnt.CreatedAtGT(timestamp),
					organizationEnt.And(
						organizationEnt.CreatedAtEQ(timestamp),
						organizationEnt.IDGT(id),
					),
				),
			)
		}
		return builder.Where(
			organizationEnt.Or(
				organizationEnt.CreatedAtLT(timestamp),
				organizationEnt.And(
					organizationEnt.CreatedAtEQ(timestamp),
					organizationEnt.IDLT(id),
				),
			),
		)
	default:
		return applyCursorCondition(builder, id, value, direction, structs.SortByCreatedAt)
	}
}

// Delete deletes an organization
func (r *organizationRepository) Delete(ctx context.Context, slug string) error {
	organization, err := r.FindOrganization(ctx, &structs.FindOrganization{Organization: slug})
	if err != nil {
		return err
	}

	builder := r.ec.Organization.Delete()
	if _, err = builder.Where(organizationEnt.IDEQ(organization.ID)).Exec(ctx); err != nil {
		logger.Errorf(ctx, "organizationRepo.Delete error: %v", err)
		return err
	}

	// Delete from Meilisearch
	if err = r.data.DeleteDocument(ctx, "organizations", organization.ID); err != nil {
		logger.Errorf(ctx, "organizationRepo.Delete error deleting Meilisearch index: %v", err)
	}

	// Invalidate cache
	go r.invalidateOrganizationCache(context.Background(), organization)

	return nil
}

// FindOrganization finds an organization.
func (r *organizationRepository) FindOrganization(ctx context.Context, params *structs.FindOrganization) (*ent.Organization, error) {
	// create builder.
	builder := r.ec.Organization.Query()

	// support slug or ID
	if validator.IsNotEmpty(params.Organization) {
		builder = builder.Where(organizationEnt.Or(
			organizationEnt.ID(params.Organization),
			organizationEnt.SlugEQ(params.Organization),
		))
	}

	// execute the builder.
	row, err := builder.Only(ctx)
	if validator.IsNotNil(err) {
		return nil, err
	}

	return row, nil
}

// GetTree retrieves the organization tree.
func (r *organizationRepository) GetTree(ctx context.Context, params *structs.FindOrganization) ([]*ent.Organization, error) {
	// create builder
	builder := r.ec.Organization.Query()

	// handle sub organizations
	if validator.IsNotEmpty(params.Organization) && params.Organization != "root" {
		return r.getSubOrganization(ctx, params.Organization, builder)
	}

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// ListBuilder - create list builder.
func (r *organizationRepository) ListBuilder(_ context.Context, params *structs.ListOrganizationParams) (*ent.OrganizationQuery, error) {
	// create builder.
	builder := r.ec.Organization.Query()

	// match parent id.
	if params.Parent == "" {
		builder.Where(organizationEnt.Or(
			organizationEnt.ParentIDIsNil(),
			organizationEnt.ParentIDEQ(""),
			organizationEnt.ParentIDEQ("root"),
		))
	} else {
		builder.Where(organizationEnt.ParentIDEQ(params.Parent))
	}

	return builder, nil
}

// CountX gets a count of organizations.
func (r *organizationRepository) CountX(ctx context.Context, params *structs.ListOrganizationParams) int {
	// create list builder
	builder, err := r.ListBuilder(ctx, params)
	if validator.IsNotNil(err) {
		return 0
	}
	return builder.CountX(ctx)
}

// getSubOrganization - get sub organizations.
func (r *organizationRepository) getSubOrganization(ctx context.Context, rootID string, builder *ent.OrganizationQuery) ([]*ent.Organization, error) {
	// set where conditions
	builder.Where(
		organizationEnt.Or(
			organizationEnt.ID(rootID),
			organizationEnt.ParentIDHasPrefix(rootID),
		),
	)

	// execute the builder
	return r.executeArrayQuery(ctx, builder)
}

// executeArrayQuery - execute the builder query and return results.
func (r *organizationRepository) executeArrayQuery(ctx context.Context, builder *ent.OrganizationQuery) ([]*ent.Organization, error) {
	rows, err := builder.All(ctx)
	if err != nil {
		logger.Errorf(ctx, "organizationRepo.executeArrayQuery error: %v", err)
		return nil, err
	}
	return rows, nil
}

func (r *organizationRepository) cacheOrganization(ctx context.Context, organization *ent.Organization) {
	// Cache by ID
	idKey := fmt.Sprintf("id:%s", organization.ID)
	if err := r.organizationCache.Set(ctx, idKey, organization, r.organizationTTL); err != nil {
		logger.Debugf(ctx, "Failed to cache organization by ID %s: %v", organization.ID, err)
	}

	// Cache slug to ID mapping
	if organization.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", organization.Slug)
		if err := r.slugMappingCache.Set(ctx, slugKey, &organization.ID, r.organizationTTL); err != nil {
			logger.Debugf(ctx, "Failed to cache slug mapping %s: %v", organization.Slug, err)
		}
	}
}

func (r *organizationRepository) invalidateOrganizationCache(ctx context.Context, organization *ent.Organization) {
	// Invalidate ID cache
	idKey := fmt.Sprintf("id:%s", organization.ID)
	if err := r.organizationCache.Delete(ctx, idKey); err != nil {
		logger.Debugf(ctx, "Failed to invalidate organization ID cache %s: %v", organization.ID, err)
	}

	// Invalidate slug mapping
	if organization.Slug != "" {
		slugKey := fmt.Sprintf("slug:%s", organization.Slug)
		if err := r.slugMappingCache.Delete(ctx, slugKey); err != nil {
			logger.Debugf(ctx, "Failed to invalidate slug mapping cache %s: %v", organization.Slug, err)
		}
	}
}

func (r *organizationRepository) getOrgIDBySlug(ctx context.Context, slug string) (string, error) {
	cacheKey := fmt.Sprintf("slug:%s", slug)
	organizationID, err := r.slugMappingCache.Get(ctx, cacheKey)
	if err != nil || organizationID == nil {
		return "", err
	}
	return *organizationID, nil
}
