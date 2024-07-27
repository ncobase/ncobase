package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository"
	"ncobase/feature/content/structs"
)

// TaxonomyRelationServiceInterface is the interface for the service.
type TaxonomyRelationServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error)
	Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error)
	Get(ctx context.Context, object string) (*structs.ReadTaxonomyRelation, error)
	List(ctx context.Context, params *structs.ListTaxonomyRelationParams) (paging.Result[*structs.ReadTaxonomyRelation], error)
	Delete(ctx context.Context, object string) error
}

// taxonomyRelationService is the struct for the service.
type taxonomyRelationService struct {
	repo *repository.Repository
}

// NewTaxonomyRelationService creates a new service.
func NewTaxonomyRelationService(repo *repository.Repository) TaxonomyRelationServiceInterface {
	return &taxonomyRelationService{
		repo: repo,
	}
}

// Create creates a new taxonomy relation.
func (s *taxonomyRelationService) Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.repo.TaxonomyRelations.Create(ctx, body)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing taxonomy relation.
func (s *taxonomyRelationService) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.repo.TaxonomyRelations.Update(ctx, body)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy relation by ID.
func (s *taxonomyRelationService) Get(ctx context.Context, object string) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.repo.TaxonomyRelations.GetByObject(ctx, object)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy relation by ID.
func (s *taxonomyRelationService) Delete(ctx context.Context, object string) error {
	err := s.repo.TaxonomyRelations.Delete(ctx, object)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return err
	}
	return nil
}

// List lists all taxonomy relations.
func (s *taxonomyRelationService) List(ctx context.Context, params *structs.ListTaxonomyRelationParams) (paging.Result[*structs.ReadTaxonomyRelation], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTaxonomyRelation, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.repo.TaxonomyRelations.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			log.Errorf(ctx, "Error listing taxonomy relations: %v\n", err)
			return nil, 0, err
		}

		total := s.repo.TaxonomyRelations.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes serializes taxonomy relations.
func (s *taxonomyRelationService) Serializes(rows []*ent.TaxonomyRelation) []*structs.ReadTaxonomyRelation {
	var rs []*structs.ReadTaxonomyRelation
	for _, row := range rows {
		rs = append(rs, s.Serialize(row))
	}
	return rs
}

// Serialize serializes a taxonomy relation.
func (s *taxonomyRelationService) Serialize(row *ent.TaxonomyRelation) *structs.ReadTaxonomyRelation {
	return &structs.ReadTaxonomyRelation{
		ID:         row.ID,
		ObjectID:   row.ObjectID,
		TaxonomyID: row.TaxonomyID,
		Type:       row.Type,
		Order:      &row.Order,
		CreatedBy:  &row.CreatedBy,
		CreatedAt:  &row.CreatedAt,
	}
}
