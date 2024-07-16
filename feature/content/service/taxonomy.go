package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/slug"
	"ncobase/common/types"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository"
	"ncobase/feature/content/structs"
)

// TaxonomyServiceInterface is the interface for the service.
type TaxonomyServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyBody) (*structs.ReadTaxonomy, error)
	Update(ctx context.Context, slug string, updates types.JSON) (*structs.ReadTaxonomy, error)
	Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error)
	List(ctx context.Context, params *structs.ListTaxonomyParams) (*paging.Result[*structs.ReadTaxonomy], error)
	Delete(ctx context.Context, slug string) error
}

// taxonomyService is the struct for the service.
type taxonomyService struct {
	taxonomy          repository.TaxonomyRepositoryInterface
	taxonomyRelations repository.TaxonomyRelationsRepositoryInterface
}

// NewTaxonomyService creates a new service.
func NewTaxonomyService(d *data.Data) TaxonomyServiceInterface {
	return &taxonomyService{
		taxonomy:          repository.NewTaxonomyRepository(d),
		taxonomyRelations: repository.NewTaxonomyRelationsRepository(d),
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
	row, err := s.taxonomy.Create(ctx, body)
	if err := handleEntError("Taxonomy", err); err != nil {
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

	row, err := s.taxonomy.Update(ctx, slug, updates)
	if err := handleEntError("Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy by ID.
func (s *taxonomyService) Get(ctx context.Context, slug string) (*structs.ReadTaxonomy, error) {
	row, err := s.taxonomy.GetBySlug(ctx, slug)
	if err := handleEntError("Taxonomy", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy by ID.
func (s *taxonomyService) Delete(ctx context.Context, slug string) error {
	err := s.taxonomy.Delete(ctx, slug)
	if err := handleEntError("Taxonomy", err); err != nil {
		return err
	}
	return nil
}

// List lists all taxonomies.
func (s *taxonomyService) List(ctx context.Context, params *structs.ListTaxonomyParams) (*paging.Result[*structs.ReadTaxonomy], error) {
	pp := paging.Params{
		Cursor: params.Cursor,
		Limit:  params.Limit,
	}

	return paging.Paginate(pp, func(cursor string, limit int) ([]*structs.ReadTaxonomy, int, string, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit

		rows, err := s.taxonomy.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, "", errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if validator.IsNotNil(err) {
			log.Errorf(ctx, "Error listing taxonomies: %v\n", err)
			return nil, 0, "", err
		}
		if err != nil {
			return nil, 0, "", err
		}

		total := s.taxonomy.CountX(ctx, params)

		var nextCursor string
		if len(rows) > 0 {
			nextCursor = paging.EncodeCursor(rows[len(rows)-1].CreatedAt)
		}

		return s.Serializes(rows), total, nextCursor, nil
	})
}

// Serializes serializes taxonomies.
func (s *taxonomyService) Serializes(rows []*ent.Taxonomy) []*structs.ReadTaxonomy {
	var rs []*structs.ReadTaxonomy
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
		ParentID:    row.ParentID,
		TenantID:    row.ParentID,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}
}
