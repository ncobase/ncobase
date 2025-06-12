package wrapper

import (
	"context"

	ext "github.com/ncobase/ncore/extension/types"
)

// SpaceQuotaServiceInterface defines space quota service interface for resource plugin
type SpaceQuotaServiceInterface interface {
	CheckQuotaLimit(ctx context.Context, spaceID string, quotaType string, requestedAmount int64) (bool, error)
	UpdateUsage(ctx context.Context, spaceID string, quotaType string, delta int64) error
	GetUsage(ctx context.Context, spaceID string, quotaType string) (int64, error)
	GetQuota(ctx context.Context, spaceID string, quotaType string) (int64, error)
	IsQuotaExceeded(ctx context.Context, spaceID string, quotaType string) (bool, error)
}

// SpaceServiceWrapper wraps space service access with fallback behavior
type SpaceServiceWrapper struct {
	em                ext.ManagerInterface
	spaceQuotaService SpaceQuotaServiceInterface
}

// NewSpaceServiceWrapper creates a new space service wrapper
func NewSpaceServiceWrapper(em ext.ManagerInterface) *SpaceServiceWrapper {
	wrapper := &SpaceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads space services using extension manager
func (w *SpaceServiceWrapper) loadServices() {
	// Try to get space quota service
	if quotaService, err := w.em.GetCrossService("space", "SpaceQuota"); err == nil {
		if service, ok := quotaService.(SpaceQuotaServiceInterface); ok {
			w.spaceQuotaService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *SpaceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// CheckQuotaLimit checks if space can use additional quota
func (w *SpaceServiceWrapper) CheckQuotaLimit(ctx context.Context, spaceID string, quotaType string, requestedAmount int64) (bool, error) {
	if w.spaceQuotaService != nil {
		return w.spaceQuotaService.CheckQuotaLimit(ctx, spaceID, quotaType, requestedAmount)
	}

	// Fallback: allow usage if service not available
	return true, nil
}

// UpdateUsage updates quota usage for a space
func (w *SpaceServiceWrapper) UpdateUsage(ctx context.Context, spaceID string, quotaType string, delta int64) error {
	if w.spaceQuotaService != nil {
		return w.spaceQuotaService.UpdateUsage(ctx, spaceID, quotaType, delta)
	}

	// Fallback: no-op if service not available
	return nil
}

// GetUsage gets current usage for a space
func (w *SpaceServiceWrapper) GetUsage(ctx context.Context, spaceID string, quotaType string) (int64, error) {
	if w.spaceQuotaService != nil {
		return w.spaceQuotaService.GetUsage(ctx, spaceID, quotaType)
	}

	// Fallback: return 0 if service not available
	return 0, nil
}

// GetQuota gets quota limit for a space
func (w *SpaceServiceWrapper) GetQuota(ctx context.Context, spaceID string, quotaType string) (int64, error) {
	if w.spaceQuotaService != nil {
		return w.spaceQuotaService.GetQuota(ctx, spaceID, quotaType)
	}

	// Fallback: return unlimited quota
	return 10 * 1024 * 1024 * 1024, nil // 10GB default
}

// IsQuotaExceeded checks if space's quota is exceeded
func (w *SpaceServiceWrapper) IsQuotaExceeded(ctx context.Context, spaceID string, quotaType string) (bool, error) {
	if w.spaceQuotaService != nil {
		return w.spaceQuotaService.IsQuotaExceeded(ctx, spaceID, quotaType)
	}

	// Fallback: not exceeded if service not available
	return false, nil
}

// HasSpaceQuotaService checks if space quota service is available
func (w *SpaceServiceWrapper) HasSpaceQuotaService() bool {
	return w.spaceQuotaService != nil
}
