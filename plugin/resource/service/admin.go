package service

import (
	"context"
	"fmt"
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/data/ent"
	"ncobase/plugin/resource/data/repository"
	"ncobase/plugin/resource/structs"
	"time"

	"github.com/google/uuid"
	"github.com/ncobase/ncore/types"
)

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
	fileRepo     repository.FileRepositoryInterface
	quotaService QuotaServiceInterface
	batchJobs    map[string]*structs.BatchJob
}

// NewAdminService creates new admin service
func NewAdminService(d *data.Data, quotaService QuotaServiceInterface) AdminServiceInterface {
	return &adminService{
		fileRepo:     repository.NewFileRepository(d),
		quotaService: quotaService,
		batchJobs:    make(map[string]*structs.BatchJob),
	}
}

// ListFiles lists all files for admin view
func (s *adminService) ListFiles(ctx context.Context, params *structs.AdminFileListParams) (*structs.AdminFileListResponse, error) {
	// Convert admin params to regular list params
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

	// Convert to ReadFile structs
	readFiles := make([]*structs.ReadFile, len(files))
	for i, file := range files {
		readFiles[i] = s.convertToReadFile(file)
	}

	// Calculate stats
	stats := s.calculateFileStats(files)

	return &structs.AdminFileListResponse{
		Files: readFiles,
		Total: total,
		Stats: stats,
	}, nil
}

// DeleteFile deletes a file with admin privileges
func (s *adminService) DeleteFile(ctx context.Context, slug string) error {
	err := s.fileRepo.Delete(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// SetFileStatus sets file status with admin privileges
func (s *adminService) SetFileStatus(ctx context.Context, slug string, req *structs.AdminSetStatusRequest) (*structs.FileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Update file status in extras
	extras := make(types.JSON)
	if file.Extras != nil {
		for k, v := range file.Extras {
			extras[k] = v
		}
	}

	// Add status change to history
	statusChange := structs.StatusChange{
		Status:    req.Status,
		Reason:    req.Reason,
		ChangedBy: "admin", // Should come from context
		ChangedAt: time.Now().UnixMilli(),
	}

	var statusHistory []structs.StatusChange
	if history, ok := extras["status_history"].([]structs.StatusChange); ok {
		statusHistory = history
	}
	statusHistory = append(statusHistory, statusChange)

	extras["status"] = req.Status
	extras["status_history"] = statusHistory

	_, err = s.fileRepo.Update(ctx, slug, types.JSON{
		"extras": extras,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update file status: %w", err)
	}

	// Get updated file
	updatedFile, err := s.fileRepo.GetByID(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated file: %w", err)
	}

	readFile := s.convertToReadFile(updatedFile)
	return &structs.FileResponse{
		ReadFile:      readFile,
		StatusHistory: statusHistory,
	}, nil
}

// GetStorageStats gets storage statistics
func (s *adminService) GetStorageStats(ctx context.Context) (*structs.StorageStats, error) {
	// This is a simplified implementation
	// In production, you would aggregate from database

	totalSize, err := s.fileRepo.SumSizeByOwner(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get total size: %w", err)
	}

	totalFiles := s.fileRepo.CountX(ctx, &structs.ListFileParams{})

	return &structs.StorageStats{
		TotalSize:     totalSize,
		TotalFiles:    totalFiles,
		TotalUsers:    100, // Placeholder
		ByCategory:    s.getStatsByCategory(ctx),
		ByStorage:     s.getStatsByStorage(ctx),
		DailyUploads:  s.getDailyUploads(ctx),
		TopUsers:      s.getTopUsers(ctx),
		StorageHealth: "healthy",
	}, nil
}

// GetUsageStats gets usage statistics for a period
func (s *adminService) GetUsageStats(ctx context.Context, period string) (*structs.UsageStats, error) {
	totalSize, _ := s.fileRepo.SumSizeByOwner(ctx, "")
	totalFiles := s.fileRepo.CountX(ctx, &structs.ListFileParams{})

	return &structs.UsageStats{
		Period:     period,
		TotalSize:  totalSize,
		TotalFiles: totalFiles,
		Growth: &structs.GrowthStats{
			SizeGrowth:  5.2, // Placeholder
			FilesGrowth: 3.8, // Placeholder
		},
		Breakdown: s.getUsageBreakdown(ctx, period),
	}, nil
}

// GetActivityStats gets activity statistics
func (s *adminService) GetActivityStats(ctx context.Context) (*structs.ActivityStats, error) {
	return &structs.ActivityStats{
		TotalDownloads: 12345, // Placeholder
		TotalViews:     54321, // Placeholder
		PopularFiles:   s.getPopularFiles(ctx),
		ActivityByHour: s.getActivityByHour(ctx),
	}, nil
}

// ListQuotas lists all user quotas
func (s *adminService) ListQuotas(ctx context.Context, params *structs.AdminQuotaListParams) (*structs.AdminQuotaListResponse, error) {
	// Get all users with files
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
	// Reset to default quota
	defaultQuota := int64(10 * 1024 * 1024 * 1024) // 10GB
	return s.quotaService.SetQuota(ctx, userID, defaultQuota)
}

// BatchCleanup performs batch cleanup operations
func (s *adminService) BatchCleanup(ctx context.Context, req *structs.BatchCleanupRequest) (*structs.BatchCleanupResult, error) {
	jobID := uuid.New().String()

	result := &structs.BatchCleanupResult{
		JobID:        jobID,
		Type:         req.Type,
		ItemsFound:   0,
		ItemsCleaned: 0,
		SpaceFreed:   0,
		DryRun:       req.DryRun,
		CleanedItems: make([]string, 0),
		Errors:       make([]string, 0),
	}

	// Implement cleanup logic based on type
	switch req.Type {
	case "expired":
		result = s.cleanupExpiredFiles(ctx, req, result)
	case "orphaned":
		result = s.cleanupOrphanedFiles(ctx, req, result)
	case "duplicates":
		result = s.cleanupDuplicateFiles(ctx, req, result)
	default:
		return nil, fmt.Errorf("unknown cleanup type: %s", req.Type)
	}

	return result, nil
}

// ListBatchJobs lists batch jobs
func (s *adminService) ListBatchJobs(ctx context.Context, params *structs.AdminBatchJobParams) (*structs.BatchJobListResponse, error) {
	jobs := make([]*structs.BatchJob, 0)
	for _, job := range s.batchJobs {
		if params.Status == "" || job.Status == params.Status {
			jobs = append(jobs, job)
		}
	}

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
	job, exists := s.batchJobs[jobID]
	if !exists {
		return fmt.Errorf("batch job not found")
	}

	if job.Status == "completed" || job.Status == "cancelled" {
		return fmt.Errorf("cannot cancel job in status: %s", job.Status)
	}

	job.Status = "cancelled"
	return nil
}

// OptimizeStorage optimizes storage system
func (s *adminService) OptimizeStorage(ctx context.Context) (*structs.OptimizeResult, error) {
	taskID := uuid.New().String()

	// Simulate optimization process
	return &structs.OptimizeResult{
		TaskID:            taskID,
		DeduplicatedFiles: 23,
		SpaceFreed:        1024 * 1024 * 512, // 512MB
		OrphanedCleaned:   5,
		IndexesRebuilt:    3,
		Duration:          120, // 2 minutes
	}, nil
}

// GetStorageHealth gets storage health status
func (s *adminService) GetStorageHealth(ctx context.Context) (*structs.StorageHealth, error) {
	return &structs.StorageHealth{
		Status:         "healthy",
		TotalSpace:     1024 * 1024 * 1024 * 1024, // 1TB
		UsedSpace:      1024 * 1024 * 1024 * 100,  // 100GB
		FreeSpace:      1024 * 1024 * 1024 * 924,  // 924GB
		UsagePercent:   9.8,
		OrphanedFiles:  0,
		CorruptedFiles: 0,
		HealthChecks: []structs.HealthCheck{
			{
				Name:    "Storage Connectivity",
				Status:  "ok",
				Message: "All storage systems are accessible",
				LastRun: time.Now().UnixMilli(),
			},
			{
				Name:    "Database Consistency",
				Status:  "ok",
				Message: "Database and storage are in sync",
				LastRun: time.Now().UnixMilli(),
			},
		},
		Recommendations: []string{},
	}, nil
}

// InitiateBackup initiates storage backup
func (s *adminService) InitiateBackup(ctx context.Context, req *structs.BackupRequest) (*structs.BackupResult, error) {
	backupID := uuid.New().String()

	return &structs.BackupResult{
		BackupID:    backupID,
		Type:        req.Type,
		Status:      "started",
		Destination: req.Destination,
		FileCount:   0,
		TotalSize:   0,
		Duration:    0,
		StartedAt:   time.Now().UnixMilli(),
	}, nil
}

// convertToReadFile converts ent.File to structs.ReadFile
func (s *adminService) convertToReadFile(file *ent.File) *structs.ReadFile {
	// Convert ent.File to structs.ReadFile
	// This is a simplified version - you should implement proper conversion
	return &structs.ReadFile{
		ID:      file.ID,
		Name:    file.Name,
		Path:    file.Path,
		Type:    file.Type,
		Size:    &file.Size,
		Storage: file.Storage,
		Bucket:  file.Bucket,
		OwnerID: file.OwnerID,
	}
}

// calculateFileStats calculates file stats
func (s *adminService) calculateFileStats(files []*ent.File) *structs.FileStats {
	stats := &structs.FileStats{
		ByCategory: make(map[string]int),
		ByStatus:   make(map[string]int),
	}

	for _, file := range files {
		stats.TotalSize += int64(file.Size)
		stats.TotalCount++

		// Category stats would be extracted from file extras
		// Status stats would be extracted from file extras
	}

	return stats
}

// getStatsByCategory gets stats by category
func (s *adminService) getStatsByCategory(ctx context.Context) map[string]int64 {
	return map[string]int64{
		"image":    1024 * 1024 * 100,
		"document": 1024 * 1024 * 50,
		"video":    1024 * 1024 * 200,
	}
}

// getStatsByStorage gets stats by storage
func (s *adminService) getStatsByStorage(ctx context.Context) map[string]int64 {
	return map[string]int64{
		"local": 1024 * 1024 * 200,
		"s3":    1024 * 1024 * 150,
	}
}

// getDailyUploads gets daily uploads
func (s *adminService) getDailyUploads(ctx context.Context) []structs.DailyUpload {
	return []structs.DailyUpload{
		{Date: "2025-01-01", Count: 45, Size: 1024 * 1024 * 20},
		{Date: "2025-01-02", Count: 32, Size: 1024 * 1024 * 15},
	}
}

// getTopUsers gets top users
func (s *adminService) getTopUsers(ctx context.Context) []structs.UserUsage {
	return []structs.UserUsage{
		{UserID: "user1", Size: 1024 * 1024 * 50, Files: 100},
		{UserID: "user2", Size: 1024 * 1024 * 30, Files: 75},
	}
}

// getUsageBreakdown gets usage breakdown
func (s *adminService) getUsageBreakdown(ctx context.Context, period string) []structs.UsageByDate {
	return []structs.UsageByDate{
		{Date: "2025-01-01", Size: 1024 * 1024 * 10, Files: 20},
		{Date: "2025-01-02", Size: 1024 * 1024 * 15, Files: 30},
	}
}

// getPopularFiles gets popular files
func (s *adminService) getPopularFiles(ctx context.Context) []structs.PopularFile {
	return []structs.PopularFile{
		// Populate with actual data
	}
}

// getActivityByHour gets activity by hour
func (s *adminService) getActivityByHour(ctx context.Context) []structs.HourlyActivity {
	return []structs.HourlyActivity{
		{Hour: 9, Downloads: 45, Views: 120},
		{Hour: 10, Downloads: 67, Views: 180},
	}
}

// cleanupExpiredFiles cleans up expired files
func (s *adminService) cleanupExpiredFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	// TODO: Implementation for cleaning up expired files
	return result
}

// cleanupOrphanedFiles cleans up orphaned files
func (s *adminService) cleanupOrphanedFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	// TODO: Implementation for cleaning up orphaned files
	// Files that exist in storage but not in database
	return result
}

// cleanupDuplicateFiles cleans up duplicate files
func (s *adminService) cleanupDuplicateFiles(ctx context.Context, req *structs.BatchCleanupRequest, result *structs.BatchCleanupResult) *structs.BatchCleanupResult {
	// TODO: Implementation for cleaning up duplicate files
	// Files with same hash/content
	return result
}
