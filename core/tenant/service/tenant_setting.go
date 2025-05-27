package service

import (
	"context"
	"errors"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// TenantSettingServiceInterface defines the interface for tenant setting service
type TenantSettingServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*structs.ReadTenantSetting, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantSetting, error)
	Get(ctx context.Context, id string) (*structs.ReadTenantSetting, error)
	GetByKey(ctx context.Context, tenantID, key string) (*structs.ReadTenantSetting, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantSettingParams) (paging.Result[*structs.ReadTenantSetting], error)
	BulkUpdate(ctx context.Context, req *structs.BulkUpdateSettingsRequest) error
	GetTenantSettings(ctx context.Context, tenantID string, publicOnly bool) (map[string]interface{}, error)
	SetSetting(ctx context.Context, tenantID, key, value string) error
	GetSettingValue(ctx context.Context, tenantID, key string) (interface{}, error)
	Serialize(row *ent.TenantSetting) *structs.ReadTenantSetting
	Serializes(rows []*ent.TenantSetting) []*structs.ReadTenantSetting
}

// tenantSettingService implements TenantSettingServiceInterface
type tenantSettingService struct {
	repo repository.TenantSettingRepositoryInterface
}

// NewTenantSettingService creates a new tenant setting service
func NewTenantSettingService(d *data.Data) TenantSettingServiceInterface {
	return &tenantSettingService{
		repo: repository.NewTenantSettingRepository(d),
	}
}

// Create creates a new tenant setting
func (s *tenantSettingService) Create(ctx context.Context, body *structs.CreateTenantSettingBody) (*structs.ReadTenantSetting, error) {
	if body.TenantID == "" {
		return nil, errors.New(ecode.FieldIsRequired("tenant_id"))
	}
	if body.SettingKey == "" {
		return nil, errors.New(ecode.FieldIsRequired("setting_key"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "TenantSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing tenant setting
func (s *tenantSettingService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantSetting, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "TenantSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a tenant setting by ID
func (s *tenantSettingService) Get(ctx context.Context, id string) (*structs.ReadTenantSetting, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "TenantSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByKey retrieves a tenant setting by tenant ID and key
func (s *tenantSettingService) GetByKey(ctx context.Context, tenantID, key string) (*structs.ReadTenantSetting, error) {
	row, err := s.repo.GetByKey(ctx, tenantID, key)
	if err := handleEntError(ctx, "TenantSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a tenant setting
func (s *tenantSettingService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "TenantSetting", err); err != nil {
		return err
	}
	return nil
}

// List lists tenant settings with pagination
func (s *tenantSettingService) List(ctx context.Context, params *structs.ListTenantSettingParams) (paging.Result[*structs.ReadTenantSetting], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTenantSetting, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing tenant settings: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// BulkUpdate updates multiple settings for a tenant
func (s *tenantSettingService) BulkUpdate(ctx context.Context, req *structs.BulkUpdateSettingsRequest) error {
	for key, value := range req.Settings {
		if err := s.SetSetting(ctx, req.TenantID, key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetTenantSettings retrieves all settings for a tenant as key-value map
func (s *tenantSettingService) GetTenantSettings(ctx context.Context, tenantID string, publicOnly bool) (map[string]interface{}, error) {
	params := &structs.ListTenantSettingParams{
		TenantID: tenantID,
		Limit:    1000, // Get all settings
	}
	if publicOnly {
		isPublic := true
		params.IsPublic = &isPublic
	}

	rows, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, row := range rows {
		setting := s.Serialize(row)
		result[setting.SettingKey] = setting.GetTypedValue()
	}

	return result, nil
}

// SetSetting creates or updates a setting
func (s *tenantSettingService) SetSetting(ctx context.Context, tenantID, key, value string) error {
	existing, err := s.repo.GetByKey(ctx, tenantID, key)
	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if existing != nil {
		// Update existing setting
		updates := types.JSON{
			"setting_value": value,
		}
		_, err = s.repo.Update(ctx, existing.ID, updates)
		return err
	}

	// Create new setting
	body := &structs.CreateTenantSettingBody{
		TenantSettingBody: structs.TenantSettingBody{
			TenantID:     tenantID,
			SettingKey:   key,
			SettingValue: value,
			SettingName:  key, // Use key as name for simplicity
			SettingType:  structs.TypeString,
			Scope:        structs.ScopeTenant,
		},
	}
	_, err = s.repo.Create(ctx, body)
	return err
}

// GetSettingValue retrieves a setting value with type conversion
func (s *tenantSettingService) GetSettingValue(ctx context.Context, tenantID, key string) (interface{}, error) {
	setting, err := s.GetByKey(ctx, tenantID, key)
	if err != nil {
		return nil, err
	}

	return setting.GetTypedValue(), nil
}

// Serialize converts entity to struct
func (s *tenantSettingService) Serialize(row *ent.TenantSetting) *structs.ReadTenantSetting {
	return &structs.ReadTenantSetting{
		ID:           row.ID,
		TenantID:     row.TenantID,
		SettingKey:   row.SettingKey,
		SettingName:  row.SettingName,
		SettingValue: row.SettingValue,
		DefaultValue: row.DefaultValue,
		SettingType:  structs.SettingType(row.SettingType),
		Scope:        structs.SettingScope(row.Scope),
		Category:     row.Category,
		Description:  row.Description,
		IsPublic:     row.IsPublic,
		IsRequired:   row.IsRequired,
		IsReadonly:   row.IsReadonly,
		Validation:   &row.Validation,
		Extras:       &row.Extras,
		CreatedBy:    &row.CreatedBy,
		CreatedAt:    &row.CreatedAt,
		UpdatedBy:    &row.UpdatedBy,
		UpdatedAt:    &row.UpdatedAt,
	}
}

// Serializes converts multiple entities to structs
func (s *tenantSettingService) Serializes(rows []*ent.TenantSetting) []*structs.ReadTenantSetting {
	result := make([]*structs.ReadTenantSetting, len(rows))
	for i, row := range rows {
		result[i] = s.Serialize(row)
	}
	return result
}
