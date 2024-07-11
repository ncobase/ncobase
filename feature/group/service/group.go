package service

import (
	"context"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/ent"
	"ncobase/feature/group/data/repository"
	"ncobase/feature/group/structs"
)

// GroupServiceInterface is the interface for the service.
type GroupServiceInterface interface {
	CreateGroupService(ctx context.Context, body *structs.CreateGroupBody) (*resp.Exception, error)
	UpdateGroupService(ctx context.Context, groupID string, updates types.JSON) (*resp.Exception, error)
	GetGroupByIDService(ctx context.Context, groupID string) (*resp.Exception, error)
	DeleteGroupService(ctx context.Context, groupID string) (*resp.Exception, error)
	SerializeGroup(group *ent.Group) *structs.ReadGroup
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

// CreateGroupService creates a new group.
func (s *groupService) CreateGroupService(ctx context.Context, body *structs.CreateGroupBody) (*resp.Exception, error) {
	if body.Name == "" {
		return resp.BadRequest("Group name is required"), nil
	}

	group, err := s.group.Create(ctx, body)
	if exception, err := handleEntError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeGroup(group),
	}, nil
}

// UpdateGroupService updates an existing group.
func (s *groupService) UpdateGroupService(ctx context.Context, groupID string, updates types.JSON) (*resp.Exception, error) {
	group, err := s.group.Update(ctx, groupID, updates)
	if exception, err := handleEntError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeGroup(group),
	}, nil
}

// GetGroupByIDService retrieves a group by its ID.
func (s *groupService) GetGroupByIDService(ctx context.Context, groupID string) (*resp.Exception, error) {
	group, err := s.group.GetByID(ctx, groupID)
	if exception, err := handleEntError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: s.SerializeGroup(group),
	}, nil
}

// DeleteGroupService deletes a group by its ID.
func (s *groupService) DeleteGroupService(ctx context.Context, groupID string) (*resp.Exception, error) {
	err := s.group.Delete(ctx, groupID)
	if exception, err := handleEntError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Group deleted successfully",
	}, nil
}

// SerializeGroup serializes a group entity to a response format.
func (s *groupService) SerializeGroup(row *ent.Group) *structs.ReadGroup {
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
