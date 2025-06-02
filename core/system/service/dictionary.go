package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/system/data"
	"ncobase/system/data/ent"
	"ncobase/system/data/repository"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// DictionaryServiceInterface represents the dictionary service interface.
type DictionaryServiceInterface interface {
	Create(ctx context.Context, body *structs.DictionaryBody) (*structs.ReadDictionary, error)
	Update(ctx context.Context, updates *structs.UpdateDictionaryBody) (*structs.ReadDictionary, error)
	Get(ctx context.Context, params *structs.FindDictionary) (*structs.ReadDictionary, error)
	GetByType(ctx context.Context, typeName string) ([]*structs.ReadDictionary, error)
	GetBySlug(ctx context.Context, slug string) (*structs.ReadDictionary, error)
	GetValueBySlug(ctx context.Context, slug string) (any, error)
	GetObjectBySlug(ctx context.Context, slug string) (map[string]any, error)
	GetEnumBySlug(ctx context.Context, slug string) (map[string]string, error)
	GetEnumOptions(ctx context.Context, slug string) ([]map[string]any, error)
	GetNestedEnumBySlug(ctx context.Context, slug string) (map[string]map[string]any, error)
	ValidateEnumValue(ctx context.Context, slug string, value string) (bool, error)
	GetEnumValueLabel(ctx context.Context, slug string, value string) (string, error)
	BatchGetBySlug(ctx context.Context, slugs []string) (map[string]*structs.ReadDictionary, error)
	Delete(ctx context.Context, params *structs.FindDictionary) (*structs.ReadDictionary, error)
	List(ctx context.Context, params *structs.ListDictionaryParams) (paging.Result[*structs.ReadDictionary], error)
	CountX(ctx context.Context, params *structs.ListDictionaryParams) int
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

// Get retrieves a dictionary by ID.
func (s *dictionaryService) Get(ctx context.Context, params *structs.FindDictionary) (*structs.ReadDictionary, error) {
	row, err := s.dictionary.Get(ctx, params)
	if err := handleEntError(ctx, "Dictionary", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByType retrieves dictionaries by type.
func (s *dictionaryService) GetByType(ctx context.Context, typeName string) ([]*structs.ReadDictionary, error) {
	if validator.IsEmpty(typeName) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	params := &structs.ListDictionaryParams{
		Type: typeName,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetBySlug retrieves a dictionary by its slug.
func (s *dictionaryService) GetBySlug(ctx context.Context, slug string) (*structs.ReadDictionary, error) {
	if validator.IsEmpty(slug) {
		return nil, errors.New(ecode.FieldIsRequired("slug"))
	}

	params := &structs.FindDictionary{
		Dictionary: slug,
	}

	row, err := s.Get(ctx, params)
	if err != nil {
		return nil, err
	}

	return row, nil
}

// GetValueBySlug retrieves and parses a dictionary's value by slug.
func (s *dictionaryService) GetValueBySlug(ctx context.Context, slug string) (any, error) {
	dict, err := s.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return dict.ParseValue()
}

// GetObjectBySlug retrieves a dictionary value as object/map by slug.
func (s *dictionaryService) GetObjectBySlug(ctx context.Context, slug string) (map[string]any, error) {
	value, err := s.GetValueBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New(ecode.TypeMismatch("object"))
	}

	return obj, nil
}

// GetEnumBySlug retrieves a dictionary as enum (simple key-value object) by slug.
func (s *dictionaryService) GetEnumBySlug(ctx context.Context, slug string) (map[string]string, error) {
	obj, err := s.GetObjectBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, value := range obj {
		strValue, ok := value.(string)
		if !ok {
			continue // Skip non-string values
		}
		result[key] = strValue
	}

	return result, nil
}

// GetEnumOptions retrieves a dictionary as options array for UI select components.
func (s *dictionaryService) GetEnumOptions(ctx context.Context, slug string) ([]map[string]any, error) {
	enum, err := s.GetEnumBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	options := make([]map[string]any, 0, len(enum))
	for key, value := range enum {
		options = append(options, map[string]any{
			"value": key,
			"label": value,
		})
	}

	return options, nil
}

// GetNestedEnumBySlug retrieves a dictionary as nested enum (complex object with sub-properties) by slug.
func (s *dictionaryService) GetNestedEnumBySlug(ctx context.Context, slug string) (map[string]map[string]any, error) {
	obj, err := s.GetObjectBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]any)
	for key, value := range obj {
		nestedObj, ok := value.(map[string]any)
		if !ok {
			continue // Skip non-object values
		}
		result[key] = nestedObj
	}

	return result, nil
}

// ValidateEnumValue checks if a value is valid for a given dictionary enum.
func (s *dictionaryService) ValidateEnumValue(ctx context.Context, slug string, value string) (bool, error) {
	enum, err := s.GetEnumBySlug(ctx, slug)
	if err != nil {
		return false, err
	}

	_, exists := enum[value]
	return exists, nil
}

// GetEnumValueLabel gets the label for a value in a dictionary enum.
func (s *dictionaryService) GetEnumValueLabel(ctx context.Context, slug string, value string) (string, error) {
	enum, err := s.GetEnumBySlug(ctx, slug)
	if err != nil {
		return "", err
	}

	label, exists := enum[value]
	if !exists {
		return "", errors.New(ecode.NotExist(fmt.Sprintf("Value '%s' in dictionary '%s'", value, slug)))
	}

	return label, nil
}

// BatchGetBySlug retrieves multiple dictionaries by their slugs.
func (s *dictionaryService) BatchGetBySlug(ctx context.Context, slugs []string) (map[string]*structs.ReadDictionary, error) {
	result := make(map[string]*structs.ReadDictionary)

	for _, slug := range slugs {
		dict, err := s.GetBySlug(ctx, slug)
		if err != nil {
			logger.Warnf(ctx, "Failed to get dictionary %s: %v", slug, err)
			continue
		}
		result[slug] = dict
	}

	return result, nil
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

// CountX counts dictionarys.
func (s *dictionaryService) CountX(ctx context.Context, params *structs.ListDictionaryParams) int {
	return s.dictionary.CountX(ctx, params)
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
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
