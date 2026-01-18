package service

import (
	"context"
	"fmt"
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/event"
	"sync"
	"time"

	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// QuotaServiceInterface defines quota management methods
type QuotaServiceInterface interface {
	CheckAndUpdateQuota(ctx context.Context, ownerID string, size int) (bool, error)
	GetUsage(ctx context.Context, ownerID string) (int64, error)
	SetQuota(ctx context.Context, ownerID string, quota int64) error
	GetQuota(ctx context.Context, ownerID string) (int64, error)
	IsQuotaExceeded(ctx context.Context, ownerID string) (bool, error)
	MonitorQuota(ctx context.Context) error
	UpdateUsage(ctx context.Context, ownerID string, quotaType string, delta int64) error
	RefreshSpaceServices()
}

// QuotaConfig represents quota configuration
type QuotaConfig struct {
	DefaultQuota      int64         `json:"default_quota"`
	WarningThreshold  float64       `json:"warning_threshold"`
	CheckInterval     time.Duration `json:"check_interval"`
	EnableEnforcement bool          `json:"enable_enforcement"`
}

type quotaService struct {
	fileRepo   repository.FileRepositoryInterface
	redis      *redis.Client
	config     *QuotaConfig
	publisher  event.PublisherInterface
	quotaCache map[string]int64
	usageCache map[string]int64
	mu         sync.RWMutex
}

// NewQuotaService creates new quota service
func NewQuotaService(d *data.Data, publisher event.PublisherInterface, config *QuotaConfig) QuotaServiceInterface {
	if config == nil {
		config = &QuotaConfig{
			DefaultQuota:      10 * 1024 * 1024 * 1024, // 10GB default
			WarningThreshold:  0.8,                     // 80% warning
			CheckInterval:     24 * time.Hour,          // Daily check
			EnableEnforcement: true,                    // Enforce quotas
		}
	}

	return &quotaService{
		fileRepo:   repository.NewFileRepository(d),
		redis:      d.GetRedis().(*redis.Client),
		config:     config,
		publisher:  publisher,
		quotaCache: make(map[string]int64),
		usageCache: make(map[string]int64),
	}
}

// CheckAndUpdateQuota checks and updates quota usage
func (s *quotaService) CheckAndUpdateQuota(ctx context.Context, ownerID string, size int) (bool, error) {
	if ownerID == "" {
		return false, fmt.Errorf("owner ID is required")
	}

	currentUsage, err := s.GetUsage(ctx, ownerID)
	if err != nil {
		logger.Errorf(ctx, "Error getting usage for owner %s: %v", ownerID, err)
		return true, nil // Allow if can't get usage
	}

	quota, err := s.GetQuota(ctx, ownerID)
	if err != nil {
		logger.Errorf(ctx, "Error getting quota for owner %s: %v", ownerID, err)
		return true, nil // Allow if can't get quota
	}

	newUsage := currentUsage + int64(size)
	if s.config.EnableEnforcement && newUsage > quota {
		if s.publisher != nil {
			eventData := &event.StorageQuotaEventData{
				SpaceID:      ownerID, // Using ownerID as spaceID for compatibility
				CurrentUsage: currentUsage,
				Quota:        quota,
				UsagePercent: float64(currentUsage) / float64(quota) * 100,
				StorageType:  "file",
			}
			s.publisher.PublishStorageQuotaExceeded(ctx, eventData)
		}
		return false, fmt.Errorf("storage quota exceeded for owner %s", ownerID)
	}

	// Update usage
	s.mu.Lock()
	s.usageCache[ownerID] = newUsage
	s.mu.Unlock()

	if s.redis != nil {
		key := fmt.Sprintf("storage:usage:%s", ownerID)
		s.redis.Set(ctx, key, newUsage, 0)
	}

	// Check warning threshold
	if float64(newUsage)/float64(quota) >= s.config.WarningThreshold && s.publisher != nil {
		eventData := &event.StorageQuotaEventData{
			SpaceID:      ownerID,
			CurrentUsage: newUsage,
			Quota:        quota,
			UsagePercent: float64(newUsage) / float64(quota) * 100,
			StorageType:  "file",
		}
		s.publisher.PublishStorageQuotaWarning(ctx, eventData)
	}

	return true, nil
}

// GetUsage returns current storage usage
func (s *quotaService) GetUsage(ctx context.Context, ownerID string) (int64, error) {
	if ownerID == "" {
		return 0, fmt.Errorf("owner ID is required")
	}

	// Check cache
	s.mu.RLock()
	if usage, found := s.usageCache[ownerID]; found {
		s.mu.RUnlock()
		return usage, nil
	}
	s.mu.RUnlock()

	// Calculate from database
	usage, err := s.calculateUsage(ctx, ownerID)
	if err != nil {
		return 0, err
	}

	// Update cache
	s.mu.Lock()
	s.usageCache[ownerID] = usage
	s.mu.Unlock()

	return usage, nil
}

// calculateUsage calculates total storage usage for an owner
func (s *quotaService) calculateUsage(ctx context.Context, ownerID string) (int64, error) {
	totalSize, err := s.fileRepo.SumSizeByOwner(ctx, ownerID)
	if err != nil {
		logger.Errorf(ctx, "Error calculating usage for owner %s: %v", ownerID, err)
		return 0, err
	}
	return totalSize, nil
}

// SetQuota sets storage quota for an owner
func (s *quotaService) SetQuota(ctx context.Context, ownerID string, quota int64) error {
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	s.mu.Lock()
	s.quotaCache[ownerID] = quota
	s.mu.Unlock()

	if s.redis != nil {
		key := fmt.Sprintf("resource_storage:quota:%s", ownerID)
		err := s.redis.Set(ctx, key, quota, 0).Err()
		if err != nil {
			logger.Errorf(ctx, "Error setting quota in Redis for owner %s: %v", ownerID, err)
			return err
		}
	}

	return nil
}

// GetQuota returns storage quota for an owner
func (s *quotaService) GetQuota(ctx context.Context, ownerID string) (int64, error) {
	if ownerID == "" {
		return 0, fmt.Errorf("owner ID is required")
	}

	// Check cache
	s.mu.RLock()
	if quota, found := s.quotaCache[ownerID]; found {
		s.mu.RUnlock()
		return quota, nil
	}
	s.mu.RUnlock()

	// Check Redis
	var quota int64
	if s.redis != nil {
		key := fmt.Sprintf("resource_storage:quota:%s", ownerID)
		val, err := s.redis.Get(ctx, key).Int64()
		if err == nil {
			s.mu.Lock()
			s.quotaCache[ownerID] = val
			s.mu.Unlock()
			return val, nil
		}
	}

	// Use default quota
	quota = s.config.DefaultQuota

	// Update cache and Redis
	s.mu.Lock()
	s.quotaCache[ownerID] = quota
	s.mu.Unlock()

	if s.redis != nil {
		key := fmt.Sprintf("resource_storage:quota:%s", ownerID)
		s.redis.Set(ctx, key, quota, 0)
	}

	return quota, nil
}

// IsQuotaExceeded checks if quota is exceeded
func (s *quotaService) IsQuotaExceeded(ctx context.Context, ownerID string) (bool, error) {
	if ownerID == "" {
		return false, fmt.Errorf("owner ID is required")
	}

	usage, err := s.GetUsage(ctx, ownerID)
	if err != nil {
		return false, err
	}

	quota, err := s.GetQuota(ctx, ownerID)
	if err != nil {
		return false, err
	}

	return usage >= quota, nil
}

// MonitorQuota monitors quotas for all owners
func (s *quotaService) MonitorQuota(ctx context.Context) error {
	owners, err := s.fileRepo.GetAllOwners(ctx)
	if err != nil {
		logger.Errorf(ctx, "Error getting owners for quota monitoring: %v", err)
		return err
	}

	for _, ownerID := range owners {
		usage, err := s.GetUsage(ctx, ownerID)
		if err != nil {
			logger.Errorf(ctx, "Error getting usage for owner %s: %v", ownerID, err)
			continue
		}

		quota, err := s.GetQuota(ctx, ownerID)
		if err != nil {
			logger.Errorf(ctx, "Error getting quota for owner %s: %v", ownerID, err)
			continue
		}

		usagePercent := float64(usage) / float64(quota) * 100

		if usage >= quota && s.publisher != nil {
			eventData := &event.StorageQuotaEventData{
				SpaceID:      ownerID,
				CurrentUsage: usage,
				Quota:        quota,
				UsagePercent: usagePercent,
				StorageType:  "file",
			}
			s.publisher.PublishStorageQuotaExceeded(ctx, eventData)
		} else if usagePercent >= s.config.WarningThreshold*100 && s.publisher != nil {
			eventData := &event.StorageQuotaEventData{
				SpaceID:      ownerID,
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

// UpdateUsage updates quota usage for external calls
func (s *quotaService) UpdateUsage(ctx context.Context, ownerID string, quotaType string, delta int64) error {
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}

	s.mu.Lock()
	if currentUsage, exists := s.usageCache[ownerID]; exists {
		newUsage := currentUsage + delta
		if newUsage < 0 {
			newUsage = 0
		}
		s.usageCache[ownerID] = newUsage
	}
	s.mu.Unlock()

	if s.redis != nil {
		key := fmt.Sprintf("storage:usage:%s", ownerID)
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

// RefreshSpaceServices refreshes space service references (placeholder)
func (s *quotaService) RefreshSpaceServices() {
	// Placeholder for space service integration
}
