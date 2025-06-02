package service

import (
	"context"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
	"ncobase/tenant/wrapper"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
)

// TenantGroupServiceInterface defines the tenant group service interface
type TenantGroupServiceInterface interface {
	AddGroupToTenant(ctx context.Context, tenantID string, groupID string) (*structs.TenantGroupRelation, error)
	RemoveGroupFromTenant(ctx context.Context, tenantID string, groupID string) error
	GetTenantGroups(ctx context.Context, tenantID string, params *structs.ListGroupParams) (paging.Result[*structs.ReadGroup], error)
	GetGroupTenants(ctx context.Context, groupID string) ([]string, error)
	IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error)
	GetTenantGroupIDs(ctx context.Context, tenantID string) ([]string, error)
	Serialize(row *ent.TenantGroup) *structs.TenantGroup
	Serializes(rows []*ent.TenantGroup) []*structs.TenantGroup
}

// tenantGroupService implements TenantGroupServiceInterface
type tenantGroupService struct {
	gsw *wrapper.SpaceServiceWrapper
	r   repository.TenantGroupRepositoryInterface
}

// NewTenantGroupService creates a new tenant group service
func NewTenantGroupService(d *data.Data, gsw *wrapper.SpaceServiceWrapper) TenantGroupServiceInterface {
	return &tenantGroupService{
		gsw: gsw,
		r:   repository.NewTenantGroupRepository(d),
	}
}

// AddGroupToTenant adds a group to a tenant
func (s *tenantGroupService) AddGroupToTenant(ctx context.Context, tenantID string, groupID string) (*structs.TenantGroupRelation, error) {
	tenantGroup := &structs.TenantGroup{
		TenantID: tenantID,
		GroupID:  groupID,
	}

	_, err := s.r.Create(ctx, tenantGroup)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return nil, err
	}

	relation := &structs.TenantGroupRelation{
		TenantID: tenantID,
		GroupID:  groupID,
		AddedAt:  time.Now().UnixMilli(),
	}

	return relation, nil
}

// RemoveGroupFromTenant removes a group from a tenant
func (s *tenantGroupService) RemoveGroupFromTenant(ctx context.Context, tenantID string, groupID string) error {
	err := s.r.Delete(ctx, tenantID, groupID)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return err
	}
	return nil
}

// GetTenantGroups retrieves all groups for a tenant
func (s *tenantGroupService) GetTenantGroups(ctx context.Context, tenantID string, params *structs.ListGroupParams) (paging.Result[*structs.ReadGroup], error) {
	groupIDs, err := s.r.GetGroupsByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return paging.Result[*structs.ReadGroup]{}, err
	}

	if len(groupIDs) == 0 {
		return paging.Result[*structs.ReadGroup]{
			Items: []*structs.ReadGroup{},
			Total: 0,
		}, nil
	}

	// Try to get full group information from space module
	var groups []*structs.ReadGroup
	if s.gsw != nil && s.gsw.HasGroupService() {
		groups, err = s.gsw.GetGroupByIDs(ctx, groupIDs)
		if err != nil {
			logger.Warnf(ctx, "Failed to get groups from space service: %v", err)
			// Create minimal group info as fallback
			groups = make([]*structs.ReadGroup, len(groupIDs))
			for i, id := range groupIDs {
				groups[i] = &structs.ReadGroup{ID: id, Name: "Group " + id}
			}
		}
	} else {
		// Fallback when space service is not available
		groups = make([]*structs.ReadGroup, len(groupIDs))
		for i, id := range groupIDs {
			groups[i] = &structs.ReadGroup{ID: id, Name: "Group " + id}
		}
	}

	// Apply parent filtering if specified
	var filteredGroups []*structs.ReadGroup
	if params.Parent != "" {
		for _, group := range groups {
			if group.ParentID != nil && *group.ParentID == params.Parent {
				filteredGroups = append(filteredGroups, group)
			}
		}
	} else {
		// Show root level groups (no parent or parent is empty/root)
		for _, group := range groups {
			if group.ParentID == nil || *group.ParentID == "" || *group.ParentID == "root" {
				filteredGroups = append(filteredGroups, group)
			}
		}
	}

	return paging.Result[*structs.ReadGroup]{
		Items: filteredGroups,
		Total: len(filteredGroups),
	}, nil
}

// GetGroupTenants retrieves all tenants that have a specific group
func (s *tenantGroupService) GetGroupTenants(ctx context.Context, groupID string) ([]string, error) {
	tenantIDs, err := s.r.GetTenantsByGroupID(ctx, groupID)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return nil, err
	}

	return tenantIDs, nil
}

// IsGroupInTenant checks if a group belongs to a tenant
func (s *tenantGroupService) IsGroupInTenant(ctx context.Context, tenantID string, groupID string) (bool, error) {
	isInTenant, err := s.r.IsGroupInTenant(ctx, tenantID, groupID)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return false, err
	}

	return isInTenant, nil
}

// GetTenantGroupIDs retrieves all group IDs for a tenant
func (s *tenantGroupService) GetTenantGroupIDs(ctx context.Context, tenantID string) ([]string, error) {
	groupIDs, err := s.r.GetGroupsByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantGroup", err); err != nil {
		return nil, err
	}

	return groupIDs, nil
}

// Serializes serializes tenant groups
func (s *tenantGroupService) Serializes(rows []*ent.TenantGroup) []*structs.TenantGroup {
	var rs []*structs.TenantGroup
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a tenant group
func (s *tenantGroupService) Serialize(row *ent.TenantGroup) *structs.TenantGroup {
	return &structs.TenantGroup{
		TenantID: row.TenantID,
		GroupID:  row.GroupID,
	}
}
