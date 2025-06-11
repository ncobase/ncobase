package service

import (
	"context"
	"errors"
	"ncobase/content/data"
	"ncobase/content/data/ent"
	"ncobase/content/data/repository"
	"ncobase/content/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/slug"
	"github.com/ncobase/ncore/validation/validator"
)

// TaxonomyServiceInterface is the interface for the service.
type TaxonomyServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*structs.ReadTaxonomy, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTaxonomy, error)
	Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) (paging.Result[*structs.ReadTaxonomy], error)
	CountX(ctx context.Context, params *structs.ListTaxonomyParams) int
	GetTree(ctx context.Context, params *structs.FindTaxonomy) (paging.Result[*structs.ReadTaxonomy], error)
	Delete(ctx context.Context, slug string) error
	Serializes(rows []*ent.Taxonomy) []*structs.ReadTaxonomy
	Serialize(row *ent.Taxonomy) *structs.ReadTaxonomy
}

// taxonomyService is the struct for the service.
type taxonomyService struct {
	r repository.TaxonomyRepositoryInterface
}

// NewTaxonomyService creates a new service.
func NewTaxonomyService(d *data.Data) TaxonomyServiceInterface {
	return &taxonomyService{
		r: repository.NewTaxonomyRepository(d),
	}
}

// Create creates a new taxonomy.
func (s *taxonomyService) Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*structs.ReadTaxonomy, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsRequired("name"))
	}
	if validator.IsEmpty(body.Type) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}
	// set slug field.
	if validator.IsEmpty(body.Slug) {
		body.Slug = slug.Unicode(body.Name)
	}
	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing taxonomy (full and partial)..
func (s *taxonomyService) Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTaxonomy, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug / id"))
	}

	// Validate the updates map
	if len(updates) == 0 {
		return nil, errors.New(ecode.FieldIsEmpty("updates fields"))
	}

	row, err := s.r.Update(ctx, slug, updates)
	if err := handleEntError(ctx, "Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy by ID.
func (s *taxonomyService) Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error) {
	row, err := s.r.GetBySlug(ctx, slug)
	if err := handleEntError(ctx, "Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy by ID.
func (s *taxonomyService) Delete(ctx context.Context, slug string) error {
	err := s.r.Delete(ctx, slug)
	if err := handleEntError(ctx, "Taxonomy", err); err != nil {
		return err
	}
	return nil
}

// List lists all taxonomies.
func (s *taxonomyService) List(ctx context.Context, params *structs.ListTaxonomyParams) (paging.Result[*structs.ReadTaxonomy], error) {
	if params.Children {
		return s.GetTree(ctx, &structs.FindTaxonomy{
			Children: true,
			SpaceID:  params.SpaceID,
			Taxonomy: params.Parent,
			SortBy:   params.SortBy,
		})
	}
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTaxonomy, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.r.ListWithCount(ctx, &lp)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
			}
			logger.Errorf(ctx, "Error listing taxonomies: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes taxonomies.
func (s *taxonomyService) Serializes(rows []*ent.Taxonomy) []*structs.ReadTaxonomy {
	rs := make([]*structs.ReadTaxonomy, 0, len(rows))
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a taxonomy.
func (s *taxonomyService) Serialize(row *ent.Taxonomy) *structs.ReadTaxonomy {
	return &structs.ReadTaxonomy{
		ID:          row.ID,
		Name:        row.Name,
		Type:        row.Type,
		Slug:        row.Slug,
		Cover:       row.Cover,
		Thumbnail:   row.Thumbnail,
		Color:       row.Color,
		Icon:        row.Icon,
		URL:         row.URL,
		Keywords:    row.Keywords,
		Description: row.Description,
		Status:      row.Status,
		Extras:      &row.Extras,
		ParentID:    &row.ParentID,
		SpaceID:     row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}

// CountX gets a count of taxonomys.
func (s *taxonomyService) CountX(ctx context.Context, params *structs.ListTaxonomyParams) int {
	return s.r.CountX(ctx, params)
}

// GetTree retrieves the taxonomy tree.
func (s *taxonomyService) GetTree(ctx context.Context, params *structs.FindTaxonomy) (paging.Result[*structs.ReadTaxonomy], error) {
	rows, err := s.r.GetTree(ctx, params)
	if err := handleEntError(ctx, "Taxonomy", err); err != nil {
		return paging.Result[*structs.ReadTaxonomy]{}, err
	}

	return paging.Result[*structs.ReadTaxonomy]{
		Items: s.buildTaxonomyTree(rows),
		Total: len(rows),
	}, nil
}

// buildTaxonomyTree builds a taxonomy tree structure.
func (s *taxonomyService) buildTaxonomyTree(taxonomies []*ent.Taxonomy) []*structs.ReadTaxonomy {
	taxonomyNodes := make([]*structs.ReadTaxonomy, len(taxonomies))
	for i, m := range taxonomies {
		taxonomyNodes[i] = s.Serialize(m)
	}

	tree := types.BuildTree(taxonomyNodes, string(structs.SortByCreatedAt))
	return tree
}
