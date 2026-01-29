package router

import (
	"ncobase/internal/middleware"
	"ncobase/plugin/resource/handler"

	"github.com/gin-gonic/gin"
)

// Router represents the router for the resource module
type Router struct {
	h *handler.Handler
}

// New creates a new router
func New(handlers *handler.Handler) *Router {
	return &Router{
		h: handlers,
	}
}

// Register registers the router
func (r *Router) Register(rg *gin.RouterGroup, prefix ...string) {
	if len(prefix) > 0 {
		rg = rg.Group("/" + prefix[0])
	}

	// Public routes (no authentication required)
	rg.GET("/view/:slug", r.h.File.GetPublic)
	rg.GET("/share/:token", r.h.File.GetShared)
	rg.GET("/thumb/:slug", r.h.File.GetThumbnail)
	rg.GET("/dl/:slug", r.h.File.DownloadPublic)

	// Protected routes (authentication required)
	protected := rg.Use(middleware.ValidateContentType(), middleware.RequireAuth())

	// Permissions
	read := protected.Use(middleware.HasPermission("read:resources"))
	manage := protected.Use(middleware.HasPermission("manage:resources"))

	// Basic file operations
	read.GET("", r.h.File.List)
	manage.POST("", r.h.File.Create)
	read.GET("/:slug", r.h.File.Get)
	manage.PUT("/:slug", r.h.File.Update)
	manage.DELETE("/:slug", r.h.File.Delete)

	// File search and discovery
	read.GET("/search", r.h.File.Search)
	read.GET("/categories", r.h.File.ListCategories)
	read.GET("/tags", r.h.File.ListTags)

	// File operations
	read.GET("/:slug/versions", r.h.File.GetVersions)
	manage.POST("/:slug/versions", r.h.File.CreateVersion)
	manage.POST("/:slug/thumbnail", r.h.File.CreateThumbnail)
	manage.PUT("/:slug/access", r.h.File.SetAccessLevel)
	manage.POST("/:slug/share", r.h.File.GenerateShareURL)
	read.GET("/:slug/download", r.h.File.Download)

	// User quota and usage
	read.GET("/quota", r.h.Quota.GetMyQuota)
	read.GET("/usage", r.h.Quota.GetMyUsage)

	// Batch operations
	manage.POST("/batch/upload", r.h.Batch.BatchUpload)
	manage.POST("/batch/process", r.h.Batch.BatchProcess)
	manage.POST("/batch/delete", r.h.Batch.BatchDelete)
	read.GET("/status/:job_id", r.h.Batch.GetBatchStatus)

	// Admin routes (admin access required)
	admin := protected.Use(middleware.RequireAdmin())

	// Admin file management
	admin.GET("/admin/files", r.h.Admin.ListFiles)
	admin.DELETE("/admin/files/:slug", r.h.Admin.DeleteFile)
	admin.PUT("/admin/files/:slug/status", r.h.Admin.SetFileStatus)

	// Admin statistics and monitoring
	admin.GET("/admin/stats", r.h.Admin.GetStorageStats)
	admin.GET("/admin/stats/usage", r.h.Admin.GetUsageStats)
	admin.GET("/admin/stats/activity", r.h.Admin.GetActivityStats)

	// Admin quota management
	admin.GET("/admin/quotas", r.h.Admin.ListQuotas)
	admin.POST("/admin/quotas/:user_id", r.h.Admin.SetQuota)
	admin.GET("/admin/quotas/:user_id", r.h.Admin.GetQuota)
	admin.DELETE("/admin/quotas/:user_id", r.h.Admin.DeleteQuota)

	// Admin batch operations
	admin.POST("/admin/batch/cleanup", r.h.Admin.BatchCleanup)
	admin.GET("/admin/batch/jobs", r.h.Admin.ListBatchJobs)
	admin.POST("/admin/batch/jobs/:job_id/cancel", r.h.Admin.CancelBatchJob)

	// Admin storage management
	admin.POST("/admin/storage/optimize", r.h.Admin.OptimizeStorage)
	admin.GET("/admin/storage/health", r.h.Admin.GetStorageHealth)
	admin.POST("/admin/storage/backup", r.h.Admin.InitiateBackup)
}
