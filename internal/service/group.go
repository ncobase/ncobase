package service

import (
	"context"
	"ncobase/internal/data/ent"
	"ncobase/internal/data/structs"
	"ncobase/pkg/resp"
	"ncobase/pkg/types"
)

// CreateGroupService creates a new group.
func (svc *Service) CreateGroupService(ctx context.Context, body *structs.CreateGroupBody) (*resp.Exception, error) {
	if body.Name == "" {
		return resp.BadRequest("Group name is required"), nil
	}

	group, err := svc.group.Create(ctx, body)
	if exception, err := handleError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeGroup(group),
	}, nil
}

// UpdateGroupService updates an existing group.
func (svc *Service) UpdateGroupService(ctx context.Context, groupID string, updates types.JSON) (*resp.Exception, error) {
	group, err := svc.group.Update(ctx, groupID, updates)
	if exception, err := handleError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeGroup(group),
	}, nil
}

// GetGroupByIDService retrieves a group by its ID.
func (svc *Service) GetGroupByIDService(ctx context.Context, groupID string) (*resp.Exception, error) {
	group, err := svc.group.GetByID(ctx, groupID)
	if exception, err := handleError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: svc.serializeGroup(group),
	}, nil
}

// DeleteGroupService deletes a group by its ID.
func (svc *Service) DeleteGroupService(ctx context.Context, groupID string) (*resp.Exception, error) {
	err := svc.group.Delete(ctx, groupID)
	if exception, err := handleError("Group", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Group deleted successfully",
	}, nil
}

// AddRoleToGroupService adds a role to a group.
func (svc *Service) AddRoleToGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error) {
	_, err := svc.groupRole.Create(ctx, &structs.GroupRole{GroupID: groupID, RoleID: roleID})
	if exception, err := handleError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role added to group successfully",
	}, nil
}

// RemoveRoleFromGroupService removes a role from a group.
func (svc *Service) RemoveRoleFromGroupService(ctx context.Context, groupID string, roleID string) (*resp.Exception, error) {
	err := svc.groupRole.Delete(ctx, groupID, roleID)
	if exception, err := handleError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: "Role removed from group successfully",
	}, nil
}

// GetGroupRolesService retrieves roles associated with a group.
func (svc *Service) GetGroupRolesService(ctx context.Context, groupID string) (*resp.Exception, error) {
	roles, err := svc.groupRole.GetRolesByGroupID(ctx, groupID)
	if exception, err := handleError("GroupRole", err); exception != nil {
		return exception, err
	}

	return &resp.Exception{
		Data: roles,
	}, nil
}

// ****** Internal methods of service

// serializeGroup serializes a group entity to a response format.
func (svc *Service) serializeGroup(row *ent.Group) *structs.ReadGroup {
	return &structs.ReadGroup{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Disabled:    row.Disabled,
		Description: row.Description,
		Leader:      &row.Leader,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		DomainID:    &row.DomainID,
		BaseEntity: structs.BaseEntity{
			CreatedBy: &row.CreatedBy,
			CreatedAt: &row.CreatedAt,
			UpdatedBy: &row.UpdatedBy,
			UpdatedAt: &row.UpdatedAt},
	}
}
