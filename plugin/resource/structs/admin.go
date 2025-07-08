package structs

import "github.com/ncobase/ncore/types"

// AdminFileListParams for admin file listing
type AdminFileListParams struct {
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
	Offset int    `form:"offset,omitempty" json:"offset,omitempty"`
	Status string `form:"status,omitempty" json:"status,omitempty"`
	UserID string `form:"user_id,omitempty" json:"user_id,omitempty"`
}

// AdminFileListResponse for admin file list response
type AdminFileListResponse struct {
	Files []*ReadFile `json:"files"`
	Total int         `json:"total"`
	Stats *FileStats  `json:"stats,omitempty"`
}

// FileStats for file statistics
type FileStats struct {
	TotalSize  int64          `json:"total_size"`
	TotalCount int            `json:"total_count"`
	ByCategory map[string]int `json:"by_category"`
	ByStatus   map[string]int `json:"by_status"`
}

// AdminSetStatusRequest for setting file status
type AdminSetStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// FileResponse for file response
type FileResponse struct {
	*ReadFile
	StatusHistory []StatusChange `json:"status_history,omitempty"`
}

// StatusChange for status change history
type StatusChange struct {
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
	ChangedBy string `json:"changed_by"`
	ChangedAt int64  `json:"changed_at"`
}

// StorageStats for storage statistics
type StorageStats struct {
	TotalSize     int64            `json:"total_size"`
	TotalFiles    int              `json:"total_files"`
	TotalUsers    int              `json:"total_users"`
	ByCategory    map[string]int64 `json:"by_category"`
	ByStorage     map[string]int64 `json:"by_storage"`
	DailyUploads  []DailyUpload    `json:"daily_uploads"`
	TopUsers      []UserUsage      `json:"top_users"`
	StorageHealth string           `json:"storage_health"`
}

// DailyUpload for daily upload statistics
type DailyUpload struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Size  int64  `json:"size"`
}

// UserUsage for user usage statistics
type UserUsage struct {
	UserID string `json:"user_id"`
	Size   int64  `json:"size"`
	Files  int    `json:"files"`
}

// UsageStats for usage statistics
type UsageStats struct {
	Period     string        `json:"period"`
	TotalSize  int64         `json:"total_size"`
	TotalFiles int           `json:"total_files"`
	Growth     *GrowthStats  `json:"growth,omitempty"`
	Breakdown  []UsageByDate `json:"breakdown"`
}

// GrowthStats for growth statistics
type GrowthStats struct {
	SizeGrowth  float64 `json:"size_growth"`  // percentage
	FilesGrowth float64 `json:"files_growth"` // percentage
}

// UsageByDate for usage breakdown by date
type UsageByDate struct {
	Date  string `json:"date"`
	Size  int64  `json:"size"`
	Files int    `json:"files"`
}

// ActivityStats for activity statistics
type ActivityStats struct {
	TotalDownloads int64            `json:"total_downloads"`
	TotalViews     int64            `json:"total_views"`
	PopularFiles   []PopularFile    `json:"popular_files"`
	ActivityByHour []HourlyActivity `json:"activity_by_hour"`
}

// PopularFile for popular file statistics
type PopularFile struct {
	*ReadFile
	Downloads int `json:"downloads"`
	Views     int `json:"views"`
}

// HourlyActivity for hourly activity statistics
type HourlyActivity struct {
	Hour      int `json:"hour"`
	Downloads int `json:"downloads"`
	Views     int `json:"views"`
}

// AdminQuotaListParams for admin quota listing
type AdminQuotaListParams struct {
	Limit  int `form:"limit,omitempty" json:"limit,omitempty"`
	Offset int `form:"offset,omitempty" json:"offset,omitempty"`
}

// AdminQuotaListResponse for admin quota list response
type AdminQuotaListResponse struct {
	Quotas []*QuotaInfo `json:"quotas"`
	Total  int          `json:"total"`
}

// QuotaInfo for quota information
type QuotaInfo struct {
	UserID       string  `json:"user_id"`
	Quota        int64   `json:"quota"`
	Usage        int64   `json:"usage"`
	UsagePercent float64 `json:"usage_percent"`
	FileCount    int     `json:"file_count"`
	LastActivity *int64  `json:"last_activity,omitempty"`
}

// QuotaSetRequest for setting quota
type QuotaSetRequest struct {
	Quota int64 `json:"quota" binding:"required"`
}

// BatchCleanupRequest for batch cleanup
type BatchCleanupRequest struct {
	Type     string          `json:"type" binding:"required"` // expired, orphaned, duplicates
	DryRun   bool            `json:"dry_run,omitempty"`
	Filters  *CleanupFilters `json:"filters,omitempty"`
	MaxItems int             `json:"max_items,omitempty"`
}

// CleanupFilters for cleanup filters
type CleanupFilters struct {
	OlderThan  *int64   `json:"older_than,omitempty"` // timestamp
	Categories []string `json:"categories,omitempty"`
	MinSize    *int64   `json:"min_size,omitempty"`
	MaxSize    *int64   `json:"max_size,omitempty"`
}

// BatchCleanupResult for batch cleanup result
type BatchCleanupResult struct {
	JobID        string   `json:"job_id"`
	Type         string   `json:"type"`
	ItemsFound   int      `json:"items_found"`
	ItemsCleaned int      `json:"items_cleaned"`
	SpaceFreed   int64    `json:"space_freed"`
	DryRun       bool     `json:"dry_run"`
	CleanedItems []string `json:"cleaned_items,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// AdminBatchJobParams for admin batch job listing
type AdminBatchJobParams struct {
	Status string `form:"status,omitempty" json:"status,omitempty"`
	Limit  int    `form:"limit,omitempty" json:"limit,omitempty"`
}

// BatchJobListResponse for batch job list response
type BatchJobListResponse struct {
	Jobs  []*BatchJob `json:"jobs"`
	Total int         `json:"total"`
}

// BatchJob for batch job information
type BatchJob struct {
	ID             string      `json:"id"`
	Type           string      `json:"type"`
	Status         string      `json:"status"`
	Progress       int         `json:"progress"`
	ItemCount      int         `json:"item_count"`
	ProcessedCount int         `json:"processed_count"`
	ErrorCount     int         `json:"error_count"`
	StartedAt      int64       `json:"started_at"`
	CompletedAt    *int64      `json:"completed_at,omitempty"`
	CreatedBy      string      `json:"created_by"`
	Result         *types.JSON `json:"result,omitempty"`
	Errors         []string    `json:"errors,omitempty"`
}

// OptimizeResult for storage optimization result
type OptimizeResult struct {
	TaskID            string `json:"task_id"`
	DeduplicatedFiles int    `json:"deduplicated_files"`
	SpaceFreed        int64  `json:"space_freed"`
	OrphanedCleaned   int    `json:"orphaned_cleaned"`
	IndexesRebuilt    int    `json:"indexes_rebuilt"`
	Duration          int64  `json:"duration"` // seconds
}

// StorageHealth for storage health status
type StorageHealth struct {
	Status          string        `json:"status"` // healthy, warning, critical
	TotalSpace      int64         `json:"total_space"`
	UsedSpace       int64         `json:"used_space"`
	FreeSpace       int64         `json:"free_space"`
	UsagePercent    float64       `json:"usage_percent"`
	OrphanedFiles   int           `json:"orphaned_files"`
	CorruptedFiles  int           `json:"corrupted_files"`
	HealthChecks    []HealthCheck `json:"health_checks"`
	Recommendations []string      `json:"recommendations,omitempty"`
}

// HealthCheck for individual health check
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok, warning, error
	Message string `json:"message,omitempty"`
	LastRun int64  `json:"last_run"`
}

// BackupRequest for backup request
type BackupRequest struct {
	Type        string   `json:"type" binding:"required"` // full, incremental
	Destination string   `json:"destination" binding:"required"`
	Compression bool     `json:"compression,omitempty"`
	Encryption  bool     `json:"encryption,omitempty"`
	Filters     []string `json:"filters,omitempty"` // file patterns to include/exclude
}

// BackupResult for backup result
type BackupResult struct {
	BackupID    string `json:"backup_id"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Destination string `json:"destination"`
	FileCount   int    `json:"file_count"`
	TotalSize   int64  `json:"total_size"`
	Duration    int64  `json:"duration"` // seconds
	StartedAt   int64  `json:"started_at"`
	CompletedAt *int64 `json:"completed_at,omitempty"`
}
