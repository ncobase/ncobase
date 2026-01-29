package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/core/space/data"
	"ncobase/core/space/data/repository"
	"ncobase/core/space/structs"

	"github.com/ncobase/ncore/data/paging"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// SpaceQuotaServiceInterface defines the interface for space quota service
type SpaceQuotaServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateSpaceQuotaBody) (*structs.ReadSpaceQuota, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceQuota, error)
	Get(ctx context.Context, id string) (*structs.ReadSpaceQuota, error)
	GetBySpaceAndType(ctx context.Context, spaceID string, quotaType structs.QuotaType) (*structs.ReadSpaceQuota, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListSpaceQuotaParams) (paging.Result[*structs.ReadSpaceQuota], error)
	GetUsage(ctx context.Context, spaceID string, quotaType string) (int64, error)
	GetQuota(ctx context.Context, spaceID string, quotaType string) (int64, error)
	IsQuotaExceeded(ctx context.Context, spaceID string, quotaType string) (bool, error)
	UpdateUsage(ctx context.Context, spaceID string, quotaType string, delta int64) error
	CheckQuotaLimit(ctx context.Context, spaceID string, quotaType structs.QuotaType, requestedAmount int64) (bool, error)
	GetSpaceQuotaSummary(ctx context.Context, spaceID string) ([]*structs.ReadSpaceQuota, error)
}

// spaceQuotaService implements SpaceQuotaServiceInterface
type spaceQuotaService struct {
	repo repository.SpaceQuotaRepositoryInterface
}

// NewSpaceQuotaService creates a new space quota service
func NewSpaceQuotaService(d *data.Data) SpaceQuotaServiceInterface {
	return &spaceQuotaService{
		repo: repository.NewSpaceQuotaRepository(d),
	}
}

// Create creates a new space quota
func (s *spaceQuotaService) Create(ctx context.Context, body *structs.CreateSpaceQuotaBody) (*structs.ReadSpaceQuota, error) {
	if body.SpaceID == "" {
		return nil, errors.New(ecode.FieldIsRequired("space_id"))
	}
	if body.QuotaType == "" {
		return nil, errors.New(ecode.FieldIsRequired("quota_type"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "SpaceQuota", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceQuota(row), nil
}

// Update updates an existing space quota
func (s *spaceQuotaService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadSpaceQuota, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "SpaceQuota", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceQuota(row), nil
}

// Get retrieves a space quota by ID
func (s *spaceQuotaService) Get(ctx context.Context, id string) (*structs.ReadSpaceQuota, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "SpaceQuota", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceQuota(row), nil
}

// GetBySpaceAndType retrieves a space quota by space ID and quota type
func (s *spaceQuotaService) GetBySpaceAndType(ctx context.Context, spaceID string, quotaType structs.QuotaType) (*structs.ReadSpaceQuota, error) {
	row, err := s.repo.GetBySpaceAndType(ctx, spaceID, quotaType)
	if err := handleEntError(ctx, "SpaceQuota", err); err != nil {
		return nil, err
	}

	return repository.SerializeSpaceQuota(row), nil
}

// Delete deletes a space quota
func (s *spaceQuotaService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "SpaceQuota", err); err != nil {
		return err
	}
	return nil
}

// List lists space quotas
func (s *spaceQuotaService) List(ctx context.Context, params *structs.ListSpaceQuotaParams) (paging.Result[*structs.ReadSpaceQuota], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadSpaceQuota, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing space quotas: %v", err)
			return nil, 0, err
		}

		return repository.SerializeSpaceQuotas(rows), total, nil
	})
}

// GetUsage gets current usage for a space
func (s *spaceQuotaService) GetUsage(ctx context.Context, spaceID string, quotaType string) (int64, error) {
	quota, err := s.repo.GetBySpaceAndType(ctx, spaceID, structs.QuotaType(quotaType))
	if err != nil {
		if repository.IsNotFound(err) {
			return 0, nil // No quota set, return 0 usage
		}
		return 0, err
	}

	return quota.CurrentUsed, nil
}

// GetQuota gets quota limit for a space
func (s *spaceQuotaService) GetQuota(ctx context.Context, spaceID string, quotaType string) (int64, error) {
	quota, err := s.repo.GetBySpaceAndType(ctx, spaceID, structs.QuotaType(quotaType))
	if err != nil {
		if repository.IsNotFound(err) {
			// Return default quota for storage
			if quotaType == "storage" {
				return 10 * 1024 * 1024 * 1024, nil // 10GB default
			}
			return 0, nil
		}
		return 0, err
	}

	return quota.MaxValue, nil
}

// IsQuotaExceeded checks if space's quota is exceeded
func (s *spaceQuotaService) IsQuotaExceeded(ctx context.Context, spaceID string, quotaType string) (bool, error) {
	quota, err := s.repo.GetBySpaceAndType(ctx, spaceID, structs.QuotaType(quotaType))
	if err != nil {
		if repository.IsNotFound(err) {
			return false, nil // No quota set, not exceeded
		}
		return false, err
	}

	if !quota.Enabled {
		return false, nil // Quota disabled, not exceeded
	}

	return quota.CurrentUsed >= quota.MaxValue, nil
}

// UpdateUsage updates quota usage for a space
func (s *spaceQuotaService) UpdateUsage(ctx context.Context, spaceID string, quotaType string, delta int64) error {
	quota, err := s.repo.GetBySpaceAndType(ctx, spaceID, structs.QuotaType(quotaType))
	if err != nil {
		if repository.IsNotFound(err) {
			// Create quota if not exists
			createBody := &structs.CreateSpaceQuotaBody{
				SpaceQuotaBody: structs.SpaceQuotaBody{
					SpaceID:     spaceID,
					QuotaType:   structs.QuotaType(quotaType),
					QuotaName:   fmt.Sprintf("%s quota", quotaType),
					MaxValue:    10 * 1024 * 1024 * 1024, // 10GB default
					CurrentUsed: delta,
					Unit:        structs.UnitBytes,
					Description: fmt.Sprintf("Auto-created %s quota", quotaType),
					Enabled:     true,
				},
			}

			_, err = s.repo.Create(ctx, createBody)
			return handleEntError(ctx, "SpaceQuota", err)
		}
		return err
	}

	newUsage := quota.CurrentUsed + delta
	if newUsage < 0 {
		newUsage = 0
	}

	updates := types.JSON{
		"current_used": newUsage,
	}

	_, err = s.repo.Update(ctx, quota.ID, updates)
	return handleEntError(ctx, "SpaceQuota", err)
}

// CheckQuotaLimit checks if space can use additional quota
func (s *spaceQuotaService) CheckQuotaLimit(ctx context.Context, spaceID string, quotaType structs.QuotaType, requestedAmount int64) (bool, error) {
	quota, err := s.repo.GetBySpaceAndType(ctx, spaceID, quotaType)
	if err != nil {
		if repository.IsNotFound(err) {
			return true, nil // No quota set, allow usage
		}
		return false, err
	}

	if !quota.Enabled {
		return true, nil // Quota disabled, allow usage
	}

	return quota.CurrentUsed+requestedAmount <= quota.MaxValue, nil
}

// GetSpaceQuotaSummary retrieves all quotas for a space
func (s *spaceQuotaService) GetSpaceQuotaSummary(ctx context.Context, spaceID string) ([]*structs.ReadSpaceQuota, error) {
	rows, err := s.repo.GetBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, handleEntError(ctx, "SpaceQuota", err)
	}

	return repository.SerializeSpaceQuotas(rows), nil
}
