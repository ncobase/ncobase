package service

import (
	"context"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
)

// TenantOptionServiceInterface is the interface for the service.
type TenantOptionServiceInterface interface {
	AddOptionsToTenant(ctx context.Context, tenantID, optionsID string) (*structs.TenantOption, error)
	RemoveOptionsFromTenant(ctx context.Context, tenantID, optionsID string) error
	IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error)
	GetTenantOption(ctx context.Context, tenantID string) ([]string, error)
	GetOptionsTenants(ctx context.Context, optionsID string) ([]string, error)
	RemoveAllOptionsFromTenant(ctx context.Context, tenantID string) error
	RemoveOptionsFromAllTenants(ctx context.Context, optionsID string) error
}

// tenantOptionService is the struct for the service.
type tenantOptionService struct {
	tenantOption repository.TenantOptionRepositoryInterface
}

// NewTenantOptionService creates a new service.
func NewTenantOptionService(d *data.Data) TenantOptionServiceInterface {
	return &tenantOptionService{
		tenantOption: repository.NewTenantOptionRepository(d),
	}
}

// AddOptionsToTenant adds options to a tenant.
func (s *tenantOptionService) AddOptionsToTenant(ctx context.Context, tenantID, optionsID string) (*structs.TenantOption, error) {
	row, err := s.tenantOption.Create(ctx, &structs.TenantOption{
		TenantID: tenantID,
		OptionID: optionsID,
	})
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return nil, err
	}
	return s.SerializeTenantOption(row), nil
}

// RemoveOptionsFromTenant removes options from a tenant.
func (s *tenantOptionService) RemoveOptionsFromTenant(ctx context.Context, tenantID, optionsID string) error {
	err := s.tenantOption.DeleteByTenantIDAndOptionID(ctx, tenantID, optionsID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return err
	}
	return nil
}

// IsOptionsInTenant checks if options belong to a tenant.
func (s *tenantOptionService) IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error) {
	exists, err := s.tenantOption.IsOptionsInTenant(ctx, tenantID, optionsID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetTenantOption retrieves all options IDs for a tenant.
func (s *tenantOptionService) GetTenantOption(ctx context.Context, tenantID string) ([]string, error) {
	optionsIDs, err := s.tenantOption.GetTenantOption(ctx, tenantID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return nil, err
	}
	return optionsIDs, nil
}

// GetOptionsTenants retrieves all tenant IDs for options.
func (s *tenantOptionService) GetOptionsTenants(ctx context.Context, optionsID string) ([]string, error) {
	tenantOption, err := s.tenantOption.GetByOptionID(ctx, optionsID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return nil, err
	}

	var tenantIDs []string
	for _, to := range tenantOption {
		tenantIDs = append(tenantIDs, to.TenantID)
	}

	return tenantIDs, nil
}

// RemoveAllOptionsFromTenant removes all options from a tenant.
func (s *tenantOptionService) RemoveAllOptionsFromTenant(ctx context.Context, tenantID string) error {
	err := s.tenantOption.DeleteAllByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return err
	}
	return nil
}

// RemoveOptionsFromAllTenants removes options from all tenants.
func (s *tenantOptionService) RemoveOptionsFromAllTenants(ctx context.Context, optionsID string) error {
	err := s.tenantOption.DeleteAllByOptionID(ctx, optionsID)
	if err := handleEntError(ctx, "TenantOption", err); err != nil {
		return err
	}
	return nil
}

// SerializeTenantOption serializes a tenant option relationship.
func (s *tenantOptionService) SerializeTenantOption(row *ent.TenantOption) *structs.TenantOption {
	return &structs.TenantOption{
		TenantID: row.TenantID,
		OptionID: row.OptionID,
	}
}
