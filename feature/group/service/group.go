package service

import (
	"context"
	"errors"
	"ncobase/common/types"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/ent"
	"ncobase/feature/group/data/repository"
	"ncobase/feature/group/structs"
)

// GroupServiceInterface is the interface for the service.
type GroupServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateGroupBody) (*structs.ReadGroup, error)
	Update(ctx context.Context, groupID string, updates types.JSON) (*structs.ReadGroup, error)
	GetByID(ctx context.Context, groupID string) (*structs.ReadGroup, error)
	GetByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error)
	Delete(ctx context.Context, groupID string) error
	Serializes(rows []*ent.Group) []*structs.ReadGroup
	Serialize(group *ent.Group) *structs.ReadGroup
}

// groupService is the struct for the service.
type groupService struct {
	group repository.GroupRepositoryInterface
}

// NewGroupService creates a new service.
func NewGroupService(d *data.Data) GroupServiceInterface {
	return &groupService{
		group: repository.NewGroupRepository(d),
	}
}

// Create creates a new group.
func (s *groupService) Create(ctx context.Context, body *structs.CreateGroupBody) (*structs.ReadGroup, error) {
	if body.Name == "" {
		return nil, errors.New("group name is required")
	}

	row, err := s.group.Create(ctx, body)
	if err := handleEntError("Group", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing group.
func (s *groupService) Update(ctx context.Context, groupID string, updates types.JSON) (*structs.ReadGroup, error) {
	row, err := s.group.Update(ctx, groupID, updates)
	if err := handleEntError("Group", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByID retrieves a group by its ID.
func (s *groupService) GetByID(ctx context.Context, groupID string) (*structs.ReadGroup, error) {
	row, err := s.group.GetByID(ctx, groupID)
	if err := handleEntError("Group", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByIDs retrieves groups by their IDs.
func (s *groupService) GetByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error) {
	rows, err := s.group.GetByIDs(ctx, groupIDs)
	if err := handleEntError("Group", err); err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
}

// Delete deletes a group by its ID.
func (s *groupService) Delete(ctx context.Context, groupID string) error {
	err := s.group.Delete(ctx, groupID)
	if err := handleEntError("Group", err); err != nil {
		return err
	}

	return nil
}

// Serializes serializes groups.
func (s *groupService) Serializes(rows []*ent.Group) []*structs.ReadGroup {
	var rs []*structs.ReadGroup
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a group.
func (s *groupService) Serialize(row *ent.Group) *structs.ReadGroup {
	return &structs.ReadGroup{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Disabled:    row.Disabled,
		Description: row.Description,
		Leader:      &row.Leader,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		TenantID:    &row.TenantID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
