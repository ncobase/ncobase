package service

import (
	"context"
	"fmt"
	"ncobase/resource/data"
	"ncobase/resource/data/repository"
	"ncobase/resource/event"
	"ncobase/resource/wrapper"
	"sync"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// QuotaServiceInterface defines the interface for quota management
type QuotaServiceInterface interface {
	CheckAndUpdateQuota(ctx context.Context, tenantID string, size int) (bool, error)
	GetUsage(ctx context.Context, tenantID string) (int64, error)
	SetQuota(ctx context.Context, tenantID string, quota int64) error
	GetQuota(ctx context.Context, tenantID string) (int64, error)
	IsQuotaExceeded(ctx context.Context, tenantID string) (bool, error)
	MonitorQuota(ctx context.Context) error
	UpdateUsage(ctx context.Context, tenantID string, quotaType string, delta int64) error
	RefreshTenantServices()
}

// QuotaConfig represents quota configuration
type QuotaConfig struct {
	DefaultQuota      int64         `json:"default_quota"`      // Default quota in bytes
	WarningThreshold  float64       `json:"warning_threshold"`  // Warning threshold percentage (0.0-1.0)
	CheckInterval     time.Duration `json:"check_interval"`     // Interval for quota checks
	EnableEnforcement bool          `json:"enable_enforcement"` // Whether to enforce quotas
}

// quotaService handles storage quota management with tenant service integration
type quotaService struct {
	fileRepo      repository.FileRepositoryInterface
	redis         *redis.Client
	config        *QuotaConfig
	publisher     event.PublisherInterface
	tenantWrapper *wrapper.TenantServiceWrapper
	quotaCache    map[string]int64
	usageCache    map[string]int64
	mu            sync.RWMutex // For cache thread safety
}

// NewQuotaService creates a new quota management service
func NewQuotaService(
	d *data.Data,
	publisher event.PublisherInterface,
	config *QuotaConfig,
	em ext.ManagerInterface,
) QuotaServiceInterface {
	// Set default config values if not provided
	if config == nil {
		config = &QuotaConfig{
			DefaultQuota:      10 * 1024 * 1024 * 1024, // 10GB default
			WarningThreshold:  0.8,                     // 80% warning
			CheckInterval:     24 * time.Hour,          // Daily check
			EnableEnforcement: true,                    // Enforce quotas
		}
	}

	return &quotaService{
		fileRepo:      repository.NewFileRepository(d),
		redis:         d.GetRedis(),
		config:        config,
		publisher:     publisher,
		tenantWrapper: wrapper.NewTenantServiceWrapper(em),
		quotaCache:    make(map[string]int64),
		usageCache:    make(map[string]int64),
	}
}

// CheckAndUpdateQuota checks if adding a file of the given size would exceed the quota
// and updates the usage if successful
func (s *quotaService) CheckAndUpdateQuota(ctx context.Context, tenantID string, size int) (bool, error) {
	if tenantID == "" {
		return false, fmt.Errorf("tenant ID is required")
	}

	// Use tenant service if available, otherwise fallback to local logic
	if s.tenantWrapper.HasTenantQuotaService() {
		// Check quota limit using tenant service
		allowed, err := s.tenantWrapper.CheckQuotaLimit(ctx, tenantID, "storage", int64(size))
		if err != nil {
			logger.Errorf(ctx, "Error checking quota limit from tenant service: %v", err)
			// Fallback to local logic
			return s.checkAndUpdateQuotaLocal(ctx, tenantID, size)
		}

		if !allowed && s.config.EnableEnforcement {
			// Publish quota exceeded event
			if s.publisher != nil {
				usage, _ := s.GetUsage(ctx, tenantID)
				quota, _ := s.GetQuota(ctx, tenantID)
				eventData := &event.StorageQuotaEventData{
					TenantID:     tenantID,
					CurrentUsage: usage,
					Quota:        quota,
					UsagePercent: float64(usage) / float64(quota) * 100,
					StorageType:  "file",
				}
				s.publisher.PublishStorageQuotaExceeded(ctx, eventData)
			}
			return false, fmt.Errorf("storage quota exceeded for tenant %s", tenantID)
		}

		// Update usage in tenant service
		if err := s.tenantWrapper.UpdateUsage(ctx, tenantID, "storage", int64(size)); err != nil {
			logger.Warnf(ctx, "Error updating usage in tenant service: %v", err)
		}

		// Check warning threshold
		if allowed {
			usage, _ := s.GetUsage(ctx, tenantID)
			quota, _ := s.GetQuota(ctx, tenantID)
			if quota > 0 && float64(usage+int64(size))/float64(quota) >= s.config.WarningThreshold && s.publisher != nil {
				eventData := &event.StorageQuotaEventData{
					TenantID:     tenantID,
					CurrentUsage: usage + int64(size),
					Quota:        quota,
					UsagePercent: float64(usage+int64(size)) / float64(quota) * 100,
					StorageType:  "file",
				}
				s.publisher.PublishStorageQuotaWarning(ctx, eventData)
			}
		}

		return allowed, nil
	}

	// Fallback to local quota management
	return s.checkAndUpdateQuotaLocal(ctx, tenantID, size)
}

// checkAndUpdateQuotaLocal handles quota checking with local logic
func (s *quotaService) checkAndUpdateQuotaLocal(ctx context.Context, tenantID string, size int) (bool, error) {
	// Get current usage
	currentUsage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		logger.Errorf(ctx, "Error getting current usage for tenant %s: %v", tenantID, err)
		return true, nil // Allow if can't get usage
	}

	// Get quota
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		logger.Errorf(ctx, "Error getting quota for tenant %s: %v", tenantID, err)
		return true, nil // Allow if can't get quota
	}

	// Check if adding this file would exceed quota
	newUsage := currentUsage + int64(size)
	if s.config.EnableEnforcement && newUsage > quota {
		// Publish quota exceeded event
		if s.publisher != nil {
			eventData := &event.StorageQuotaEventData{
				TenantID:     tenantID,
				CurrentUsage: currentUsage,
				Quota:        quota,
				UsagePercent: float64(currentUsage) / float64(quota) * 100,
				StorageType:  "file",
			}
			s.publisher.PublishStorageQuotaExceeded(ctx, eventData)
		}
		return false, fmt.Errorf("storage quota exceeded for tenant %s", tenantID)
	}

	// Update usage in cache and Redis
	s.mu.Lock()
	s.usageCache[tenantID] = newUsage
	s.mu.Unlock()

	// Set in Redis
	key := fmt.Sprintf("storage:usage:%s", tenantID)
	if s.redis != nil {
		err = s.redis.Set(ctx, key, newUsage, 0).Err()
		if err != nil {
			logger.Errorf(ctx, "Error updating usage in Redis for tenant %s: %v", tenantID, err)
		}
	}

	// Check warning threshold
	if float64(newUsage)/float64(quota) >= s.config.WarningThreshold && s.publisher != nil {
		eventData := &event.StorageQuotaEventData{
			TenantID:     tenantID,
			CurrentUsage: newUsage,
			Quota:        quota,
			UsagePercent: float64(newUsage) / float64(quota) * 100,
			StorageType:  "file",
		}
		s.publisher.PublishStorageQuotaWarning(ctx, eventData)
	}

	return true, nil
}

// GetUsage returns the current storage usage for a tenant
func (s *quotaService) GetUsage(ctx context.Context, tenantID string) (int64, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant ID is required")
	}

	// Use tenant service if available
	if s.tenantWrapper.HasTenantQuotaService() {
		usage, err := s.tenantWrapper.GetUsage(ctx, tenantID, "storage")
		if err == nil {
			return usage, nil
		}
		logger.Warnf(ctx, "Error getting usage from tenant service, fallback to local: %v", err)
	}

	// Fallback to local calculation
	// Check cache first
	s.mu.RLock()
	if usage, found := s.usageCache[tenantID]; found {
		s.mu.RUnlock()
		return usage, nil
	}
	s.mu.RUnlock()

	// Calculate from database
	usage, err := s.calculateUsage(ctx, tenantID)
	if err != nil {
		return 0, err
	}

	// Update cache
	s.mu.Lock()
	s.usageCache[tenantID] = usage
	s.mu.Unlock()

	return usage, nil
}

// calculateUsage calculates the total storage usage for a tenant from the database
func (s *quotaService) calculateUsage(ctx context.Context, tenantID string) (int64, error) {
	totalSize, err := s.fileRepo.SumSizeByTenant(ctx, tenantID)
	if err != nil {
		logger.Errorf(ctx, "Error calculating usage for tenant %s: %v", tenantID, err)
		return 0, err
	}
	return totalSize, nil
}

// SetQuota sets the storage quota for a tenant
func (s *quotaService) SetQuota(ctx context.Context, tenantID string, quota int64) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	// Update cache
	s.mu.Lock()
	s.quotaCache[tenantID] = quota
	s.mu.Unlock()

	// Update Redis
	key := fmt.Sprintf("storage:quota:%s", tenantID)
	if s.redis != nil {
		err := s.redis.Set(ctx, key, quota, 0).Err()
		if err != nil {
			logger.Errorf(ctx, "Error setting quota in Redis for tenant %s: %v", tenantID, err)
			return err
		}
	}

	return nil
}

// GetQuota returns the storage quota for a tenant
func (s *quotaService) GetQuota(ctx context.Context, tenantID string) (int64, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant ID is required")
	}

	// Use tenant service if available
	if s.tenantWrapper.HasTenantQuotaService() {
		quota, err := s.tenantWrapper.GetQuota(ctx, tenantID, "storage")
		if err == nil {
			return quota, nil
		}
		logger.Warnf(ctx, "Error getting quota from tenant service, fallback to local: %v", err)
	}

	// Fallback to local cache/redis
	// Check cache first
	s.mu.RLock()
	if quota, found := s.quotaCache[tenantID]; found {
		s.mu.RUnlock()
		return quota, nil
	}
	s.mu.RUnlock()

	// Check Redis
	key := fmt.Sprintf("storage:quota:%s", tenantID)
	var quota int64

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Int64()
		if err == nil {
			// Found in Redis
			s.mu.Lock()
			s.quotaCache[tenantID] = val
			s.mu.Unlock()
			return val, nil
		}
	}

	// Use default quota
	quota = s.config.DefaultQuota

	// Update cache and Redis
	s.mu.Lock()
	s.quotaCache[tenantID] = quota
	s.mu.Unlock()

	if s.redis != nil {
		err := s.redis.Set(ctx, key, quota, 0).Err()
		if err != nil {
			logger.Warnf(ctx, "Error setting quota in Redis for tenant %s: %v", tenantID, err)
		}
	}

	return quota, nil
}

// IsQuotaExceeded checks if the tenant's quota is exceeded
func (s *quotaService) IsQuotaExceeded(ctx context.Context, tenantID string) (bool, error) {
	if tenantID == "" {
		return false, fmt.Errorf("tenant ID is required")
	}

	// Use tenant service if available
	if s.tenantWrapper.HasTenantQuotaService() {
		exceeded, err := s.tenantWrapper.IsQuotaExceeded(ctx, tenantID, "storage")
		if err == nil {
			return exceeded, nil
		}
		logger.Warnf(ctx, "Error checking quota exceeded from tenant service, fallback to local: %v", err)
	}

	// Fallback to local calculation
	usage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		return false, err
	}

	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return false, err
	}

	return usage >= quota, nil
}

// MonitorQuota starts a background task to monitor quotas for all tenants
func (s *quotaService) MonitorQuota(ctx context.Context) error {
	// Get all tenants with files
	tenants, err := s.fileRepo.GetAllTenants(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error getting tenants for quota monitoring: %v", err)
		return err
	}

	// Check each tenant
	for _, tenantID := range tenants {
		usage, err := s.GetUsage(ctx, tenantID)
		if err != nil {
			logger.Errorf(ctx, "Error getting usage for tenant %s: %v", tenantID, err)
			continue
		}

		quota, err := s.GetQuota(ctx, tenantID)
		if err != nil {
			logger.Errorf(ctx, "Error getting quota for tenant %s: %v", tenantID, err)
			continue
		}

		// Calculate usage percentage
		usagePercent := float64(usage) / float64(quota) * 100

		// Check if quota is exceeded
		if usage >= quota && s.publisher != nil {
			eventData := &event.StorageQuotaEventData{
				TenantID:     tenantID,
				CurrentUsage: usage,
				Quota:        quota,
				UsagePercent: usagePercent,
				StorageType:  "file",
			}
			s.publisher.PublishStorageQuotaExceeded(ctx, eventData)
		} else if usagePercent >= s.config.WarningThreshold*100 && s.publisher != nil {
			// Check if warning threshold is reached
			eventData := &event.StorageQuotaEventData{
				TenantID:     tenantID,
				CurrentUsage: usage,
				Quota:        quota,
				UsagePercent: usagePercent,
				StorageType:  "file",
			}
			s.publisher.PublishStorageQuotaWarning(ctx, eventData)
		}
	}

	return nil
}

// UpdateUsage updates quota usage for a tenant (for external calls like file deletion)
func (s *quotaService) UpdateUsage(ctx context.Context, tenantID string, quotaType string, delta int64) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID is required")
	}

	// Use tenant service if available
	if s.tenantWrapper.HasTenantQuotaService() {
		err := s.tenantWrapper.UpdateUsage(ctx, tenantID, quotaType, delta)
		if err != nil {
			logger.Warnf(ctx, "Error updating usage in tenant service, fallback to local: %v", err)
		} else {
			return nil // Success with tenant service
		}
	}

	// Fallback to local cache update
	s.mu.Lock()
	if currentUsage, exists := s.usageCache[tenantID]; exists {
		newUsage := currentUsage + delta
		if newUsage < 0 {
			newUsage = 0
		}
		s.usageCache[tenantID] = newUsage
	}
	s.mu.Unlock()

	// Update Redis if available
	if s.redis != nil {
		key := fmt.Sprintf("storage:usage:%s", tenantID)
		// Get current value and update
		currentVal, err := s.redis.Get(ctx, key).Int64()
		if err == nil {
			newVal := currentVal + delta
			if newVal < 0 {
				newVal = 0
			}
			s.redis.Set(ctx, key, newVal, 0)
		}
	}

	return nil
}

// RefreshTenantServices refreshes tenant service references
func (s *quotaService) RefreshTenantServices() {
	s.tenantWrapper.RefreshServices()
}
