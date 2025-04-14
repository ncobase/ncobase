package service

import (
	"context"
	"errors"
	"ncobase/core/system/data"
	"ncobase/core/system/data/ent"
	"ncobase/core/system/data/repository"
	"ncobase/core/system/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// DictionaryServiceInterface represents the dictionary service interface.
type DictionaryServiceInterface interface {
	Create(ctx context.Context, body *structs.DictionaryBody) (*structs.ReadDictionary, error)
	Update(ctx context.Context, updates *structs.UpdateDictionaryBody) (*structs.ReadDictionary, error)
	Get(ctx context.Context, params *structs.FindDictionary) (any, error)
	Delete(ctx context.Context, params *structs.FindDictionary) (*structs.ReadDictionary, error)
	List(ctx context.Context, params *structs.ListDictionaryParams) (paging.Result[*structs.ReadDictionary], error)
}

// DictionaryService represents the dictionary service.
type dictionaryService struct {
	dictionary repository.DictionaryRepositoryInterface
}

// NewDictionaryService creates a new dictionary service.
func NewDictionaryService(d *data.Data) DictionaryServiceInterface {
	return &dictionaryService{
		dictionary: repository.NewDictionaryRepository(d),
	}
}

// Create creates a new dictionary.
func (s *dictionaryService) Create(ctx context.Context, body *structs.DictionaryBody) (*structs.ReadDictionary, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := s.dictionary.Create(ctx, body)
	if err := handleEntError(ctx, "Dictionary", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing dictionary (full and partial).
func (s *dictionaryService) Update(ctx context.Context, updates *structs.UpdateDictionaryBody) (*structs.ReadDictionary, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := s.dictionary.Update(ctx, updates)
	if err := handleEntError(ctx, "Dictionary", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a dictionary by ID.
func (s *dictionaryService) Get(ctx context.Context, params *structs.FindDictionary) (any, error) {
	row, err := s.dictionary.Get(ctx, params)
	if err := handleEntError(ctx, "Dictionary", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a dictionary by ID.
func (s *dictionaryService) Delete(ctx context.Context, params *structs.FindDictionary) (*structs.ReadDictionary, error) {
	err := s.dictionary.Delete(ctx, params)
	if err := handleEntError(ctx, "Dictionary", err); err != nil {
		return nil, err
	}

	return nil, nil
}

// List lists all dictionarys.
func (s *dictionaryService) List(ctx context.Context, params *structs.ListDictionaryParams) (paging.Result[*structs.ReadDictionary], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadDictionary, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, err := s.dictionary.List(ctx, &lp)
		if ent.IsNotFound(err) {
			return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
		}
		if err != nil {
			logger.Errorf(ctx, "Error listing dictionarys: %v", err)
			return nil, 0, err
		}

		total := s.dictionary.CountX(ctx, params)

		return s.Serializes(rows), total, nil
	})
}

// Serializes dictionarys.
func (s *dictionaryService) Serializes(rows []*ent.Dictionary) []*structs.ReadDictionary {
	rs := make([]*structs.ReadDictionary, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes a dictionary.
func (s *dictionaryService) Serialize(row *ent.Dictionary) *structs.ReadDictionary {
	return &structs.ReadDictionary{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Type:      row.Type,
		Value:     row.Value,
		TenantID:  row.TenantID,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
