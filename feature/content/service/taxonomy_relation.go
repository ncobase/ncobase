package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/feature/content/data"
	"ncobase/feature/content/data/ent"
	"ncobase/feature/content/data/repository"
	"ncobase/feature/content/structs"
)

// TaxonomyRelationServiceInterface is the interface for the service.
type TaxonomyRelationServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error)
	Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error)
	Get(ctx context.Context, object string) (*structs.ReadTaxonomyRelation, error)
	List(ctx context.Context, params *structs.ListTaxonomyRelationParams) (*paging.Result[*structs.ReadTaxonomyRelation], error)
	Delete(ctx context.Context, object string) error
}

// taxonomyRelationService is the struct for the service.
type taxonomyRelationService struct {
	taxonomyRelations repository.TaxonomyRelationsRepositoryInterface
}

// NewTaxonomyRelationService creates a new service.
func NewTaxonomyRelationService(d *data.Data) TaxonomyRelationServiceInterface {
	return &taxonomyRelationService{
		taxonomyRelations: repository.NewTaxonomyRelationsRepository(d),
	}
}

// Create creates a new taxonomy relation.
func (s *taxonomyRelationService) Create(ctx context.Context, body *structs.CreateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.taxonomyRelations.Create(ctx, body)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing taxonomy relation.
func (s *taxonomyRelationService) Update(ctx context.Context, body *structs.UpdateTaxonomyRelationBody) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.taxonomyRelations.Update(ctx, body)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a taxonomy relation by ID.
func (s *taxonomyRelationService) Get(ctx context.Context, object string) (*structs.ReadTaxonomyRelation, error) {
	row, err := s.taxonomyRelations.GetByObject(ctx, object)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a taxonomy relation by ID.
func (s *taxonomyRelationService) Delete(ctx context.Context, object string) error {
	err := s.taxonomyRelations.Delete(ctx, object)
	if err := handleEntError("Taxonomy relation", err); err != nil {
		return err
	}
	return nil
}

// List lists all taxonomy relations.
func (s *taxonomyRelationService) List(ctx context.Context, params *structs.ListTaxonomyRelationParams) (*paging.Result[*structs.ReadTaxonomyRelation], error) {
	pp := paging.Params{
		Cursor: params.Cursor,
		Limit:  params.Limit,
	}

	return paging.Paginate(pp, func(cursor string, limit int) ([]*structs.ReadTaxonomyRelation, int, string, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit

		rows, err := s.taxonomyRelations.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, "", errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if validator.IsNotNil(err) {
			log.Errorf(ctx, "Error listing taxonomy relations: %v\n", err)
			return nil, 0, "", err
		}
		if err != nil {
			return nil, 0, "", err
		}

		total := s.taxonomyRelations.CountX(ctx, params)

		var nextCursor string
		if len(rows) > 0 {
			nextCursor = paging.EncodeCursor(rows[len(rows)-1].CreatedAt)
		}

		return s.Serializes(rows), total, nextCursor, nil
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
