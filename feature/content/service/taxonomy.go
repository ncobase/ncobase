package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
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
	List(ctx context.Context, params *structs.ListTaxonomyParams) (*types.JSON, error)
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
func (s *taxonomyService) List(ctx context.Context, params *structs.ListTaxonomyParams) (*types.JSON, error) {
	// limit default value
	if validator.IsEmpty(params.Limit) {
		params.Limit = 20
	}
	// limit must less than 100
	if params.Limit > 100 {
		return nil, errors.New(ecode.FieldIsInvalid("limit"))
	}

	rows, err := s.taxonomy.List(ctx, params)

	if ent.IsNotFound(err) {
		return nil, errors.New(ecode.FieldIsInvalid("cursor"))
	}
	if validator.IsNotNil(err) {
		return nil, err
	}

	total := s.taxonomy.CountX(ctx, params)

	return &types.JSON{
		"content": s.Serializes(rows),
		"total":   total,
	}, nil
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
