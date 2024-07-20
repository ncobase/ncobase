package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/types"
	"ncobase/feature/group/data"
	"ncobase/feature/group/data/ent"
	"ncobase/feature/group/data/repository"
	"ncobase/feature/group/structs"
	"sort"
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

// Get retrieves a group by its ID.
func (s *groupService) Get(ctx context.Context, params *structs.FindGroup) (*structs.ReadGroup, error) {
	row, err := s.group.Get(ctx, params)
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

// List lists all groups.
func (s *groupService) List(ctx context.Context, params *structs.ListGroupParams) (paging.Result[*structs.ReadGroup], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindGroup{
			Children: true,
			Tenant:   params.Tenant,
			Group:    params.Parent,
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

		rows, err := s.group.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing groups: %v\n", err)
			return nil, 0, err
		}

		total := s.group.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
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

// CountX gets a count of groups.
func (s *groupService) CountX(ctx context.Context, params *structs.ListGroupParams) int {
	return s.group.CountX(ctx, params)
}

// GetTree retrieves the group tree.
func (s *groupService) GetTree(ctx context.Context, params *structs.FindGroup) (paging.Result[*structs.ReadGroup], error) {
	rows, err := s.group.GetTree(ctx, params)
	if err := handleEntError("Group", err); err != nil {
		return paging.Result[*structs.ReadGroup]{}, err
	}

	return paging.Result[*structs.ReadGroup]{
		Items: s.buildGroupTree(rows),
		Total: len(rows),
	}, nil
}

// buildGroupTree builds a group tree structure.
func (s *groupService) buildGroupTree(groups []*ent.Group) []*structs.ReadGroup {
	// Convert groups to ReadGroup objects
	groupNodes := make([]types.TreeNode, len(groups))
	for i, m := range groups {
		groupNodes[i] = s.Serialize(m)
	}

	// Sort group nodes
	sortGroupNodes(groupNodes)

	// Build tree structure
	tree := types.BuildTree(groupNodes)

	result := make([]*structs.ReadGroup, len(tree))
	for i, node := range tree {
		result[i] = node.(*structs.ReadGroup)
	}

	return result
}

// sortGroupNodes sorts group nodes.
func sortGroupNodes(groupNodes []types.TreeNode) {
	// Recursively sort children nodes first
	for _, node := range groupNodes {
		children := node.GetChildren()
		sortGroupNodes(children)

		// Sort children and set back to node
		sort.SliceStable(children, func(i, j int) bool {
			nodeI := children[i].(*structs.ReadGroup)
			nodeJ := children[j].(*structs.ReadGroup)
			return types.ToValue(nodeI.CreatedAt) < (types.ToValue(nodeJ.CreatedAt))
		})
		node.SetChildren(children)
	}

	// Sort the immediate children of the current level
	sort.SliceStable(groupNodes, func(i, j int) bool {
		nodeI := groupNodes[i].(*structs.ReadGroup)
		nodeJ := groupNodes[j].(*structs.ReadGroup)
		return types.ToValue(nodeI.CreatedAt) < (types.ToValue(nodeJ.CreatedAt))
	})
}
