package service

import (
	"context"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
)

// TenantOptionsServiceInterface is the interface for the service.
type TenantOptionsServiceInterface interface {
	AddOptionsToTenant(ctx context.Context, tenantID, optionsID string) (*structs.TenantOptions, error)
	RemoveOptionsFromTenant(ctx context.Context, tenantID, optionsID string) error
	IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error)
	GetTenantOptions(ctx context.Context, tenantID string) ([]string, error)
	GetOptionsTenants(ctx context.Context, optionsID string) ([]string, error)
	RemoveAllOptionsFromTenant(ctx context.Context, tenantID string) error
	RemoveOptionsFromAllTenants(ctx context.Context, optionsID string) error
}

// tenantOptionsService is the struct for the service.
type tenantOptionsService struct {
	tenantOptions repository.TenantOptionsRepositoryInterface
}

// NewTenantOptionsService creates a new service.
func NewTenantOptionsService(d *data.Data) TenantOptionsServiceInterface {
	return &tenantOptionsService{
		tenantOptions: repository.NewTenantOptionsRepository(d),
	}
}

// AddOptionsToTenant adds options to a tenant.
func (s *tenantOptionsService) AddOptionsToTenant(ctx context.Context, tenantID, optionsID string) (*structs.TenantOptions, error) {
	row, err := s.tenantOptions.Create(ctx, &structs.TenantOptions{
		TenantID:  tenantID,
		OptionsID: optionsID,
	})
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return nil, err
	}
	return s.SerializeTenantOptions(row), nil
}

// RemoveOptionsFromTenant removes options from a tenant.
func (s *tenantOptionsService) RemoveOptionsFromTenant(ctx context.Context, tenantID, optionsID string) error {
	err := s.tenantOptions.DeleteByTenantIDAndOptionsID(ctx, tenantID, optionsID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return err
	}
	return nil
}

// IsOptionsInTenant checks if options belong to a tenant.
func (s *tenantOptionsService) IsOptionsInTenant(ctx context.Context, tenantID, optionsID string) (bool, error) {
	exists, err := s.tenantOptions.IsOptionsInTenant(ctx, tenantID, optionsID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetTenantOptions retrieves all options IDs for a tenant.
func (s *tenantOptionsService) GetTenantOptions(ctx context.Context, tenantID string) ([]string, error) {
	optionsIDs, err := s.tenantOptions.GetTenantOptions(ctx, tenantID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return nil, err
	}
	return optionsIDs, nil
}

// GetOptionsTenants retrieves all tenant IDs for options.
func (s *tenantOptionsService) GetOptionsTenants(ctx context.Context, optionsID string) ([]string, error) {
	tenantOptions, err := s.tenantOptions.GetByOptionsID(ctx, optionsID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return nil, err
	}

	var tenantIDs []string
	for _, to := range tenantOptions {
		tenantIDs = append(tenantIDs, to.TenantID)
	}

	return tenantIDs, nil
}

// RemoveAllOptionsFromTenant removes all options from a tenant.
func (s *tenantOptionsService) RemoveAllOptionsFromTenant(ctx context.Context, tenantID string) error {
	err := s.tenantOptions.DeleteAllByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return err
	}
	return nil
}

// RemoveOptionsFromAllTenants removes options from all tenants.
func (s *tenantOptionsService) RemoveOptionsFromAllTenants(ctx context.Context, optionsID string) error {
	err := s.tenantOptions.DeleteAllByOptionsID(ctx, optionsID)
	if err := handleEntError(ctx, "TenantOptions", err); err != nil {
		return err
	}
	return nil
}

// SerializeTenantOptions serializes a tenant options relationship.
func (s *tenantOptionsService) SerializeTenantOptions(row *ent.TenantOptions) *structs.TenantOptions {
	return &structs.TenantOptions{
		TenantID:  row.TenantID,
		OptionsID: row.OptionsID,
	}
}
