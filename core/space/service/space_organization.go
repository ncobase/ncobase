package service

import (
	"context"
	"ncobase/core/space/data"
	"ncobase/core/space/data/ent"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"
	"ncobase/core/space/wrapper"
	"time"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/logging/logger"
)

// SpaceOrganizationServiceInterface defines the space organization service interface
type SpaceOrganizationServiceInterface interface {
	AddGroupToSpace(ctx context.Context, spaceID string, orgID string) (*structs.SpaceOrganizationRelation, error)
	RemoveGroupFromSpace(ctx context.Context, spaceID string, orgID string) error
	GetSpaceOrganizations(ctx context.Context, spaceID string, params *structs.ListOrganizationParams) (paging.Result[*structs.ReadOrganization], error)
	GetOrganizationSpaces(ctx context.Context, orgID string) ([]string, error)
	IsGroupInSpace(ctx context.Context, spaceID string, orgID string) (bool, error)
	GetSpaceOrgIDs(ctx context.Context, spaceID string) ([]string, error)
	Serialize(row *ent.SpaceOrganization) *structs.SpaceOrganization
	Serializes(rows []*ent.SpaceOrganization) []*structs.SpaceOrganization
}

// spaceOrganizationService implements SpaceOrganizationServiceInterface
type spaceOrganizationService struct {
	gsw *wrapper.OrganizationServiceWrapper
	r   repository.SpaceOrganizationRepositoryInterface
}

// NewSpaceOrganizationService creates a new space organization service
func NewSpaceOrganizationService(d *data.Data, gsw *wrapper.OrganizationServiceWrapper) SpaceOrganizationServiceInterface {
	return &spaceOrganizationService{
		gsw: gsw,
		r:   repository.NewSpaceOrganizationRepository(d),
	}
}

// AddGroupToSpace adds a organization to a space
func (s *spaceOrganizationService) AddGroupToSpace(ctx context.Context, spaceID string, orgID string) (*structs.SpaceOrganizationRelation, error) {
	spaceGroup := &structs.SpaceOrganization{
		SpaceID: spaceID,
		OrgID:   orgID,
	}

	_, err := s.r.Create(ctx, spaceGroup)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return nil, err
	}

	relation := &structs.SpaceOrganizationRelation{
		SpaceID: spaceID,
		OrgID:   orgID,
		AddedAt: time.Now().UnixMilli(),
	}

	return relation, nil
}

// RemoveGroupFromSpace removes a organization from a space
func (s *spaceOrganizationService) RemoveGroupFromSpace(ctx context.Context, spaceID string, orgID string) error {
	err := s.r.Delete(ctx, spaceID, orgID)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return err
	}
	return nil
}

// GetSpaceOrganizations retrieves all orgs for a space
func (s *spaceOrganizationService) GetSpaceOrganizations(ctx context.Context, spaceID string, params *structs.ListOrganizationParams) (paging.Result[*structs.ReadOrganization], error) {
	orgIDs, err := s.r.GetOrgsBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return paging.Result[*structs.ReadOrganization]{}, err
	}

	if len(orgIDs) == 0 {
		return paging.Result[*structs.ReadOrganization]{
			Items: []*structs.ReadOrganization{},
			Total: 0,
		}, nil
	}

	// Try to get full group information from space module
	var orgs []*structs.ReadOrganization
	if s.gsw != nil && s.gsw.HasOrganizationService() {
		orgs, err = s.gsw.GetOrganizationByIDs(ctx, orgIDs)
		if err != nil {
			logger.Warnf(ctx, "Failed to get orgs from organization service: %v", err)
			// Create minimal group info as fallback
			orgs = make([]*structs.ReadOrganization, len(orgIDs))
			for i, id := range orgIDs {
				orgs[i] = &structs.ReadOrganization{ID: id, Name: "Group " + id}
			}
		}
	} else {
		// Fallback when organization service is not available
		orgs = make([]*structs.ReadOrganization, len(orgIDs))
		for i, id := range orgIDs {
			orgs[i] = &structs.ReadOrganization{ID: id, Name: "Group " + id}
		}
	}

	// Apply parent filtering if specified
	var filteredGroups []*structs.ReadOrganization
	if params.Parent != "" {
		for _, group := range orgs {
			if group.ParentID != nil && *group.ParentID == params.Parent {
				filteredGroups = append(filteredGroups, group)
			}
		}
	} else {
		// Show root level orgs (no parent or parent is empty/root)
		for _, group := range orgs {
			if group.ParentID == nil || *group.ParentID == "" || *group.ParentID == "root" {
				filteredGroups = append(filteredGroups, group)
			}
		}
	}

	return paging.Result[*structs.ReadOrganization]{
		Items: filteredGroups,
		Total: len(filteredGroups),
	}, nil
}

// GetOrganizationSpaces retrieves all spaces that have a specific group
func (s *spaceOrganizationService) GetOrganizationSpaces(ctx context.Context, orgID string) ([]string, error) {
	spaceIDs, err := s.r.GetSpacesByOrgID(ctx, orgID)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return nil, err
	}

	return spaceIDs, nil
}

// IsGroupInSpace checks if a organization belongs to a space
func (s *spaceOrganizationService) IsGroupInSpace(ctx context.Context, spaceID string, orgID string) (bool, error) {
	isInSpace, err := s.r.IsGroupInSpace(ctx, spaceID, orgID)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return false, err
	}

	return isInSpace, nil
}

// GetSpaceOrgIDs retrieves all group IDs for a space
func (s *spaceOrganizationService) GetSpaceOrgIDs(ctx context.Context, spaceID string) ([]string, error) {
	orgIDs, err := s.r.GetOrgsBySpaceID(ctx, spaceID)
	if err := handleEntError(ctx, "SpaceOrganization", err); err != nil {
		return nil, err
	}

	return orgIDs, nil
}

// Serializes serializes space orgs
func (s *spaceOrganizationService) Serializes(rows []*ent.SpaceOrganization) []*structs.SpaceOrganization {
	rs := make([]*structs.SpaceOrganization, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a space group
func (s *spaceOrganizationService) Serialize(row *ent.SpaceOrganization) *structs.SpaceOrganization {
	return &structs.SpaceOrganization{
		SpaceID: row.SpaceID,
		OrgID:   row.OrgID,
	}
}
