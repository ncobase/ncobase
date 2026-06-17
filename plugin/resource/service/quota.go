package service

import (
	"context"
	"fmt"
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/event"
	"ncobase/plugin/resource/structs"
	"ncobase/plugin/resource/wrapper"
	"sync"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/redis/go-redis/v9"
)

// QuotaServiceInterface defines quota management methods
type QuotaServiceInterface interface {
	CheckAndUpdateQuota(ctx context.Context, ownerID string, size int) (bool, error)
	GetUsage(ctx context.Context, ownerID string) (int64, error)
	GetFileCount(ctx context.Context, ownerID string) (int, error)
	SetQuota(ctx context.Context, ownerID string, quota int64) error
	GetQuota(ctx context.Context, ownerID string) (int64, error)
	IsQuotaExceeded(ctx context.Context, ownerID string) (bool, error)
	MonitorQuota(ctx context.Context) error
	UpdateUsage(ctx context.Context, ownerID string, quotaType string, delta int64) error
	RefreshUsage(ctx context.Context, ownerID string) (int64, error)
	RefreshSpaceUsage(ctx context.Context, spaceID string) (int64, error)
	RefreshSpaceServices()
}

// QuotaConfig represents quota configuration
type QuotaConfig struct {
	DefaultQuota      int64         `json:"default_quota"`
	WarningThreshold  float64       `json:"warning_threshold"`
	CheckInterval     time.Duration `json:"check_interval"`
	EnableQuotas      bool          `json:"enable_quotas"`
	EnableEnforcement bool          `json:"enable_enforcement"`
}

type quotaService struct {
	fileRepo       repository.FileRepositoryInterface
	redis          *redis.Client
	configProvider ResourceConfigProvider
	publisher      event.PublisherInterface
	space          *wrapper.SpaceServiceWrapper
	quotaCache     map[string]int64
	usageCache     map[string]int64
	mu             sync.RWMutex
}

// NewQuotaService creates new quota service
func NewQuotaService(
	d *data.Data,
	publisher event.PublisherInterface,
	configProvider ResourceConfigProvider,
	spaceWrapper *wrapper.SpaceServiceWrapper,
) QuotaServiceInterface {
	if configProvider == nil {
		configProvider = NewDefaultConfigProvider()
	}

	return &quotaService{
		fileRepo:       repository.NewFileRepository(d),
		redis:          d.GetRedis().(*redis.Client),
		configProvider: configProvider,
		publisher:      publisher,
		space:          spaceWrapper,
		quotaCache:     make(map[string]int64),
		usageCache:     make(map[string]int64),
	}
}

func (s *quotaService) currentConfig(ctx context.Context) *QuotaConfig {
	if s.configProvider != nil {
		return s.configProvider.QuotaConfig(ctx)
	}
	return NewDefaultConfigProvider().QuotaConfig(ctx)
}

// CheckAndUpdateQuota checks and updates quota usage
func (s *quotaService) CheckAndUpdateQuota(ctx context.Context, ownerID string, size int) (bool, error) {
	cfg := s.currentConfig(ctx)
	if !cfg.EnableQuotas {
		return true, nil
	}
	if ownerID == "" {
		return false, fmt.Errorf("owner ID is required")
	}

	spaceID := ctxutil.GetSpaceID(ctx)
	if cfg.EnableEnforcement {
		allowed, err := s.checkSpaceQuota(ctx, spaceID, int64(size))
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, fmt.Errorf("storage quota exceeded for space %s", spaceID)
		}
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
	if cfg.EnableEnforcement && quota > 0 && newUsage > quota {
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

	if _, err := s.applyOwnerUsageDelta(ctx, ownerID, int64(size)); err != nil {
		return false, err
	}
	if err := s.applySpaceUsageDelta(ctx, spaceID, int64(size)); err != nil {
		if _, rollbackErr := s.applyOwnerUsageDelta(ctx, ownerID, -int64(size)); rollbackErr != nil {
			logger.Warnf(ctx, "Failed to rollback owner quota after space quota update failure: %v", rollbackErr)
		}
		return false, err
	}

	// Check warning threshold
	if quota > 0 && float64(newUsage)/float64(quota) >= cfg.WarningThreshold && s.publisher != nil {
		eventData := &event.StorageQuotaEventData{
			SpaceID:      ownerID,
			CurrentUsage: newUsage,
			Quota:        quota,
			UsagePercent: float64(newUsage) / float64(quota) * 100,
			StorageType:  "file",
		}
		s.publisher.PublishStorageQuotaWarning(ctx, eventData)
	}
	s.publishSpaceQuotaWarningIfNeeded(ctx, spaceID, cfg.WarningThreshold)

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

// GetFileCount returns the number of files owned by an owner.
func (s *quotaService) GetFileCount(ctx context.Context, ownerID string) (int, error) {
	if ownerID == "" {
		return 0, fmt.Errorf("owner ID is required")
	}
	return s.fileRepo.CountX(ctx, &structs.ListFileParams{OwnerID: ownerID}), nil
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

	return s.currentConfig(ctx).DefaultQuota, nil
}

// IsQuotaExceeded checks if quota is exceeded
func (s *quotaService) IsQuotaExceeded(ctx context.Context, ownerID string) (bool, error) {
	if !s.currentConfig(ctx).EnableQuotas {
		return false, nil
	}
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
	cfg := s.currentConfig(ctx)
	if !cfg.EnableQuotas {
		return nil
	}
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

		if quota <= 0 {
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
		} else if usagePercent >= cfg.WarningThreshold*100 && s.publisher != nil {
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

	if s.space != nil && s.space.HasSpaceQuotaService() {
		spaces, err := s.fileRepo.GetAllSpaces(ctx)
		if err != nil {
			logger.Errorf(ctx, "Error getting spaces for quota monitoring: %v", err)
			return err
		}

		for _, spaceID := range spaces {
			usage, err := s.RefreshSpaceUsage(ctx, spaceID)
			if err != nil {
				logger.Errorf(ctx, "Error refreshing storage usage for space %s: %v", spaceID, err)
				continue
			}

			quota, err := s.space.GetQuota(ctx, spaceID, "storage")
			if err != nil {
				logger.Errorf(ctx, "Error getting storage quota for space %s: %v", spaceID, err)
				continue
			}
			if quota <= 0 {
				continue
			}

			usagePercent := float64(usage) / float64(quota) * 100
			if usage >= quota && s.publisher != nil {
				s.publisher.PublishStorageQuotaExceeded(ctx, &event.StorageQuotaEventData{
					SpaceID:      spaceID,
					CurrentUsage: usage,
					Quota:        quota,
					UsagePercent: usagePercent,
					StorageType:  "file",
				})
			} else if usagePercent >= cfg.WarningThreshold*100 && s.publisher != nil {
				s.publisher.PublishStorageQuotaWarning(ctx, &event.StorageQuotaEventData{
					SpaceID:      spaceID,
					CurrentUsage: usage,
					Quota:        quota,
					UsagePercent: usagePercent,
					StorageType:  "file",
				})
			}
		}
	}

	return nil
}

// UpdateUsage updates quota usage for external calls
func (s *quotaService) UpdateUsage(ctx context.Context, ownerID string, quotaType string, delta int64) error {
	if ownerID == "" {
		return fmt.Errorf("owner ID is required")
	}
	if quotaType != "" && quotaType != "storage" {
		return nil
	}

	if _, err := s.applyOwnerUsageDelta(ctx, ownerID, delta); err != nil {
		return err
	}

	if err := s.applySpaceUsageDelta(ctx, ctxutil.GetSpaceID(ctx), delta); err != nil {
		return err
	}

	return nil
}

// RefreshUsage recalculates and stores authoritative owner usage from file records.
func (s *quotaService) RefreshUsage(ctx context.Context, ownerID string) (int64, error) {
	if ownerID == "" {
		return 0, fmt.Errorf("owner ID is required")
	}

	usage, err := s.calculateUsage(ctx, ownerID)
	if err != nil {
		return 0, err
	}
	if err := s.setOwnerUsage(ctx, ownerID, usage); err != nil {
		return 0, err
	}
	return usage, nil
}

// RefreshSpaceUsage recalculates and stores authoritative storage usage for a space.
func (s *quotaService) RefreshSpaceUsage(ctx context.Context, spaceID string) (int64, error) {
	if spaceID == "" {
		return 0, nil
	}

	usage, err := s.fileRepo.SumSizeBySpace(ctx, spaceID)
	if err != nil {
		return 0, err
	}
	if s.space == nil || !s.space.HasSpaceQuotaService() {
		return usage, nil
	}

	currentUsage, err := s.space.GetUsage(ctx, spaceID, "storage")
	if err != nil {
		return 0, err
	}
	delta := usage - currentUsage
	if delta != 0 {
		if err := s.space.UpdateUsage(ctx, spaceID, "storage", delta); err != nil {
			return 0, err
		}
	}
	return usage, nil
}

func (s *quotaService) checkSpaceQuota(ctx context.Context, spaceID string, size int64) (bool, error) {
	if spaceID == "" || s.space == nil || !s.space.HasSpaceQuotaService() {
		return true, nil
	}

	allowed, err := s.space.CheckQuotaLimit(ctx, spaceID, "storage", size)
	if err != nil {
		return false, fmt.Errorf("failed to check storage quota for space %s: %w", spaceID, err)
	}
	if allowed {
		return true, nil
	}

	if s.publisher != nil {
		usage, _ := s.space.GetUsage(ctx, spaceID, "storage")
		quota, _ := s.space.GetQuota(ctx, spaceID, "storage")
		usagePercent := 0.0
		if quota > 0 {
			usagePercent = float64(usage) / float64(quota) * 100
		}
		s.publisher.PublishStorageQuotaExceeded(ctx, &event.StorageQuotaEventData{
			SpaceID:      spaceID,
			CurrentUsage: usage,
			Quota:        quota,
			UsagePercent: usagePercent,
			StorageType:  "file",
		})
	}

	return false, nil
}

func (s *quotaService) applyOwnerUsageDelta(ctx context.Context, ownerID string, delta int64) (int64, error) {
	currentUsage, found := s.cachedOwnerUsage(ctx, ownerID)
	if !found {
		var err error
		currentUsage, err = s.calculateUsage(ctx, ownerID)
		if err != nil {
			return 0, err
		}
	}

	newUsage := currentUsage + delta
	if newUsage < 0 {
		newUsage = 0
	}
	if err := s.setOwnerUsage(ctx, ownerID, newUsage); err != nil {
		return 0, err
	}
	return newUsage, nil
}

func (s *quotaService) cachedOwnerUsage(ctx context.Context, ownerID string) (int64, bool) {
	s.mu.RLock()
	if usage, found := s.usageCache[ownerID]; found {
		s.mu.RUnlock()
		return usage, true
	}
	s.mu.RUnlock()

	if s.redis == nil {
		return 0, false
	}

	key := fmt.Sprintf("storage:usage:%s", ownerID)
	usage, err := s.redis.Get(ctx, key).Int64()
	if err != nil {
		return 0, false
	}

	s.mu.Lock()
	s.usageCache[ownerID] = usage
	s.mu.Unlock()
	return usage, true
}

func (s *quotaService) setOwnerUsage(ctx context.Context, ownerID string, usage int64) error {
	s.mu.Lock()
	s.usageCache[ownerID] = usage
	s.mu.Unlock()

	if s.redis == nil {
		return nil
	}
	key := fmt.Sprintf("storage:usage:%s", ownerID)
	if err := s.redis.Set(ctx, key, usage, 0).Err(); err != nil {
		logger.Errorf(ctx, "Error setting usage in Redis for owner %s: %v", ownerID, err)
		return err
	}
	return nil
}

func (s *quotaService) applySpaceUsageDelta(ctx context.Context, spaceID string, delta int64) error {
	if spaceID == "" || s.space == nil || !s.space.HasSpaceQuotaService() || delta == 0 {
		return nil
	}
	if err := s.space.UpdateUsage(ctx, spaceID, "storage", delta); err != nil {
		return fmt.Errorf("failed to update storage usage for space %s: %w", spaceID, err)
	}
	return nil
}

func (s *quotaService) publishSpaceQuotaWarningIfNeeded(ctx context.Context, spaceID string, warningThreshold float64) {
	if spaceID == "" || s.space == nil || !s.space.HasSpaceQuotaService() || s.publisher == nil {
		return
	}
	usage, err := s.space.GetUsage(ctx, spaceID, "storage")
	if err != nil {
		logger.Warnf(ctx, "Failed to get storage usage for space %s warning check: %v", spaceID, err)
		return
	}
	quota, err := s.space.GetQuota(ctx, spaceID, "storage")
	if err != nil {
		logger.Warnf(ctx, "Failed to get storage quota for space %s warning check: %v", spaceID, err)
		return
	}
	if quota <= 0 {
		return
	}
	usagePercent := float64(usage) / float64(quota) * 100
	if usagePercent >= warningThreshold*100 {
		s.publisher.PublishStorageQuotaWarning(ctx, &event.StorageQuotaEventData{
			SpaceID:      spaceID,
			CurrentUsage: usage,
			Quota:        quota,
			UsagePercent: usagePercent,
			StorageType:  "file",
		})
	}
}

// RefreshSpaceServices refreshes space service references.
func (s *quotaService) RefreshSpaceServices() {
	if s.space != nil {
		s.space.RefreshServices()
	}
}
