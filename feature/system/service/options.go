package service

import (
	"context"
	"errors"
	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/paging"
	"ncobase/common/validator"
	"ncobase/feature/system/data"
	"ncobase/feature/system/data/ent"
	"ncobase/feature/system/data/repository"
	"ncobase/feature/system/structs"
)

// OptionsServiceInterface represents the options service interface.
type OptionsServiceInterface interface {
	Create(ctx context.Context, body *structs.OptionsBody) (*structs.ReadOptions, error)
	Update(ctx context.Context, updates *structs.UpdateOptionsBody) (*structs.ReadOptions, error)
	Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOptions, error)
	Delete(ctx context.Context, params *structs.FindOptions) error
	List(ctx context.Context, params *structs.ListOptionsParams) (paging.Result[*structs.ReadOptions], error)
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

// Update updates an existing option (full and partial).
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

// Get retrieves an option by ID or name.
func (s *optionsService) Get(ctx context.Context, params *structs.FindOptions) (*structs.ReadOptions, error) {
	row, err := s.options.Get(ctx, params)
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
			log.Errorf(ctx, "Error listing options: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
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
		TenantID:  row.TenantID,
		CreatedBy: &row.CreatedBy,
		CreatedAt: &row.CreatedAt,
		UpdatedBy: &row.UpdatedBy,
		UpdatedAt: &row.UpdatedAt,
	}
}
