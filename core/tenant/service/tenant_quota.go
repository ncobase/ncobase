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

// TenantQuotaServiceInterface defines the interface for tenant quota service
type TenantQuotaServiceInterface interface {
	Create(ctx context.Context, body *structs.CreateTenantQuotaBody) (*structs.ReadTenantQuota, error)
	Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantQuota, error)
	Get(ctx context.Context, id string) (*structs.ReadTenantQuota, error)
	GetByTenantAndType(ctx context.Context, tenantID string, quotaType structs.QuotaType) (*structs.ReadTenantQuota, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, params *structs.ListTenantQuotaParams) (paging.Result[*structs.ReadTenantQuota], error)
	UpdateUsage(ctx context.Context, req *structs.QuotaUsageRequest) error
	CheckQuotaLimit(ctx context.Context, tenantID string, quotaType structs.QuotaType, requestedAmount int64) (bool, error)
	GetTenantQuotaSummary(ctx context.Context, tenantID string) ([]*structs.ReadTenantQuota, error)
	Serialize(row *ent.TenantQuota) *structs.ReadTenantQuota
	Serializes(rows []*ent.TenantQuota) []*structs.ReadTenantQuota
}

// tenantQuotaService implements TenantQuotaServiceInterface
type tenantQuotaService struct {
	repo repository.TenantQuotaRepositoryInterface
}

// NewTenantQuotaService creates a new tenant quota service
func NewTenantQuotaService(d *data.Data) TenantQuotaServiceInterface {
	return &tenantQuotaService{
		repo: repository.NewTenantQuotaRepository(d),
	}
}

// Create creates a new tenant quota
func (s *tenantQuotaService) Create(ctx context.Context, body *structs.CreateTenantQuotaBody) (*structs.ReadTenantQuota, error) {
	if body.TenantID == "" {
		return nil, errors.New(ecode.FieldIsRequired("tenant_id"))
	}
	if body.QuotaType == "" {
		return nil, errors.New(ecode.FieldIsRequired("quota_type"))
	}

	row, err := s.repo.Create(ctx, body)
	if err := handleEntError(ctx, "TenantQuota", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Update updates an existing tenant quota
func (s *tenantQuotaService) Update(ctx context.Context, id string, updates types.JSON) (*structs.ReadTenantQuota, error) {
	row, err := s.repo.Update(ctx, id, updates)
	if err := handleEntError(ctx, "TenantQuota", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Get retrieves a tenant quota by ID
func (s *tenantQuotaService) Get(ctx context.Context, id string) (*structs.ReadTenantQuota, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err := handleEntError(ctx, "TenantQuota", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// GetByTenantAndType retrieves a tenant quota by tenant ID and quota type
func (s *tenantQuotaService) GetByTenantAndType(ctx context.Context, tenantID string, quotaType structs.QuotaType) (*structs.ReadTenantQuota, error) {
	row, err := s.repo.GetByTenantAndType(ctx, tenantID, quotaType)
	if err := handleEntError(ctx, "TenantQuota", err); err != nil {
		return nil, err
	}

	return s.Serialize(row), nil
}

// Delete deletes a tenant quota
func (s *tenantQuotaService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err := handleEntError(ctx, "TenantQuota", err); err != nil {
		return err
	}
	return nil
}

// List lists tenant quotas
func (s *tenantQuotaService) List(ctx context.Context, params *structs.ListTenantQuotaParams) (paging.Result[*structs.ReadTenantQuota], error) {
	pp := paging.Params{
		Cursor:    params.Cursor,
		Limit:     params.Limit,
		Direction: params.Direction,
	}

	return paging.Paginate(pp, func(cursor string, limit int, direction string) ([]*structs.ReadTenantQuota, int, error) {
		lp := *params
		lp.Cursor = cursor
		lp.Limit = limit
		lp.Direction = direction

		rows, total, err := s.repo.ListWithCount(ctx, &lp)
		if err != nil {
			logger.Errorf(ctx, "Error listing tenant quotas: %v", err)
			return nil, 0, err
		}

		return s.Serializes(rows), total, nil
	})
}

// UpdateUsage updates quota usage for a tenant
func (s *tenantQuotaService) UpdateUsage(ctx context.Context, req *structs.QuotaUsageRequest) error {
	quota, err := s.repo.GetByTenantAndType(ctx, req.TenantID, req.QuotaType)
	if err != nil {
		return handleEntError(ctx, "TenantQuota", err)
	}

	newUsage := quota.CurrentUsed + req.Delta
	if newUsage < 0 {
		newUsage = 0
	}

	updates := types.JSON{
		"current_used": newUsage,
	}

	_, err = s.repo.Update(ctx, quota.ID, updates)
	return handleEntError(ctx, "TenantQuota", err)
}

// CheckQuotaLimit checks if tenant can use additional quota
func (s *tenantQuotaService) CheckQuotaLimit(ctx context.Context, tenantID string, quotaType structs.QuotaType, requestedAmount int64) (bool, error) {
	quota, err := s.repo.GetByTenantAndType(ctx, tenantID, quotaType)
	if err != nil {
		if ent.IsNotFound(err) {
			return true, nil // No quota set, allow usage
		}
		return false, err
	}

	if !quota.Enabled {
		return true, nil // Quota disabled, allow usage
	}

	return quota.CurrentUsed+requestedAmount <= quota.MaxValue, nil
}

// GetTenantQuotaSummary retrieves all quotas for a tenant
func (s *tenantQuotaService) GetTenantQuotaSummary(ctx context.Context, tenantID string) ([]*structs.ReadTenantQuota, error) {
	rows, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, handleEntError(ctx, "TenantQuota", err)
	}

	return s.Serializes(rows), nil
}

// Serialize converts entity to struct
func (s *tenantQuotaService) Serialize(row *ent.TenantQuota) *structs.ReadTenantQuota {
	result := &structs.ReadTenantQuota{
		ID:          row.ID,
		TenantID:    row.TenantID,
		QuotaType:   structs.QuotaType(row.QuotaType),
		QuotaName:   row.QuotaName,
		MaxValue:    row.MaxValue,
		CurrentUsed: row.CurrentUsed,
		Unit:        structs.QuotaUnit(row.Unit),
		Description: row.Description,
		Enabled:     row.Enabled,
		Extras:      &row.Extras,
		CreatedBy:   &row.CreatedBy,
		CreatedAt:   &row.CreatedAt,
		UpdatedBy:   &row.UpdatedBy,
		UpdatedAt:   &row.UpdatedAt,
	}

	result.CalculateUtilization()
	return result
}

// Serializes converts multiple entities to structs
func (s *tenantQuotaService) Serializes(rows []*ent.TenantQuota) []*structs.ReadTenantQuota {
	result := make([]*structs.ReadTenantQuota, len(rows))
	for i, row := range rows {
		result[i] = s.Serialize(row)
	}
	return result
}
