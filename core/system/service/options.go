package service

import (
	"context"
	"encoding/json"
	"errors"
	"ncobase/system/data"
	"ncobase/system/data/ent"
	"ncobase/system/data/repository"
	"ncobase/system/structs"
	"strconv"
	"strings"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/validation/validator"
)

// OptionsServiceInterface represents the options service interface.
type OptionsServiceInterface interface {
	Create(ctx context.Context, body *structs.OptionsBody) (*structs.ReadOptions, error)
	Update(ctx context.Context, updates *structs.UpdateOptionsBody) (*structs.ReadOptions, error)
	Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOptions, error)
	GetByName(ctx context.Context, name string) (*structs.ReadOptions, error)
	GetByType(ctx context.Context, typeName string) ([]*structs.ReadOptions, error)
	GetValueByName(ctx context.Context, name string) (any, error)
	ParseValue(option *structs.ReadOptions) (any, error)
	GetObjectByName(ctx context.Context, name string) (map[string]any, error)
	GetArrayByName(ctx context.Context, name string) ([]any, error)
	GetStringByName(ctx context.Context, name string) (string, error)
	GetBoolByName(ctx context.Context, name string, defaultValue bool) (bool, error)
	GetNumberByName(ctx context.Context, name string, defaultValue float64) (float64, error)
	BatchGetByNames(ctx context.Context, names []string) (map[string]*structs.ReadOptions, error)
	Delete(ctx context.Context, params *structs.FindOptions) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	List(ctx context.Context, params *structs.ListOptionsParams) (paging.Result[*structs.ReadOptions], error)
	CountX(ctx context.Context, params *structs.ListOptionsParams) int
}

// OptionsService represents the options service.
type optionsService struct {
	options repository.OptionsRepositoryInterface
}

// NewOptionsService creates a new options service.
func NewOptionsService(d *data.Data) OptionsServiceInterface {
	return &optionsService{
		options: repository.NewOptionsRepository(d),
	}
}

// Create creates a new option.
func (s *optionsService) Create(ctx context.Context, body *structs.OptionsBody) (*structs.ReadOptions, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := s.options.Create(ctx, body)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves an option by ID or name.
func (s *optionsService) Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOptions, error) {
	row, err := s.options.Get(ctx, params)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByName retrieves an option by name.
func (s *optionsService) GetByName(ctx context.Context, name string) (*structs.ReadOptions, error) {
	if validator.IsEmpty(name) {
		return nil, errors.New(ecode.FieldIsRequired("name"))
	}

	params := &structs.FindOptions{
		Option: name,
	}

	row, err := s.options.Get(ctx, params)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByType retrieves options by type.
func (s *optionsService) GetByType(ctx context.Context, typeName string) ([]*structs.ReadOptions, error) {
	if validator.IsEmpty(typeName) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	params := &structs.ListOptionsParams{
		Type: typeName,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetValueByName retrieves the option value by name and parses it according to its type.
func (s *optionsService) GetValueByName(ctx context.Context, name string) (any, error) {
	option, err := s.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return s.ParseValue(option)
}

// ParseValue parses an option's value according to its type.
func (s *optionsService) ParseValue(option *structs.ReadOptions) (any, error) {
	switch option.Type {
	case "string":
		return option.Value, nil
	case "boolean", "bool":
		value := strings.ToLower(option.Value)
		return value == "true" || value == "1" || value == "yes", nil
	case "number", "int", "float":
		if strings.Contains(option.Value, ".") {
			return strconv.ParseFloat(option.Value, 64)
		}
		return strconv.ParseInt(option.Value, 10, 64)
	case "object", "json":
		var result any
		err := json.Unmarshal([]byte(option.Value), &result)
		return result, err
	case "array":
		var result []any
		err := json.Unmarshal([]byte(option.Value), &result)
		return result, err
	default:
		return option.Value, nil
	}
}

// GetObjectByName retrieves and parses an option of object type by name.
func (s *optionsService) GetObjectByName(ctx context.Context, name string) (map[string]any, error) {
	value, err := s.GetValueByName(ctx, name)
	if err != nil {
		return nil, err
	}

	obj, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New(ecode.TypeMismatch("object"))
	}

	return obj, nil
}

// GetArrayByName retrieves and parses an option of array type by name.
func (s *optionsService) GetArrayByName(ctx context.Context, name string) ([]any, error) {
	value, err := s.GetValueByName(ctx, name)
	if err != nil {
		return nil, err
	}

	arr, ok := value.([]any)
	if !ok {
		return nil, errors.New(ecode.TypeMismatch("array"))
	}

	return arr, nil
}

// GetStringByName retrieves an option as string by name.
func (s *optionsService) GetStringByName(ctx context.Context, name string) (string, error) {
	option, err := s.GetByName(ctx, name)
	if err != nil {
		return "", err
	}

	if option.Type != "string" {
		logger.Warnf(ctx, "Option %s is not of string type, attempting conversion", name)
	}

	return option.Value, nil
}

// GetBoolByName retrieves an option as boolean by name.
func (s *optionsService) GetBoolByName(ctx context.Context, name string, defaultValue bool) (bool, error) {
	value, err := s.GetValueByName(ctx, name)
	if err != nil {
		return defaultValue, err
	}

	boolValue, ok := value.(bool)
	if !ok {
		return defaultValue, errors.New(ecode.TypeMismatch("boolean"))
	}

	return boolValue, nil
}

// GetNumberByName retrieves an option as number by name.
func (s *optionsService) GetNumberByName(ctx context.Context, name string, defaultValue float64) (float64, error) {
	value, err := s.GetValueByName(ctx, name)
	if err != nil {
		return defaultValue, err
	}

	switch v := value.(type) {
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return defaultValue, errors.New(ecode.TypeMismatch("number"))
	}
}

// BatchGetByNames retrieves multiple options by their names.
func (s *optionsService) BatchGetByNames(ctx context.Context, names []string) (map[string]*structs.ReadOptions, error) {
	result := make(map[string]*structs.ReadOptions)

	for _, name := range names {
		option, err := s.GetByName(ctx, name)
		if err != nil {
			logger.Warnf(ctx, "Failed to get option %s: %v", name, err)
			continue
		}
		result[name] = option
	}

	return result, nil
}

// Update updates an existing option.
func (s *optionsService) Update(ctx context.Context, updates *structs.UpdateOptionsBody) (*structs.ReadOptions, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := s.options.Update(ctx, updates)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes an option by ID or name.
func (s *optionsService) Delete(ctx context.Context, params *structs.FindOptions) error {
	err := s.options.Delete(ctx, params)
	return handleEntError(ctx, "Options", err)
}

// DeleteByPrefix deletes options by prefix.
func (s *optionsService) DeleteByPrefix(ctx context.Context, prefix string) error {
	if validator.IsEmpty(prefix) {
		return errors.New(ecode.FieldIsRequired("prefix"))
	}

	return s.options.DeleteByPrefix(ctx, prefix)
}

// List lists all options.
func (s *optionsService) List(ctx context.Context, params *structs.ListOptionsParams) (paging.Result[*structs.ReadOptions], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadOptions, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.options.ListWithCount(ctx, &lp)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, 0, errors.New(ecode.FieldIsInvalid("cursor"))
			}
			logger.Errorf(ctx, "Error listing options: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// CountX counts all options.
func (s *optionsService) CountX(ctx context.Context, params *structs.ListOptionsParams) int {
	return s.options.CountX(ctx, params)
}

// Serializes options.
func (s *optionsService) Serializes(rows []*ent.Options) []*structs.ReadOptions {
	rs := make([]*structs.ReadOptions, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes an option.
func (s *optionsService) Serialize(row *ent.Options) *structs.ReadOptions {
	return &structs.ReadOptions{
		ID:        row.ID,
		Name:      row.Name,
		Type:      row.Type,
		Value:     row.Value,
		Autoload:  row.Autoload,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
