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

// OptionServiceInterface represents the option service interface.
type OptionServiceInterface interface {
	Create(ctx context.Context, body *structs.OptionBody) (*structs.ReadOption, error)
	Update(ctx context.Context, updates *structs.UpdateOptionBody) (*structs.ReadOption, error)
	Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOption, error)
	GetByName(ctx context.Context, name string) (*structs.ReadOption, error)
	GetByType(ctx context.Context, typeName string) ([]*structs.ReadOption, error)
	GetValueByName(ctx context.Context, name string) (any, error)
	ParseValue(option *structs.ReadOption) (any, error)
	GetObjectByName(ctx context.Context, name string) (map[string]any, error)
	GetArrayByName(ctx context.Context, name string) ([]any, error)
	GetStringByName(ctx context.Context, name string) (string, error)
	GetBoolByName(ctx context.Context, name string, defaultValue bool) (bool, error)
	GetNumberByName(ctx context.Context, name string, defaultValue float64) (float64, error)
	BatchGetByNames(ctx context.Context, names []string) (map[string]*structs.ReadOption, error)
	Delete(ctx context.Context, params *structs.FindOptions) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	List(ctx context.Context, params *structs.ListOptionParams) (paging.Result[*structs.ReadOption], error)
	CountX(ctx context.Context, params *structs.ListOptionParams) int
}

// OptionService represents the option service.
type optionService struct {
	option repository.OptionRepositoryInterface
}

// NewOptionService creates a new option service.
func NewOptionService(d *data.Data) OptionServiceInterface {
	return &optionService{
		option: repository.NewOptionRepository(d),
	}
}

// Create creates a new option.
func (s *optionService) Create(ctx context.Context, body *structs.OptionBody) (*structs.ReadOption, error) {
	if validator.IsEmpty(body.Name) {
		return nil, errors.New(ecode.FieldIsInvalid("name"))
	}

	row, err := s.option.Create(ctx, body)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves an option by ID or name.
func (s *optionService) Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOption, error) {
	row, err := s.option.Get(ctx, params)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByName retrieves an option by name.
func (s *optionService) GetByName(ctx context.Context, name string) (*structs.ReadOption, error) {
	if validator.IsEmpty(name) {
		return nil, errors.New(ecode.FieldIsRequired("name"))
	}

	params := &structs.FindOptions{
		Option: name,
	}

	row, err := s.option.Get(ctx, params)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByType retrieves options by type.
func (s *optionService) GetByType(ctx context.Context, typeName string) ([]*structs.ReadOption, error) {
	if validator.IsEmpty(typeName) {
		return nil, errors.New(ecode.FieldIsRequired("type"))
	}

	params := &structs.ListOptionParams{
		Type: typeName,
	}

	result, err := s.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

// GetValueByName retrieves the option value by name and parses it according to its type.
func (s *optionService) GetValueByName(ctx context.Context, name string) (any, error) {
	option, err := s.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return s.ParseValue(option)
}

// ParseValue parses an option's value according to its type.
func (s *optionService) ParseValue(option *structs.ReadOption) (any, error) {
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
func (s *optionService) GetObjectByName(ctx context.Context, name string) (map[string]any, error) {
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
func (s *optionService) GetArrayByName(ctx context.Context, name string) ([]any, error) {
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
func (s *optionService) GetStringByName(ctx context.Context, name string) (string, error) {
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
func (s *optionService) GetBoolByName(ctx context.Context, name string, defaultValue bool) (bool, error) {
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
func (s *optionService) GetNumberByName(ctx context.Context, name string, defaultValue float64) (float64, error) {
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
func (s *optionService) BatchGetByNames(ctx context.Context, names []string) (map[string]*structs.ReadOption, error) {
	result := make(map[string]*structs.ReadOption)

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
func (s *optionService) Update(ctx context.Context, updates *structs.UpdateOptionBody) (*structs.ReadOption, error) {
	if validator.IsEmpty(updates.ID) {
		return nil, errors.New(ecode.FieldIsRequired("id"))
	}

	row, err := s.option.Update(ctx, updates)
	if err := handleEntError(ctx, "Options", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes an option by ID or name.
func (s *optionService) Delete(ctx context.Context, params *structs.FindOptions) error {
	err := s.option.Delete(ctx, params)
	return handleEntError(ctx, "Options", err)
}

// DeleteByPrefix deletes options by prefix.
func (s *optionService) DeleteByPrefix(ctx context.Context, prefix string) error {
	if validator.IsEmpty(prefix) {
		return errors.New(ecode.FieldIsRequired("prefix"))
	}

	return s.option.DeleteByPrefix(ctx, prefix)
}

// List lists all options.
func (s *optionService) List(ctx context.Context, params *structs.ListOptionParams) (paging.Result[*structs.ReadOption], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadOption, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.option.ListWithCount(ctx, &lp)
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
func (s *optionService) CountX(ctx context.Context, params *structs.ListOptionParams) int {
	return s.option.CountX(ctx, params)
}

// Serializes options.
func (s *optionService) Serializes(rows []*ent.Options) []*structs.ReadOption {
	rs := make([]*structs.ReadOption, len(rows))
	for i, row := range rows {
		rs[i] = s.Serialize(row)
	}
	return rs
}

// Serialize serializes an option.
func (s *optionService) Serialize(row *ent.Options) *structs.ReadOption {
	return &structs.ReadOption{
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
