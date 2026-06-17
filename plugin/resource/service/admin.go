package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/structs"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// ErrResourceAdminOperationUnavailable marks routes whose durable backend is not configured yet.
var ErrResourceAdminOperationUnavailable = errors.New("resource admin operation is unavailable")

// AdminServiceInterface defines admin service methods
type AdminServiceInterface interface {
	// File management
	ListFiles(ctx context.Context, params *structs.AdminFileListParams) (*structs.AdminFileListResponse, error)
	DeleteFile(ctx context.Context, slug string) error
	SetFileStatus(ctx context.Context, slug string, req *structs.AdminSetStatusRequest) (*structs.FileResponse, error)

	// Statistics
	GetStorageStats(ctx context.Context) (*structs.StorageStats, error)
	GetUsageStats(ctx context.Context, period string) (*structs.UsageStats, error)
	GetActivityStats(ctx context.Context) (*structs.ActivityStats, error)

	// Quota management
	ListQuotas(ctx context.Context, params *structs.AdminQuotaListParams) (*structs.AdminQuotaListResponse, error)
	SetQuota(ctx context.Context, userID string, req *structs.QuotaSetRequest) (*structs.QuotaInfo, error)
	GetQuota(ctx context.Context, userID string) (*structs.QuotaInfo, error)
	DeleteQuota(ctx context.Context, userID string) error

	// Batch operations
	BatchCleanup(ctx context.Context, req *structs.BatchCleanupRequest) (*structs.BatchCleanupResult, error)
	ListBatchJobs(ctx context.Context, params *structs.AdminBatchJobParams) (*structs.BatchJobListResponse, error)
	CancelBatchJob(ctx context.Context, jobID string) error

	// Storage management
	OptimizeStorage(ctx context.Context) (*structs.OptimizeResult, error)
	GetStorageHealth(ctx context.Context) (*structs.StorageHealth, error)
	InitiateBackup(ctx context.Context, req *structs.BackupRequest) (*structs.BackupResult, error)
}

type adminService struct {
	data         *data.Data
	fileRepo     repository.FileRepositoryInterface
	quotaService QuotaServiceInterface
	batchJobs    map[string]*structs.BatchJob
	batchJobsMu  sync.RWMutex
}

// NewAdminService creates new admin service
func NewAdminService(d *data.Data, quotaService QuotaServiceInterface) AdminServiceInterface {
	return &adminService{
		data:         d,
		fileRepo:     repository.NewFileRepository(d),
		quotaService: quotaService,
		batchJobs:    make(map[string]*structs.BatchJob),
	}
}

// ListFiles lists all files for admin view
func (s *adminService) ListFiles(ctx context.Context, params *structs.AdminFileListParams) (*structs.AdminFileListResponse, error) {
	listParams := &structs.ListFileParams{
		Limit: params.Limit,
		User:  params.UserID,
	}

	if listParams.Limit == 0 {
		listParams.Limit = 50
	}

	files, err := s.fileRepo.List(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	total := s.fileRepo.CountX(ctx, listParams)

	readFiles := make([]*structs.ReadFile, len(files))
	for i, file := range files {
		readFiles[i] = repository.SerializeFile(file)
	}

	stats := s.calculateFileStats(readFiles)

	return &structs.AdminFileListResponse{
		Files: readFiles,
		Total: total,
		Stats: stats,
	}, nil
}

// DeleteFile deletes a file with admin privileges
func (s *adminService) DeleteFile(ctx context.Context, slug string) error {
	file, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to get file before deletion: %w", err)
	}

	err = s.fileRepo.Delete(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	s.refreshQuotaAfterFileRemoval(ctx, file.OwnerID, file.Extras)
	return nil
}

// SetFileStatus sets file status with admin privileges
func (s *adminService) SetFileStatus(ctx context.Context, slug string, req *structs.AdminSetStatusRequest) (*structs.FileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	extras := repository.CloneExtras(file.Extras)

	statusChange := structs.StatusChange{
		Status:    req.Status,
		Reason:    req.Reason,
		ChangedBy: firstNonEmpty(ctxutil.GetUserID(ctx), "admin"),
		ChangedAt: time.Now().UnixMilli(),
	}

	statusHistory := statusHistoryFromExtras(extras["status_history"])
	statusHistory = append(statusHistory, statusChange)

	extras["status"] = req.Status
	extras["status_history"] = statusHistory

	_, err = s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	updatedFile, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated file: %w", err)
	}

	readFile := repository.SerializeFile(updatedFile)
	return &structs.FileResponse{
		ReadFile:      readFile,
		StatusHistory: statusHistory,
	}, nil
}

// GetStorageStats gets storage statistics
func (s *adminService) GetStorageStats(ctx context.Context) (*structs.StorageStats, error) {
	totalSize, err := s.fileRepo.SumSizeByOwner(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}

	totalFiles := s.fileRepo.CountX(ctx, &structs.ListFileParams{})
	owners, err := s.fileRepo.GetAllOwners(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage owners: %w", err)
	}
	byCategory, err := s.fileRepo.AggregateSizeByCategory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate category storage: %w", err)
	}
	byStorage, err := s.fileRepo.AggregateSizeByStorage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate storage providers: %w", err)
	}
	dailyUploads, err := s.getDailyUploads(ctx)
	if err != nil {
		return nil, err
	}
	topUsers, err := s.getTopUsers(ctx)
	if err != nil {
		return nil, err
	}
	storageHealth, err := s.resolveStorageHealth(ctx, totalSize, owners)
	if err != nil {
		return nil, err
	}

	return &structs.StorageStats{
		TotalSize:     totalSize,
		TotalFiles:    totalFiles,
		TotalUsers:    len(owners),
		ByCategory:    byCategory,
		ByStorage:     byStorage,
		DailyUploads:  dailyUploads,
		TopUsers:      topUsers,
		StorageHealth: storageHealth,
	}, nil
}

// GetUsageStats gets usage statistics for a period
func (s *adminService) GetUsageStats(ctx context.Context, period string) (*structs.UsageStats, error) {
	totalSize, err := s.fileRepo.SumSizeByOwner(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}
	totalFiles := s.fileRepo.CountX(ctx, &structs.ListFileParams{})
	from, to, previousFrom, previousTo := usagePeriodBounds(period)
	currentSize, currentFiles, err := s.fileRepo.AggregateUsageBetween(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate current usage: %w", err)
	}
	previousSize, previousFiles, err := s.fileRepo.AggregateUsageBetween(ctx, previousFrom, previousTo)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate previous usage: %w", err)
	}
	breakdown, err := s.getUsageBreakdown(ctx, from, to)
	if err != nil {
		return nil, err
	}

	return &structs.UsageStats{
		Period:     period,
		TotalSize:  totalSize,
		TotalFiles: totalFiles,
		Growth: &structs.GrowthStats{
			SizeGrowth:  growthPercent(currentSize, previousSize),
			FilesGrowth: growthPercent(int64(currentFiles), int64(previousFiles)),
		},
		Breakdown: breakdown,
	}, nil
}

// GetActivityStats gets activity statistics
func (s *adminService) GetActivityStats(ctx context.Context) (*structs.ActivityStats, error) {
	return &structs.ActivityStats{
		TotalDownloads:     0,
		TotalViews:         0,
		PopularFiles:       []structs.PopularFile{},
		ActivityByHour:     []structs.HourlyActivity{},
		TelemetryAvailable: false,
		Message:            "Download and view telemetry is not configured; returned counters are unavailable rather than aggregated activity.",
	}, nil
}

// ListQuotas lists all user quotas
func (s *adminService) ListQuotas(ctx context.Context, params *structs.AdminQuotaListParams) (*structs.AdminQuotaListResponse, error) {
	users, err := s.fileRepo.GetAllOwners(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	quotas := make([]*structs.QuotaInfo, 0, len(users))
	for _, userID := range users {
		quota, _ := s.quotaService.GetQuota(ctx, userID)
		usage, _ := s.quotaService.GetUsage(ctx, userID)

		usagePercent := float64(0)
		if quota > 0 {
			usagePercent = float64(usage) / float64(quota) * 100
		}

		quotas = append(quotas, &structs.QuotaInfo{
			UserID:       userID,
			Quota:        quota,
			Usage:        usage,
			UsagePercent: usagePercent,
			FileCount:    s.fileRepo.CountX(ctx, &structs.ListFileParams{OwnerID: userID}),
		})
	}

	return &structs.AdminQuotaListResponse{
		Quotas: quotas,
		Total:  len(quotas),
	}, nil
}

// SetQuota sets quota for a user
func (s *adminService) SetQuota(ctx context.Context, userID string, req *structs.QuotaSetRequest) (*structs.QuotaInfo, error) {
	err := s.quotaService.SetQuota(ctx, userID, req.Quota)
	if err != nil {
		return nil, fmt.Errorf("failed to set quota: %w", err)
	}

	usage, _ := s.quotaService.GetUsage(ctx, userID)
	usagePercent := float64(0)
	if req.Quota > 0 {
		usagePercent = float64(usage) / float64(req.Quota) * 100
	}

	return &structs.QuotaInfo{
		UserID:       userID,
		Quota:        req.Quota,
		Usage:        usage,
		UsagePercent: usagePercent,
		FileCount:    s.fileRepo.CountX(ctx, &structs.ListFileParams{OwnerID: userID}),
	}, nil
}

// GetQuota gets quota for a user
func (s *adminService) GetQuota(ctx context.Context, userID string) (*structs.QuotaInfo, error) {
	quota, err := s.quotaService.GetQuota(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	usage, _ := s.quotaService.GetUsage(ctx, userID)
	usagePercent := float64(0)
	if quota > 0 {
		usagePercent = float64(usage) / float64(quota) * 100
	}

	return &structs.QuotaInfo{
		UserID:       userID,
		Quota:        quota,
		Usage:        usage,
		UsagePercent: usagePercent,
		FileCount:    s.fileRepo.CountX(ctx, &structs.ListFileParams{OwnerID: userID}),
	}, nil
}

// DeleteQuota deletes quota for a user
func (s *adminService) DeleteQuota(ctx context.Context, userID string) error {
	defaultQuota := int64(10 * 1024 * 1024 * 1024) // 10GB
	return s.quotaService.SetQuota(ctx, userID, defaultQuota)
}

// BatchCleanup performs batch cleanup operations
func (s *adminService) BatchCleanup(ctx context.Context, req *structs.BatchCleanupRequest) (*structs.BatchCleanupResult, error) {
	if req == nil {
		return nil, fmt.Errorf("cleanup request is required")
	}
	switch req.Type {
	case "expired", "orphaned", "duplicates":
	default:
		return nil, fmt.Errorf("unknown cleanup type: %s", req.Type)
	}

	jobID := uuid.New().String()
	startedAt := time.Now().UnixMilli()
	job := &structs.BatchJob{
		ID:        jobID,
		Type:      req.Type,
		Status:    "processing",
		Progress:  0,
		StartedAt: startedAt,
		CreatedBy: firstNonEmpty(ctxutil.GetUserID(ctx), "system"),
	}
	s.storeBatchJob(job)

	result := &structs.BatchCleanupResult{
		JobID:          jobID,
		Type:           req.Type,
		ItemsFound:     0,
		ItemsCleaned:   0,
		SpaceFreed:     0,
		DryRun:         req.DryRun,
		CandidateItems: make([]string, 0),
		CleanedItems:   make([]string, 0),
		Errors:         make([]string, 0),
	}

	switch req.Type {
	case "expired":
		result = s.cleanupExpiredFiles(ctx, req, result)
	case "orphaned":
		result = s.cleanupOrphanedFiles(ctx, req, result)
	case "duplicates":
		result = s.cleanupDuplicateFiles(ctx, req, result)
	}

	completedAt := time.Now().UnixMilli()
	status := "completed"
	if len(result.Errors) > 0 {
		status = "partial_failure"
		if result.ItemsCleaned == 0 {
			status = "failed"
		}
	}
	job.Status = status
	job.Progress = 100
	job.ItemCount = result.ItemsFound
	job.ProcessedCount = result.ItemsCleaned
	job.ErrorCount = len(result.Errors)
	job.CompletedAt = &completedAt
	job.Result = cleanupResultJSON(result)
	job.Errors = append([]string(nil), result.Errors...)
	s.storeBatchJob(job)

	return result, nil
}

// ListBatchJobs lists batch jobs
func (s *adminService) ListBatchJobs(ctx context.Context, params *structs.AdminBatchJobParams) (*structs.BatchJobListResponse, error) {
	s.batchJobsMu.RLock()
	jobs := make([]*structs.BatchJob, 0, len(s.batchJobs))
	for _, job := range s.batchJobs {
		if params.Status == "" || job.Status == params.Status {
			jobs = append(jobs, cloneBatchJob(job))
		}
	}
	s.batchJobsMu.RUnlock()

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartedAt > jobs[j].StartedAt
	})

	if params.Limit > 0 && len(jobs) > params.Limit {
		jobs = jobs[:params.Limit]
	}

	return &structs.BatchJobListResponse{
		Jobs:  jobs,
		Total: len(jobs),
	}, nil
}

// CancelBatchJob cancels a batch job
func (s *adminService) CancelBatchJob(ctx context.Context, jobID string) error {
	s.batchJobsMu.RLock()
	job, exists := s.batchJobs[jobID]
	s.batchJobsMu.RUnlock()
	if !exists {
		return fmt.Errorf("batch job not found")
	}

	if job.Status == "processing" {
		return fmt.Errorf("%w: admin cleanup jobs run synchronously and cannot be cancelled after the request has started", ErrResourceAdminOperationUnavailable)
	}
	if job.Status == "completed" || job.Status == "cancelled" || job.Status == "failed" || job.Status == "partial_failure" {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	s.batchJobsMu.Lock()
	job.Status = "cancelled"
	s.batchJobsMu.Unlock()
	return nil
}

// OptimizeStorage optimizes storage system
func (s *adminService) OptimizeStorage(ctx context.Context) (*structs.OptimizeResult, error) {
	taskID := uuid.New().String()
	startedAt := time.Now()

	duplicates, err := s.fileRepo.FindDuplicateFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect duplicate files: %w", err)
	}
	deduplicatedFiles := 0
	var spaceFreed int64
	for _, group := range duplicates {
		if len(group) <= 1 {
			continue
		}
		for _, duplicate := range group[1:] {
			deduplicatedFiles++
			spaceFreed += int64(duplicate.Size)
		}
	}

	orphanedFiles, err := s.fileRepo.FindOrphanedFiles(ctx, nil, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect orphaned files: %w", err)
	}

	return &structs.OptimizeResult{
		TaskID:                  taskID,
		Mode:                    "analysis",
		DeduplicatedFiles:       0,
		SpaceFreed:              0,
		OrphanedCleaned:         0,
		IndexesRebuilt:          0,
		PotentialDuplicateFiles: deduplicatedFiles,
		PotentialSpaceFreed:     spaceFreed,
		OrphanedFiles:           len(orphanedFiles),
		PerformedActions:        []string{"metadata_scan", "duplicate_hash_analysis", "orphan_metadata_analysis"},
		Duration:                int64(time.Since(startedAt).Seconds()),
	}, nil
}

// GetStorageHealth gets storage health status
func (s *adminService) GetStorageHealth(ctx context.Context) (*structs.StorageHealth, error) {
	totalSize, err := s.fileRepo.SumSizeByOwner(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}
	owners, err := s.fileRepo.GetAllOwners(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage owners: %w", err)
	}
	totalQuota, exceededOwners, err := s.aggregateOwnerQuotaHealth(ctx, owners)
	if err != nil {
		return nil, err
	}
	orphanedFiles, err := s.fileRepo.FindOrphanedFiles(ctx, nil, 10000)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect orphaned files: %w", err)
	}

	status := "healthy"
	if exceededOwners > 0 {
		status = "critical"
	} else if totalQuota > 0 && float64(totalSize)/float64(totalQuota) >= 0.8 {
		status = "warning"
	}
	usagePercent := 0.0
	freeSpace := int64(0)
	if totalQuota > 0 {
		usagePercent = float64(totalSize) / float64(totalQuota) * 100
		freeSpace = totalQuota - totalSize
		if freeSpace < 0 {
			freeSpace = 0
		}
	}

	recommendations := []string{}
	if exceededOwners > 0 {
		recommendations = append(recommendations, "Review owners that exceed configured storage quota.")
	}
	if len(orphanedFiles) > 0 {
		recommendations = append(recommendations, "Review orphaned resource records before cleanup.")
	}

	return &structs.StorageHealth{
		Status:         status,
		TotalSpace:     totalQuota,
		UsedSpace:      totalSize,
		FreeSpace:      freeSpace,
		UsagePercent:   usagePercent,
		OrphanedFiles:  len(orphanedFiles),
		CorruptedFiles: 0,
		HealthChecks: []structs.HealthCheck{
			{
				Name:    "Resource Metadata",
				Status:  "ok",
				Message: "Resource metadata queries completed successfully",
				LastRun: time.Now().UnixMilli(),
			},
			{
				Name:    "Quota Utilization",
				Status:  status,
				Message: fmt.Sprintf("%d owner quota records inspected", len(owners)),
				LastRun: time.Now().UnixMilli(),
			},
		},
		Recommendations: recommendations,
	}, nil
}

// InitiateBackup initiates storage backup
func (s *adminService) InitiateBackup(ctx context.Context, req *structs.BackupRequest) (*structs.BackupResult, error) {
	return nil, fmt.Errorf("%w: storage backup requires a configured durable backup provider and object storage export implementation", ErrResourceAdminOperationUnavailable)
}

// calculateFileStats calculates file stats
func (s *adminService) calculateFileStats(files []*structs.ReadFile) *structs.FileStats {
	stats := &structs.FileStats{
		ByCategory: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	for _, file := range files {
		if file.Size != nil {
			stats.TotalSize += int64(*file.Size)
		}
		stats.TotalCount++
	}

	return stats
}

// getDailyUploads gets daily uploads
func (s *adminService) getDailyUploads(ctx context.Context) ([]structs.DailyUpload, error) {
	from, to, _, _ := usagePeriodBounds("30d")
	daily, err := s.fileRepo.AggregateDailyUsage(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate daily uploads: %w", err)
	}

	result := make([]structs.DailyUpload, 0, len(daily))
	for _, item := range daily {
		result = append(result, structs.DailyUpload{
			Date:  item.Date,
			Count: item.Files,
			Size:  item.Size,
		})
	}
	return result, nil
}

// getTopUsers gets top users
func (s *adminService) getTopUsers(ctx context.Context) ([]structs.UserUsage, error) {
	owners, err := s.fileRepo.AggregateOwnerUsage(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate top storage owners: %w", err)
	}

	result := make([]structs.UserUsage, 0, len(owners))
	for _, owner := range owners {
		result = append(result, structs.UserUsage{
			UserID: owner.OwnerID,
			Size:   owner.Size,
			Files:  owner.Files,
		})
	}
	return result, nil
}

// getUsageBreakdown gets usage breakdown
func (s *adminService) getUsageBreakdown(ctx context.Context, from, to int64) ([]structs.UsageByDate, error) {
	daily, err := s.fileRepo.AggregateDailyUsage(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate usage breakdown: %w", err)
	}

	result := make([]structs.UsageByDate, 0, len(daily))
	for _, item := range daily {
		result = append(result, structs.UsageByDate{
			Date:  item.Date,
			Size:  item.Size,
			Files: item.Files,
		})
	}
	return result, nil
}

func usagePeriodBounds(period string) (from, to, previousFrom, previousTo int64) {
	now := time.Now().UTC()
	var duration time.Duration
	switch period {
	case "24h", "1d":
		duration = 24 * time.Hour
	case "7d", "week":
		duration = 7 * 24 * time.Hour
	case "90d", "quarter":
		duration = 90 * 24 * time.Hour
	case "365d", "year":
		duration = 365 * 24 * time.Hour
	default:
		duration = 30 * 24 * time.Hour
	}
	start := now.Add(-duration)
	previousStart := start.Add(-duration)
	return start.UnixMilli(), now.UnixMilli(), previousStart.UnixMilli(), start.UnixMilli()
}

func growthPercent(current, previous int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return (float64(current-previous) / float64(previous)) * 100
}

func (s *adminService) resolveStorageHealth(ctx context.Context, totalSize int64, owners []string) (string, error) {
	totalQuota, exceededOwners, err := s.aggregateOwnerQuotaHealth(ctx, owners)
	if err != nil {
		return "", err
	}
	if exceededOwners > 0 {
		return "critical", nil
	}
	if totalQuota > 0 && float64(totalSize)/float64(totalQuota) >= 0.8 {
		return "warning", nil
	}
	return "healthy", nil
}

func (s *adminService) aggregateOwnerQuotaHealth(ctx context.Context, owners []string) (int64, int, error) {
	var totalQuota int64
	exceededOwners := 0
	for _, ownerID := range owners {
		quota, err := s.quotaService.GetQuota(ctx, ownerID)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get quota for owner %s: %w", ownerID, err)
		}
		usage, err := s.quotaService.GetUsage(ctx, ownerID)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get usage for owner %s: %w", ownerID, err)
		}
		if quota > 0 {
			totalQuota += quota
			if usage > quota {
				exceededOwners++
			}
		}
	}
	return totalQuota, exceededOwners, nil
}

func (s *adminService) refreshQuotaAfterFileRemoval(ctx context.Context, ownerID string, extras types.JSON) {
	if s.quotaService == nil {
		return
	}
	if ownerID != "" {
		if _, err := s.quotaService.RefreshUsage(ctx, ownerID); err != nil {
			logger.Warnf(ctx, "Failed to refresh owner quota after admin file deletion: %v", err)
		}
	}
	if spaceID := spaceIDFromExtras(extras); spaceID != "" {
		if _, err := s.quotaService.RefreshSpaceUsage(ctx, spaceID); err != nil {
			logger.Warnf(ctx, "Failed to refresh space quota after admin file deletion: %v", err)
		}
	}
}

func spaceIDFromExtras(extras types.JSON) string {
	if extras == nil {
		return ""
	}
	if value, ok := extras["space_id"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func statusHistoryFromExtras(value any) []structs.StatusChange {
	if value == nil {
		return nil
	}
	if history, ok := value.([]structs.StatusChange); ok {
		return append([]structs.StatusChange(nil), history...)
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var history []structs.StatusChange
	if err := json.Unmarshal(raw, &history); err != nil {
		return nil
	}
	return history
}

func cleanupResultJSON(result *structs.BatchCleanupResult) *types.JSON {
	if result == nil {
		return nil
	}
	payload := types.JSON{
		"job_id":                result.JobID,
		"type":                  result.Type,
		"items_found":           result.ItemsFound,
		"items_cleaned":         result.ItemsCleaned,
		"space_freed":           result.SpaceFreed,
		"potential_space_freed": result.PotentialSpaceFreed,
		"dry_run":               result.DryRun,
		"candidate_items":       result.CandidateItems,
		"cleaned_items":         result.CleanedItems,
		"errors":                result.Errors,
		"completed_at":          time.Now().UnixMilli(),
		"execution_mode":        "synchronous",
	}
	return &payload
}

func cloneBatchJob(job *structs.BatchJob) *structs.BatchJob {
	if job == nil {
		return nil
	}
	cloned := *job
	if job.CompletedAt != nil {
		completedAt := *job.CompletedAt
		cloned.CompletedAt = &completedAt
	}
	if job.Result != nil {
		result := make(types.JSON, len(*job.Result))
		for key, value := range *job.Result {
			result[key] = value
		}
		cloned.Result = &result
	}
	cloned.Errors = append([]string(nil), job.Errors...)
	return &cloned
}

func (s *adminService) storeBatchJob(job *structs.BatchJob) {
	if job == nil {
		return
	}
	s.batchJobsMu.Lock()
	defer s.batchJobsMu.Unlock()
	s.batchJobs[job.ID] = cloneBatchJob(job)
}

// cleanupExpiredFiles cleans up expired files via repository
func (s *adminService) cleanupExpiredFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	limit := req.MaxItems
	if limit <= 0 {
		limit = 1000
	}

	files, err := s.fileRepo.FindExpiredFiles(ctx, req.Filters, limit)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to query expired files: %v", err))
		return result
	}

	result.ItemsFound = len(files)

	for _, file := range files {
		result.CandidateItems = append(result.CandidateItems, file.ID)
		result.PotentialSpaceFreed += int64(file.Size)
		if !req.DryRun {
			if err := s.fileRepo.Delete(ctx, file.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to delete file %s: %v", file.ID, err))
				continue
			}
			s.refreshQuotaAfterFileRemoval(ctx, file.OwnerID, file.Extras)
			result.ItemsCleaned++
			result.SpaceFreed += int64(file.Size)
			result.CleanedItems = append(result.CleanedItems, file.ID)
		}
	}

	return result
}

// cleanupOrphanedFiles cleans up orphaned files via repository
func (s *adminService) cleanupOrphanedFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	limit := req.MaxItems
	if limit <= 0 {
		limit = 1000
	}

	files, err := s.fileRepo.FindOrphanedFiles(ctx, req.Filters, limit)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to query files: %v", err))
		return result
	}

	result.ItemsFound = len(files)

	for _, file := range files {
		result.CandidateItems = append(result.CandidateItems, file.ID)
		result.PotentialSpaceFreed += int64(file.Size)
		if !req.DryRun {
			if err := s.fileRepo.Delete(ctx, file.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to delete orphaned file %s: %v", file.ID, err))
				continue
			}
			s.refreshQuotaAfterFileRemoval(ctx, file.OwnerID, file.Extras)
			result.ItemsCleaned++
			result.SpaceFreed += int64(file.Size)
			result.CleanedItems = append(result.CleanedItems, file.ID)
		}
	}

	return result
}

// cleanupDuplicateFiles cleans up duplicate files via repository
func (s *adminService) cleanupDuplicateFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	hashGroups, err := s.fileRepo.FindDuplicateFiles(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to query files for deduplication: %v", err))
		return result
	}

	cleaned := 0
	limit := req.MaxItems
	if limit <= 0 {
		limit = 1000
	}

	for hash, group := range hashGroups {
		if len(group) <= 1 {
			continue
		}

		// Keep the first (oldest by created_at), delete the rest
		for _, duplicate := range group[1:] {
			if cleaned >= limit {
				break
			}

			if req.Filters != nil {
				if req.Filters.MinSize != nil && int64(duplicate.Size) < *req.Filters.MinSize {
					continue
				}
				if req.Filters.MaxSize != nil && int64(duplicate.Size) > *req.Filters.MaxSize {
					continue
				}
			}

			result.ItemsFound++
			result.CandidateItems = append(result.CandidateItems, duplicate.ID)
			result.PotentialSpaceFreed += int64(duplicate.Size)
			if !req.DryRun {
				if err := s.fileRepo.Delete(ctx, duplicate.ID); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to delete duplicate file %s (hash: %s): %v", duplicate.ID, hash, err))
					continue
				}
				s.refreshQuotaAfterFileRemoval(ctx, duplicate.OwnerID, duplicate.Extras)
				result.ItemsCleaned++
				result.SpaceFreed += int64(duplicate.Size)
				result.CleanedItems = append(result.CleanedItems, duplicate.ID)
			}
			cleaned++
		}

		if cleaned >= limit {
			break
		}
	}

	logger.Infof(ctx, "duplicate cleanup: found %d duplicates, cleaned %d, freed %d bytes",
		result.ItemsFound, result.ItemsCleaned, result.SpaceFreed)

	return result
}
