package service

import (
	"context"
	"errors"
	"ncobase/organization/data"
	"ncobase/organization/data/ent"
	"ncobase/organization/data/repository"
	"ncobase/organization/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// OrganizationServiceInterface is the interface for the service.
type OrganizationServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateOrganizationBody) (*structs.ReadOrganization, error)
	Update(ctx context.Context, organizationID string, updates types.JSON) (*structs.ReadOrganization, error)
	Get(ctx context.Context, params *structs.FindOrganization) (*structs.ReadOrganization, error)
	GetByIDs(ctx context.Context, organizationIDs []string) ([]*structs.ReadOrganization, error)
	Delete(ctx context.Context, organizationID string) error
	List(ctx context.Context, params *structs.ListOrganizationParams) (paging.Result[*structs.ReadOrganization], error)
	CountX(ctx context.Context, params *structs.ListOrganizationParams) int
	GetTree(ctx context.Context, params *structs.FindOrganization) (paging.Result[*structs.ReadOrganization], error)
	Serializes(rows []*ent.Organization) []*structs.ReadOrganization
	Serialize(organization *ent.Organization) *structs.ReadOrganization
}

// organizationService is the struct for the service.
type organizationService struct {
	r repository.OrganizationRepositoryInterface
}

// NewOrganizationService creates a new service.
func NewOrganizationService(d *data.Data) OrganizationServiceInterface {
	return &organizationService{
		r: repository.NewOrganizationRepository(d),
	}
}

// Create creates a new organization.
func (s *organizationService) Create(ctx context.Context, body *structs.CreateOrganizationBody) (*structs.ReadOrganization, error) {
	if body.Name == "" {
		return nil, errors.New("organization name is required")
	}

	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing organization.
func (s *organizationService) Update(ctx context.Context, organizationID string, updates types.JSON) (*structs.ReadOrganization, error) {
	row, err := s.r.Update(ctx, organizationID, updates)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves an organization by its ID.
func (s *organizationService) Get(ctx context.Context, params *structs.FindOrganization) (*structs.ReadOrganization, error) {
	row, err := s.r.Get(ctx, params)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}
	return s.Serialize(row), nil
}

// GetByIDs retrieves organizations by their IDs.
func (s *organizationService) GetByIDs(ctx context.Context, organizationIDs []string) ([]*structs.ReadOrganization, error) {
	rows, err := s.r.GetByIDs(ctx, organizationIDs)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return s.Serializes(rows), nil
}

// Delete deletes an organization by its ID.
func (s *organizationService) Delete(ctx context.Context, organizationID string) error {
	err := s.r.Delete(ctx, organizationID)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return err
	}

	return nil
}

// List lists all organizations.
func (s *organizationService) List(ctx context.Context, params *structs.ListOrganizationParams) (paging.Result[*structs.ReadOrganization], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindOrganization{
			Children: true,
			Parent:   params.Parent,
			SortBy:   params.SortBy,
		})
	}

	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadOrganization, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.r.ListWithCount(ctx, &lp)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
			}
			logger.Errorf(ctx, "Error listing organizations: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes organizations.
func (s *organizationService) Serializes(rows []*ent.Organization) []*structs.ReadOrganization {
	rs := make([]*structs.ReadOrganization, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes an organization.
func (s *organizationService) Serialize(row *ent.Organization) *structs.ReadOrganization {
	return &structs.ReadOrganization{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Type:        row.Type,
		Disabled:    row.Disabled,
		Description: row.Description,
		Leader:      &row.Leader,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// CountX gets a count of organizations.
func (s *organizationService) CountX(ctx context.Context, params *structs.ListOrganizationParams) int {
	return s.r.CountX(ctx, params)
}

// GetTree retrieves the organization tree.
func (s *organizationService) GetTree(ctx context.Context, params *structs.FindOrganization) (paging.Result[*structs.ReadOrganization], error) {
	var rows []*ent.Organization
	var err error

	// Get all organizations for tree
	rows, err = s.r.GetTree(ctx, params)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return paging.Result[*structs.ReadOrganization]{}, err
	}

	return paging.Result[*structs.ReadOrganization]{
		Items: s.buildOrganizationTree(rows),
		Total: len(rows),
	}, nil
}

// buildOrganizationTree builds an organization tree structure.
func (s *organizationService) buildOrganizationTree(organizations []*ent.Organization) []*structs.ReadOrganization {
	organizationNodes := make([]*structs.ReadOrganization, len(organizations))
	for i, m := range organizations {
		organizationNodes[i] = s.Serialize(m)
	}

	tree := types.BuildTree(organizationNodes, string(structs.SortByCreatedAt))
	return tree
}
