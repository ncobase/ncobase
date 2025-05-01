package service

import (
	"context"
	"fmt"
	"ncobase/domain/resource/data"
	"ncobase/domain/resource/data/repository"
	"ncobase/domain/resource/event"
	"sync"
	"time"

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
}

// QuotaConfig represents quota configuration
type QuotaConfig struct {
	DefaultQuota      int64         `json:"default_quota"`      // Default quota in bytes
	WarningThreshold  float64       `json:"warning_threshold"`  // Warning threshold percentage (0.0-1.0)
	CheckInterval     time.Duration `json:"check_interval"`     // Interval for quota checks
	EnableEnforcement bool          `json:"enable_enforcement"` // Whether to enforce quotas
}

// quotaService handles storage quota management
type quotaService struct {
	fielRepo   repository.FileRepositoryInterface
	redis      *redis.Client
	config     *QuotaConfig
	publisher  event.PublisherInterface
	quotaCache map[string]int64
	usageCache map[string]int64
	mu         sync.RWMutex // For cache thread safety
}

// NewQuotaService creates a new quota management service
func NewQuotaService(
	d *data.Data,
	publisher event.PublisherInterface,
	config *QuotaConfig,
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
		fielRepo:   repository.NewFileRepository(d),
		redis:      d.GetRedis(),
		config:     config,
		publisher:  publisher,
		quotaCache: make(map[string]int64),
		usageCache: make(map[string]int64),
	}
}

// CheckAndUpdateQuota checks if adding a file of the given size would exceed the quota
// and updates the usage if successful
func (s *quotaService) CheckAndUpdateQuota(ctx context.Context, tenantID string, size int) (bool, error) {
	if tenantID == "" {
		return false, fmt.Errorf("tenant ID is required")
	}

	// Get current usage
	currentUsage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		logger.Errorf(ctx, "Error getting current usage for tenant %s: %v", tenantID, err)
		// If we can't get usage, allow the operation but don't update usage
		return true, nil
	}

	// Get quota
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		logger.Errorf(ctx, "Error getting quota for tenant %s: %v", tenantID, err)
		// If we can't get quota, allow the operation but don't update usage
		return true, nil
	}

	// Check if adding this file would exceed quota
	newUsage := currentUsage + int64(size)
	if s.config.EnableEnforcement && newUsage > quota {
		// Exceeds quota, publish event
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

	// Check if we've crossed the warning threshold
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

	// Check cache first
	s.mu.RLock()
	if usage, found := s.usageCache[tenantID]; found {
		s.mu.RUnlock()
		return usage, nil
	}
	s.mu.RUnlock()

	// If not in cache, check Redis
	key := fmt.Sprintf("storage:usage:%s", tenantID)
	var usage int64

	if s.redis != nil {
		val, err := s.redis.Get(ctx, key).Int64()
		if err == nil {
			// Found in Redis
			s.mu.Lock()
			s.usageCache[tenantID] = val
			s.mu.Unlock()
			return val, nil
		}

		// If not found in Redis (or error), calculate from database
		if err != redis.Nil {
			logger.Warnf(ctx, "Error getting usage from Redis for tenant %s: %v", tenantID, err)
		}
	}

	// Calculate from database
	usage, err := s.calculateUsage(ctx, tenantID)
	if err != nil {
		return 0, err
	}

	// Update cache and Redis
	s.mu.Lock()
	s.usageCache[tenantID] = usage
	s.mu.Unlock()

	if s.redis != nil {
		err = s.redis.Set(ctx, key, usage, 0).Err()
		if err != nil {
			logger.Warnf(ctx, "Error setting usage in Redis for tenant %s: %v", tenantID, err)
		}
	}

	return usage, nil
}

// calculateUsage calculates the total storage usage for a tenant from the database
func (s *quotaService) calculateUsage(ctx context.Context, tenantID string) (int64, error) {
	// Implement database query to sum all file sizes for the tenant
	totalSize, err := s.fielRepo.SumSizeByTenant(ctx, tenantID)
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

	// Check cache first
	s.mu.RLock()
	if quota, found := s.quotaCache[tenantID]; found {
		s.mu.RUnlock()
		return quota, nil
	}
	s.mu.RUnlock()

	// If not in cache, check Redis
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

		// If not found in Redis (or error), use default quota
		if err != redis.Nil {
			logger.Warnf(ctx, "Error getting quota from Redis for tenant %s: %v", tenantID, err)
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

	// Get current usage
	usage, err := s.GetUsage(ctx, tenantID)
	if err != nil {
		return false, err
	}

	// Get quota
	quota, err := s.GetQuota(ctx, tenantID)
	if err != nil {
		return false, err
	}

	return usage >= quota, nil
}

// MonitorQuota starts a background task to monitor quotas for all tenants
func (s *quotaService) MonitorQuota(ctx context.Context) error {
	// This would typically be run as a background goroutine or scheduled task
	// For simplicity, we'll just implement the check logic here

	// Get all tenants with files
	tenants, err := s.fielRepo.GetAllTenants(ctx)
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
