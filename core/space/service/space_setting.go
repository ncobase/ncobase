package service

import (
	"context"
	"errors"
	"ncobase/space/data"
	"ncobase/space/data/ent"
	"ncobase/space/data/repository"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// SpaceSettingServiceInterface defines the interface for space setting service
type SpaceSettingServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceSettingBody) (*structs.ReadSpaceSetting, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceSetting, error)
	Get(ctx context.Context, id string) (*structs.ReadSpaceSetting, error)
	GetByKey(ctx context.Context, spaceID, key string) (*structs.ReadSpaceSetting, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceSettingParams) (paging.Result[*structs.ReadSpaceSetting], error)
	BulkUpdate(ctx context.Context, req *structs.BulkUpdateSettingsRequest) error
	GetSpaceSettings(ctx context.Context, spaceID string, publicOnly bool) (map[string]any, error)
	SetSetting(ctx context.Context, spaceID, key, value string) error
	GetSettingValue(ctx context.Context, spaceID, key string) (any, error)
	Serialize(row *ent.SpaceSetting) *structs.ReadSpaceSetting
	Serializes(rows []*ent.SpaceSetting) []*structs.ReadSpaceSetting
}

// spaceSettingService implements SpaceSettingServiceInterface
type spaceSettingService struct {
	repo repository.SpaceSettingRepositoryInterface
}

// NewSpaceSettingService creates a new space setting service
func NewSpaceSettingService(d *data.Data) SpaceSettingServiceInterface {
	return &spaceSettingService{
		repo: repository.NewSpaceSettingRepository(d),
	}
}

// Create creates a new space setting
func (s *spaceSettingService) Create(ctx context.Context, body *structs.CreateSpaceSettingBody) (*structs.ReadSpaceSetting, error) {
	if body.SpaceID == "" {
		return nil, errors.New(ecode.FieldIsRequired("space_id"))
	}
	if body.SettingKey == "" {
		return nil, errors.New(ecode.FieldIsRequired("setting_key"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "SpaceSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing space setting
func (s *spaceSettingService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceSetting, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "SpaceSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a space setting by ID
func (s *spaceSettingService) Get(ctx context.Context, id string) (*structs.ReadSpaceSetting, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "SpaceSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByKey retrieves a space setting by space ID and key
func (s *spaceSettingService) GetByKey(ctx context.Context, spaceID, key string) (*structs.ReadSpaceSetting, error) {
	row, err := s.repo.GetByKey(ctx, spaceID, key)
	if err := handleEntError(ctx, "SpaceSetting", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a space setting
func (s *spaceSettingService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "SpaceSetting", err); err != nil {
		return err
	}
	return nil
}

// List lists space settings
func (s *spaceSettingService) List(ctx context.Context, params *structs.ListSpaceSettingParams) (paging.Result[*structs.ReadSpaceSetting], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadSpaceSetting, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing space settings: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// BulkUpdate updates multiple settings for a space
func (s *spaceSettingService) BulkUpdate(ctx context.Context, req *structs.BulkUpdateSettingsRequest) error {
	for key, value := range req.Settings {
		if err := s.SetSetting(ctx, req.SpaceID, key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetSpaceSettings retrieves all settings for a space as key-value map
func (s *spaceSettingService) GetSpaceSettings(ctx context.Context, spaceID string, publicOnly bool) (map[string]any, error) {
	params := &structs.ListSpaceSettingParams{
		SpaceID: spaceID,
		Limit:   1000, // Get all settings
	}
	if publicOnly {
		isPublic := true
		params.IsPublic = &isPublic
	}

	rows, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for _, row := range rows {
		setting := s.Serialize(row)
		result[setting.SettingKey] = setting.GetTypedValue()
	}

	return result, nil
}

// SetSetting creates or updates a setting
func (s *spaceSettingService) SetSetting(ctx context.Context, spaceID, key, value string) error {
	existing, err := s.repo.GetByKey(ctx, spaceID, key)
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
	body := &structs.CreateSpaceSettingBody{
		SpaceSettingBody: structs.SpaceSettingBody{
			SpaceID:      spaceID,
			SettingKey:   key,
			SettingValue: value,
			SettingName:  key, // Use key as name for simplicity
			SettingType:  structs.TypeString,
			Scope:        structs.Scope,
		},
	}
	_, err = s.repo.Create(ctx, body)
	return err
}

// GetSettingValue retrieves a setting value with type conversion
func (s *spaceSettingService) GetSettingValue(ctx context.Context, spaceID, key string) (any, error) {
	setting, err := s.GetByKey(ctx, spaceID, key)
	if err != nil {
		return nil, err
	}

	return setting.GetTypedValue(), nil
}

// Serialize converts entity to struct
func (s *spaceSettingService) Serialize(row *ent.SpaceSetting) *structs.ReadSpaceSetting {
	return &structs.ReadSpaceSetting{
		ID:           row.ID,
		SpaceID:      row.SpaceID,
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
func (s *spaceSettingService) Serializes(rows []*ent.SpaceSetting) []*structs.ReadSpaceSetting {
	result := make([]*structs.ReadSpaceSetting, 0, len(rows))
	for _, row := range rows {
		result = append(result, s.Serialize(row))
	}
	return result
}
