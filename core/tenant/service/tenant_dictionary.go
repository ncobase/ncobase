package service

import (
	"context"
	"ncobase/tenant/data"
	"ncobase/tenant/data/ent"
	"ncobase/tenant/data/repository"
	"ncobase/tenant/structs"
)

// TenantDictionaryServiceInterface is the interface for the service.
type TenantDictionaryServiceInterface interface {
	AddDictionaryToTenant(ctx context.Context, tenantID, dictionaryID string) (*structs.TenantDictionary, error)
	RemoveDictionaryFromTenant(ctx context.Context, tenantID, dictionaryID string) error
	IsDictionaryInTenant(ctx context.Context, tenantID, dictionaryID string) (bool, error)
	GetTenantDictionaries(ctx context.Context, tenantID string) ([]string, error)
	GetDictionaryTenants(ctx context.Context, dictionaryID string) ([]string, error)
	RemoveAllDictionariesFromTenant(ctx context.Context, tenantID string) error
	RemoveDictionaryFromAllTenants(ctx context.Context, dictionaryID string) error
}

// tenantDictionaryService is the struct for the service.
type tenantDictionaryService struct {
	tenantDictionary repository.TenantDictionaryRepositoryInterface
}

// NewTenantDictionaryService creates a new service.
func NewTenantDictionaryService(d *data.Data) TenantDictionaryServiceInterface {
	return &tenantDictionaryService{
		tenantDictionary: repository.NewTenantDictionaryRepository(d),
	}
}

// AddDictionaryToTenant adds a dictionary to a tenant.
func (s *tenantDictionaryService) AddDictionaryToTenant(ctx context.Context, tenantID, dictionaryID string) (*structs.TenantDictionary, error) {
	row, err := s.tenantDictionary.Create(ctx, &structs.TenantDictionary{
		TenantID:     tenantID,
		DictionaryID: dictionaryID,
	})
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return nil, err
	}
	return s.SerializeTenantDictionary(row), nil
}

// RemoveDictionaryFromTenant removes a dictionary from a tenant.
func (s *tenantDictionaryService) RemoveDictionaryFromTenant(ctx context.Context, tenantID, dictionaryID string) error {
	err := s.tenantDictionary.DeleteByTenantIDAndDictionaryID(ctx, tenantID, dictionaryID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return err
	}
	return nil
}

// IsDictionaryInTenant checks if a dictionary belongs to a tenant.
func (s *tenantDictionaryService) IsDictionaryInTenant(ctx context.Context, tenantID, dictionaryID string) (bool, error) {
	exists, err := s.tenantDictionary.IsDictionaryInTenant(ctx, tenantID, dictionaryID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return false, err
	}
	return exists, nil
}

// GetTenantDictionaries retrieves all dictionary IDs for a tenant.
func (s *tenantDictionaryService) GetTenantDictionaries(ctx context.Context, tenantID string) ([]string, error) {
	dictionaryIDs, err := s.tenantDictionary.GetTenantDictionaries(ctx, tenantID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return nil, err
	}
	return dictionaryIDs, nil
}

// GetDictionaryTenants retrieves all tenant IDs for a dictionary.
func (s *tenantDictionaryService) GetDictionaryTenants(ctx context.Context, dictionaryID string) ([]string, error) {
	tenantDictionaries, err := s.tenantDictionary.GetByDictionaryID(ctx, dictionaryID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return nil, err
	}

	var tenantIDs []string
	for _, td := range tenantDictionaries {
		tenantIDs = append(tenantIDs, td.TenantID)
	}

	return tenantIDs, nil
}

// RemoveAllDictionariesFromTenant removes all dictionaries from a tenant.
func (s *tenantDictionaryService) RemoveAllDictionariesFromTenant(ctx context.Context, tenantID string) error {
	err := s.tenantDictionary.DeleteAllByTenantID(ctx, tenantID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return err
	}
	return nil
}

// RemoveDictionaryFromAllTenants removes a dictionary from all tenants.
func (s *tenantDictionaryService) RemoveDictionaryFromAllTenants(ctx context.Context, dictionaryID string) error {
	err := s.tenantDictionary.DeleteAllByDictionaryID(ctx, dictionaryID)
	if err := handleEntError(ctx, "TenantDictionary", err); err != nil {
		return err
	}
	return nil
}

// SerializeTenantDictionary serializes a tenant dictionary relationship.
func (s *tenantDictionaryService) SerializeTenantDictionary(row *ent.TenantDictionary) *structs.TenantDictionary {
	return &structs.TenantDictionary{
		TenantID:     row.TenantID,
		DictionaryID: row.DictionaryID,
	}
}
