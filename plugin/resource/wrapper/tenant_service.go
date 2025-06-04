package wrapper

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
)

// TenantQuotaServiceInterface defines tenant quota service interface for resource plugin
type TenantQuotaServiceInterface interface {
	CheckQuotaLimit(ctx context.Context, tenantID string, quotaType string, requestedAmount int64) (bool, error)
	UpdateUsage(ctx context.Context, tenantID string, quotaType string, delta int64) error
	GetUsage(ctx context.Context, tenantID string, quotaType string) (int64, error)
	GetQuota(ctx context.Context, tenantID string, quotaType string) (int64, error)
	IsQuotaExceeded(ctx context.Context, tenantID string, quotaType string) (bool, error)
}

// TenantServiceWrapper wraps tenant service access with fallback behavior
type TenantServiceWrapper struct {
	em                 ext.ManagerInterface
	tenantQuotaService TenantQuotaServiceInterface
}

// NewTenantServiceWrapper creates a new tenant service wrapper
func NewTenantServiceWrapper(em ext.ManagerInterface) *TenantServiceWrapper {
	wrapper := &TenantServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads tenant services using extension manager
func (w *TenantServiceWrapper) loadServices() {
	// Try to get tenant quota service
	if quotaService, err := w.em.GetCrossService("tenant", "TenantQuota"); err == nil {
		if service, ok := quotaService.(TenantQuotaServiceInterface); ok {
			w.tenantQuotaService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *TenantServiceWrapper) RefreshServices() {
	w.loadServices()
}

// CheckQuotaLimit checks if tenant can use additional quota
func (w *TenantServiceWrapper) CheckQuotaLimit(ctx context.Context, tenantID string, quotaType string, requestedAmount int64) (bool, error) {
	if w.tenantQuotaService != nil {
		return w.tenantQuotaService.CheckQuotaLimit(ctx, tenantID, quotaType, requestedAmount)
	}

	// Fallback: allow usage if service not available
	return true, nil
}

// UpdateUsage updates quota usage for a tenant
func (w *TenantServiceWrapper) UpdateUsage(ctx context.Context, tenantID string, quotaType string, delta int64) error {
	if w.tenantQuotaService != nil {
		return w.tenantQuotaService.UpdateUsage(ctx, tenantID, quotaType, delta)
	}

	// Fallback: no-op if service not available
	return nil
}

// GetUsage gets current usage for a tenant
func (w *TenantServiceWrapper) GetUsage(ctx context.Context, tenantID string, quotaType string) (int64, error) {
	if w.tenantQuotaService != nil {
		return w.tenantQuotaService.GetUsage(ctx, tenantID, quotaType)
	}

	// Fallback: return 0 if service not available
	return 0, nil
}

// GetQuota gets quota limit for a tenant
func (w *TenantServiceWrapper) GetQuota(ctx context.Context, tenantID string, quotaType string) (int64, error) {
	if w.tenantQuotaService != nil {
		return w.tenantQuotaService.GetQuota(ctx, tenantID, quotaType)
	}

	// Fallback: return unlimited quota
	return 10 * 1024 * 1024 * 1024, nil // 10GB default
}

// IsQuotaExceeded checks if tenant's quota is exceeded
func (w *TenantServiceWrapper) IsQuotaExceeded(ctx context.Context, tenantID string, quotaType string) (bool, error) {
	if w.tenantQuotaService != nil {
		return w.tenantQuotaService.IsQuotaExceeded(ctx, tenantID, quotaType)
	}

	// Fallback: not exceeded if service not available
	return false, nil
}

// HasTenantQuotaService checks if tenant quota service is available
func (w *TenantServiceWrapper) HasTenantQuotaService() bool {
	return w.tenantQuotaService != nil
}
