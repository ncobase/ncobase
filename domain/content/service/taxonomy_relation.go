package service

import (
	"context"
	"errors"
	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/logger"
	"github.com/ncobase/ncore/pkg/paging"
	"ncobase/domain/content/data"
	"ncobase/domain/content/data/ent"
	"ncobase/domain/content/data/repository"
	"ncobase/domain/content/structs"
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
	r repository.TaxonomyRelationsRepositoryInterface
}

// NewTaxonomyRelationService creates a new service.
func NewTaxonomyRelationService(d *data.Data) TaxonomyRelationServiceInterface {
	return &taxonomyRelationService{
		r: repository.NewTaxonomyRelationsRepository(d),
	}
}

// Create creates a new taxonomy relation.
func (s *taxonomyRelationService) Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.r.Create(ctx, body)
	if err := handleEntError(ctx, "Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing taxonomy relation.
func (s *taxonomyRelationService) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.r.Update(ctx, body)
	if err := handleEntError(ctx, "Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy relation by ID.
func (s *taxonomyRelationService) Get(ctx context.Context, object string) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.r.GetByObject(ctx, object)
	if err := handleEntError(ctx, "Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy relation by ID.
func (s *taxonomyRelationService) Delete(ctx context.Context, object string) error {
	err := s.r.Delete(ctx, object)
	if err := handleEntError(ctx, "Taxonomy relation", err); err != nil {
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

		rows, err := s.r.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing taxonomy relations: %v", err)
			return nil, 0, err
		}

		total := s.r.CountX(ctx, params)

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
