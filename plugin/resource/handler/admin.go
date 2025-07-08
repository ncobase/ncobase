package handler

import (
	"ncobase/resource/service"
	"ncobase/resource/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// AdminHandlerInterface defines admin handler methods
type AdminHandlerInterface interface {
	ListFiles(c *gin.Context)
	DeleteFile(c *gin.Context)
	SetFileStatus(c *gin.Context)

	GetStorageStats(c *gin.Context)
	GetUsageStats(c *gin.Context)
	GetActivityStats(c *gin.Context)

	ListQuotas(c *gin.Context)
	SetQuota(c *gin.Context)
	GetQuota(c *gin.Context)
	DeleteQuota(c *gin.Context)

	BatchCleanup(c *gin.Context)
	ListBatchJobs(c *gin.Context)
	CancelBatchJob(c *gin.Context)

	OptimizeStorage(c *gin.Context)
	GetStorageHealth(c *gin.Context)
	InitiateBackup(c *gin.Context)
}

// adminHandler implements AdminHandlerInterface
type adminHandler struct {
	adminService service.AdminServiceInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService service.AdminServiceInterface) AdminHandlerInterface {
	return &adminHandler{
		adminService: adminService,
	}
}

// ListFiles lists all files for admin
//
// @Summary Admin list files
// @Description List all files with admin view
// @Tags Resource Admin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param status query string false "Status filter"
// @Param user_id query string false "User filter"
// @Success 200 {object} structs.AdminFileListResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/files [get]
// @Security Bearer
func (h *adminHandler) ListFiles(c *gin.Context) {
	params := &structs.AdminFileListParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.adminService.ListFiles(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteFile deletes a file (admin)
//
// @Summary Admin delete file
// @Description Delete file with admin privileges
// @Tags Resource Admin
// @Produce json
// @Param slug path string true "File slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/files/{slug} [delete]
// @Security Bearer
func (h *adminHandler) DeleteFile(c *gin.Context) {
	slug := c.Param("slug")

	err := h.adminService.DeleteFile(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// SetFileStatus sets file status (admin)
//
// @Summary Admin set file status
// @Description Set file status with admin privileges
// @Tags Resource Admin
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body structs.AdminSetStatusRequest true "Status update"
// @Success 200 {object} structs.FileResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/files/{slug}/status [put]
// @Security Bearer
func (h *adminHandler) SetFileStatus(c *gin.Context) {
	slug := c.Param("slug")

	var body structs.AdminSetStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.adminService.SetFileStatus(c.Request.Context(), slug, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetStorageStats gets storage statistics
//
// @Summary Get storage stats
// @Description Get storage statistics for admin dashboard
// @Tags Resource Admin
// @Produce json
// @Success 200 {object} structs.StorageStats "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/stats [get]
// @Security Bearer
func (h *adminHandler) GetStorageStats(c *gin.Context) {
	result, err := h.adminService.GetStorageStats(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetUsageStats gets usage statistics
//
// @Summary Get usage stats
// @Description Get detailed usage statistics
// @Tags Resource Admin
// @Produce json
// @Param period query string false "Time period (day/week/month)"
// @Success 200 {object} structs.UsageStats "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/stats/usage [get]
// @Security Bearer
func (h *adminHandler) GetUsageStats(c *gin.Context) {
	period := c.DefaultQuery("period", "week")

	result, err := h.adminService.GetUsageStats(c.Request.Context(), period)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetActivityStats gets activity statistics
//
// @Summary Get activity stats
// @Description Get file activity statistics
// @Tags Resource Admin
// @Produce json
// @Success 200 {object} structs.ActivityStats "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/stats/activity [get]
// @Security Bearer
func (h *adminHandler) GetActivityStats(c *gin.Context) {
	result, err := h.adminService.GetActivityStats(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListQuotas lists all user quotas
//
// @Summary Admin list quotas
// @Description List all user quotas
// @Tags Resource Admin
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} structs.AdminQuotaListResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/quotas [get]
// @Security Bearer
func (h *adminHandler) ListQuotas(c *gin.Context) {
	params := &structs.AdminQuotaListParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.adminService.ListQuotas(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// SetQuota sets quota for a user
//
// @Summary Admin set user quota
// @Description Set storage quota for a specific user
// @Tags Resource Admin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param quota body structs.QuotaSetRequest true "Quota settings"
// @Success 200 {object} structs.QuotaInfo "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/quotas/{user_id} [post]
// @Security Bearer
func (h *adminHandler) SetQuota(c *gin.Context) {
	userID := c.Param("user_id")

	var body structs.QuotaSetRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.adminService.SetQuota(c.Request.Context(), userID, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetQuota gets quota for a user
//
// @Summary Admin get user quota
// @Description Get storage quota for a specific user
// @Tags Resource Admin
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.QuotaInfo "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/quotas/{user_id} [get]
// @Security Bearer
func (h *adminHandler) GetQuota(c *gin.Context) {
	userID := c.Param("user_id")

	result, err := h.adminService.GetQuota(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteQuota deletes quota for a user
//
// @Summary Admin delete user quota
// @Description Delete storage quota for a specific user
// @Tags Resource Admin
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/quotas/{user_id} [delete]
// @Security Bearer
func (h *adminHandler) DeleteQuota(c *gin.Context) {
	userID := c.Param("user_id")

	err := h.adminService.DeleteQuota(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// BatchCleanup performs batch cleanup
//
// @Summary Admin batch cleanup
// @Description Perform batch cleanup operations
// @Tags Resource Admin
// @Accept json
// @Produce json
// @Param body body structs.BatchCleanupRequest true "Cleanup parameters"
// @Success 200 {object} structs.BatchCleanupResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/batch/cleanup [post]
// @Security Bearer
func (h *adminHandler) BatchCleanup(c *gin.Context) {
	var body structs.BatchCleanupRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.adminService.BatchCleanup(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListBatchJobs lists batch jobs
//
// @Summary Admin list batch jobs
// @Description List all batch jobs
// @Tags Resource Admin
// @Produce json
// @Param status query string false "Job status filter"
// @Param limit query int false "Limit"
// @Success 200 {object} structs.BatchJobListResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/batch/jobs [get]
// @Security Bearer
func (h *adminHandler) ListBatchJobs(c *gin.Context) {
	params := &structs.AdminBatchJobParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.adminService.ListBatchJobs(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// CancelBatchJob cancels a batch job
//
// @Summary Admin cancel batch job
// @Description Cancel a specific batch job
// @Tags Resource Admin
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/batch/jobs/{job_id}/cancel [post]
// @Security Bearer
func (h *adminHandler) CancelBatchJob(c *gin.Context) {
	jobID := c.Param("job_id")

	err := h.adminService.CancelBatchJob(c.Request.Context(), jobID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// OptimizeStorage optimizes storage
//
// @Summary Admin optimize storage
// @Description Optimize storage system
// @Tags Resource Admin
// @Produce json
// @Success 200 {object} structs.OptimizeResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/storage/optimize [post]
// @Security Bearer
func (h *adminHandler) OptimizeStorage(c *gin.Context) {
	result, err := h.adminService.OptimizeStorage(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetStorageHealth gets storage health
//
// @Summary Admin get storage health
// @Description Get storage system health status
// @Tags Resource Admin
// @Produce json
// @Success 200 {object} structs.StorageHealth "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/storage/health [get]
// @Security Bearer
func (h *adminHandler) GetStorageHealth(c *gin.Context) {
	result, err := h.adminService.GetStorageHealth(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// InitiateBackup initiates storage backup
//
// @Summary Admin initiate backup
// @Description Initiate storage backup process
// @Tags Resource Admin
// @Accept json
// @Produce json
// @Param body body structs.BackupRequest true "Backup parameters"
// @Success 200 {object} structs.BackupResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/admin/storage/backup [post]
// @Security Bearer
func (h *adminHandler) InitiateBackup(c *gin.Context) {
	var body structs.BackupRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.adminService.InitiateBackup(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
