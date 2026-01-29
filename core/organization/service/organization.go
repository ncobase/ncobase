package service

import (
	"context"
	"errors"
	"ncobase/core/organization/data"
	"ncobase/core/organization/data/repository"
	"ncobase/core/organization/structs"

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

	return repository.SerializeOrganization(row), nil
}

// Update updates an existing organization.
func (s *organizationService) Update(ctx context.Context, organizationID string, updates types.JSON) (*structs.ReadOrganization, error) {
	row, err := s.r.Update(ctx, organizationID, updates)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return repository.SerializeOrganization(row), nil
}

// Get retrieves an organization by its ID.
func (s *organizationService) Get(ctx context.Context, params *structs.FindOrganization) (*structs.ReadOrganization, error) {
	row, err := s.r.Get(ctx, params)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}
	return repository.SerializeOrganization(row), nil
}

// GetByIDs retrieves organizations by their IDs.
func (s *organizationService) GetByIDs(ctx context.Context, organizationIDs []string) ([]*structs.ReadOrganization, error) {
	rows, err := s.r.GetByIDs(ctx, organizationIDs)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return nil, err
	}

	return repository.SerializeOrganizations(rows), nil
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
		if repository.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
			logger.Errorf(ctx, "Error listing organizations: %v", err)
			return nil, 0, err
		}

		return repository.SerializeOrganizations(rows), total, nil
	})
}

// CountX gets a count of organizations.
func (s *organizationService) CountX(ctx context.Context, params *structs.ListOrganizationParams) int {
	return s.r.CountX(ctx, params)
}

// GetTree retrieves the organization tree.
func (s *organizationService) GetTree(ctx context.Context, params *structs.FindOrganization) (paging.Result[*structs.ReadOrganization], error) {
	// Get all organizations for tree
	rows, err := s.r.GetTree(ctx, params)
	if err := handleEntError(ctx, "Organization", err); err != nil {
		return paging.Result[*structs.ReadOrganization]{}, err
	}

	return paging.Result[*structs.ReadOrganization]{
		Items: s.buildOrganizationTree(repository.SerializeOrganizations(rows)),
		Total: len(rows),
	}, nil
}

// buildOrganizationTree builds an organization tree structure.
func (s *organizationService) buildOrganizationTree(organizations []*structs.ReadOrganization) []*structs.ReadOrganization {
	tree := types.BuildTree(organizations, string(structs.SortByCreatedAt))
	return tree
}
