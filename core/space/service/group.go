package service

import (
	"context"
	"errors"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// GroupServiceInterface is the interface for the service.
type GroupServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateGroupBody) (*structs.ReadGroup, error)
	Update(ctx context.Context, groupID string, updates types.JSON) (*structs.ReadGroup, error)
	Get(ctx context.Context, params *structs.FindGroup) (*structs.ReadGroup, error)
	GetByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error)
	Delete(ctx context.Context, groupID string) error
	List(ctx context.Context, params *structs.ListGroupParams) (paging.Result[*structs.ReadGroup], error)
	CountX(ctx context.Context, params *structs.ListGroupParams) int
	GetTree(ctx context.Context, params *structs.FindGroup) (paging.Result[*structs.ReadGroup], error)
	Serializes(rows []*ent.Group) []*structs.ReadGroup
	Serialize(group *ent.Group) *structs.ReadGroup
}

// spaceService is the struct for the service.
type spaceService struct {
	r repository.GroupRepositoryInterface
}

// NewGroupService creates a new service.
func NewGroupService(d *data.Data) GroupServiceInterface {
	return &spaceService{
		r: repository.NewGroupRepository(d),
	}
}

// Create creates a new group.
func (s *spaceService) Create(ctx context.Context, body *structs.CreateGroupBody) (*structs.ReadGroup, error) {
	if body.Name == "" {
		return nil, errors.New("group name is required")
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing group.
func (s *spaceService) Update(ctx context.Context, groupID string, updates types.JSON) (*structs.ReadGroup, error) {
	row, err := s.r.Update(ctx, groupID, updates)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a group by its ID.
func (s *spaceService) Get(ctx context.Context, params *structs.FindGroup) (*structs.ReadGroup, error) {
	row, err := s.r.Get(ctx, params)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByIDs retrieves groups by their IDs.
func (s *spaceService) GetByIDs(ctx context.Context, groupIDs []string) ([]*structs.ReadGroup, error) {
	rows, err := s.r.GetByIDs(ctx, groupIDs)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
}

// Delete deletes a group by its ID.
func (s *spaceService) Delete(ctx context.Context, groupID string) error {
	err := s.r.Delete(ctx, groupID)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return err
	}

	return nil
}

// List lists all groups.
func (s *spaceService) List(ctx context.Context, params *structs.ListGroupParams) (paging.Result[*structs.ReadGroup], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindGroup{
			Children: true,
			Tenant:   params.Tenant,
			Group:    params.Parent,
			SortBy:   params.SortBy,
		})
	}

	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadGroup, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.r.ListWithCount(ctx, &lp)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
			}
			logger.Errorf(ctx, "Error listing groups: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes groups.
func (s *spaceService) Serializes(rows []*ent.Group) []*structs.ReadGroup {
	var rs []*structs.ReadGroup
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a group.
func (s *spaceService) Serialize(row *ent.Group) *structs.ReadGroup {
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

// CountX gets a count of groups.
func (s *spaceService) CountX(ctx context.Context, params *structs.ListGroupParams) int {
	return s.r.CountX(ctx, params)
}

// GetTree retrieves the group tree.
func (s *spaceService) GetTree(ctx context.Context, params *structs.FindGroup) (paging.Result[*structs.ReadGroup], error) {
	rows, err := s.r.GetTree(ctx, params)
	if err := handleEntError(ctx, "Group", err); err != nil {
		return paging.Result[*structs.ReadGroup]{}, err
	}

	return paging.Result[*structs.ReadGroup]{
		Items: s.buildGroupTree(rows),
		Total: len(rows),
	}, nil
}

// buildGroupTree builds a group tree structure.
func (s *spaceService) buildGroupTree(groups []*ent.Group) []*structs.ReadGroup {
	groupNodes := make([]*structs.ReadGroup, len(groups))
	for i, m := range groups {
		groupNodes[i] = s.Serialize(m)
	}

	tree := types.BuildTree(groupNodes, string(structs.SortByCreatedAt))
	return tree
}
