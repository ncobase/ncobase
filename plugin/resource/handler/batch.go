package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"ncobase/plugin/resource/service"
	"ncobase/plugin/resource/structs"
	"ncobase/plugin/resource/wrapper"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils"
)

// BatchHandlerInterface defines batch handler methods
type BatchHandlerInterface interface {
	BatchUpload(c *gin.Context)
	BatchProcess(c *gin.Context)
	BatchDelete(c *gin.Context)
	GetBatchStatus(c *gin.Context)
}

type batchHandler struct {
	fileService    service.FileServiceInterface
	batchService   service.BatchServiceInterface
	configProvider service.ResourceConfigProvider
	spaceWrapper   *wrapper.SpaceServiceWrapper
}

// NewBatchHandler creates new batch handler
func NewBatchHandler(
	fileService service.FileServiceInterface,
	batchService service.BatchServiceInterface,
	configProvider service.ResourceConfigProvider,
	spaceWrapper *wrapper.SpaceServiceWrapper,
) BatchHandlerInterface {
	return &batchHandler{
		fileService:    fileService,
		batchService:   batchService,
		configProvider: configProvider,
		spaceWrapper:   spaceWrapper,
	}
}

// BatchUpload handles uploading multiple files in a batch
//
// @Summary Batch upload
// @Description Upload multiple files in a batch
// @Tags Resource
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload"
// @Param owner_id formData string true "Owner ID"
// @Param path_prefix formData string false "Path prefix"
// @Param access_level formData string false "Access level"
// @Param is_public formData boolean false "Public access flag"
// @Param tags formData string false "Comma-separated tags"
// @Param processing_options formData string false "Processing options (JSON)"
// @Param expires_at formData integer false "Expiration timestamp"
// @Param extras formData string false "Additional metadata (JSON)"
// @Success 200 {object} structs.BatchUploadResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/batch/upload [post]
// @Security Bearer
func (h *batchHandler) BatchUpload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(h.multipartMemoryLimit(c.Request.Context())); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse multipart form"))
		return
	}

	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No files provided"))
		return
	}

	ownerID := c.PostForm("owner_id")
	if ownerID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("owner_id")))
		return
	}
	if err := h.authorizeOwnerAccess(c.Request.Context(), ownerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	params := &structs.BatchUploadParams{
		OwnerID:    ownerID,
		PathPrefix: c.PostForm("path_prefix"),
	}

	if accessLevel := c.PostForm("access_level"); accessLevel != "" {
		level := structs.AccessLevel(accessLevel)
		switch level {
		case structs.AccessLevelPublic, structs.AccessLevelPrivate, structs.AccessLevelShared:
			params.AccessLevel = level
		default:
			resp.Fail(c.Writer, resp.BadRequest("Invalid access level"))
			return
		}
	}

	if isPublic := c.PostForm("is_public"); isPublic != "" {
		params.IsPublic = isPublic == "true" || isPublic == "1"
	}
	if params.AccessLevel == structs.AccessLevelPublic {
		params.IsPublic = true
	} else if params.IsPublic && params.AccessLevel == "" {
		params.AccessLevel = structs.AccessLevelPublic
	}

	if expiresAt := c.PostForm("expires_at"); expiresAt != "" {
		value, err := strconv.ParseInt(expiresAt, 10, 64)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid expires_at format"))
			return
		}
		params.ExpiresAt = &value
	}

	if tags := c.PostForm("tags"); tags != "" {
		rawTags := strings.Split(tags, ",")
		params.Tags = make([]string, 0, len(rawTags))
		for _, tag := range rawTags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				params.Tags = append(params.Tags, tag)
			}
		}
	}

	if extrasStr := c.PostForm("extras"); extrasStr != "" {
		var extras types.JSON
		if err := json.Unmarshal([]byte(extrasStr), &extras); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid extras format"))
			return
		}
		params.Extras = &extras
	}

	if optionsStr := c.PostForm("processing_options"); optionsStr != "" {
		var options structs.ProcessingOptions
		if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid processing options format"))
			return
		}
		if h.configProvider != nil {
			params.ProcessingOptions = h.configProvider.NormalizeProcessingOptions(c.Request.Context(), &options)
		} else {
			params.ProcessingOptions = &options
		}
	} else {
		if h.configProvider != nil {
			params.ProcessingOptions = h.configProvider.DefaultProcessingOptions(c.Request.Context())
		}
	}

	result, err := h.batchService.BatchUpload(c.Request.Context(), files, params)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error in batch upload: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to upload files"))
		return
	}

	resp.Success(c.Writer, result)
}

func (h *batchHandler) multipartMemoryLimit(ctx context.Context) int64 {
	limit := int64(32 << 20)
	if h.configProvider != nil {
		if cfg := h.configProvider.Get(ctx); cfg != nil && cfg.MaxUploadSize > 0 && cfg.MaxUploadSize < limit {
			limit = cfg.MaxUploadSize
		}
	}
	return limit
}

// BatchProcess handles processing multiple files in a batch
//
// @Summary Batch process
// @Description Process multiple files in a batch
// @Tags Resource
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Processing parameters"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/batch/process [post]
// @Security Bearer
func (h *batchHandler) BatchProcess(c *gin.Context) {
	var body struct {
		IDs     []string                   `json:"ids" binding:"required"`
		OwnerID string                     `json:"owner_id" binding:"required"`
		Options *structs.ProcessingOptions `json:"options"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body"))
		return
	}
	if err := h.authorizeOwnerAccess(c.Request.Context(), body.OwnerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	if len(body.IDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No file IDs provided"))
		return
	}

	// Get files
	files := make([]*structs.ReadFile, 0, len(body.IDs))
	for _, id := range body.IDs {
		file, err := h.fileService.Get(c.Request.Context(), id)
		if err != nil {
			logger.Warnf(c.Request.Context(), "Error retrieving file %s: %v", id, err)
			continue
		}
		if !ctxutil.GetUserIsAdmin(c.Request.Context()) && file.OwnerID != body.OwnerID {
			resp.Fail(c.Writer, resp.Forbidden("owner mismatch for file: "+id))
			return
		}
		if err := h.authorizeOwnerAccess(c.Request.Context(), file.OwnerID); err != nil {
			resp.Fail(c.Writer, resp.Forbidden(err.Error()))
			return
		}
		files = append(files, file)
	}

	if len(files) == 0 {
		resp.Fail(c.Writer, resp.NotFound("No files found"))
		return
	}

	processed, err := h.batchService.ProcessImages(c.Request.Context(), files, body.Options)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error processing files: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to process files"))
		return
	}

	resp.Success(c.Writer, processed)
}

// BatchDelete handles deleting multiple files in a batch
//
// @Summary Batch delete
// @Description Delete multiple files in a batch
// @Tags Resource
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Delete parameters"
// @Success 200 {object} structs.BatchDeleteResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/batch/delete [post]
// @Security Bearer
func (h *batchHandler) BatchDelete(c *gin.Context) {
	var body struct {
		IDs     []string `json:"ids" binding:"required"`
		OwnerID string   `json:"owner_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body"))
		return
	}
	if err := h.authorizeOwnerAccess(c.Request.Context(), body.OwnerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	if len(body.IDs) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No file IDs provided"))
		return
	}

	for _, id := range body.IDs {
		file, err := h.fileService.Get(c.Request.Context(), id)
		if err != nil {
			resp.Fail(c.Writer, resp.NotFound("File not found: "+id))
			return
		}
		if !ctxutil.GetUserIsAdmin(c.Request.Context()) && file.OwnerID != body.OwnerID {
			resp.Fail(c.Writer, resp.Forbidden("owner mismatch for file: "+id))
			return
		}
		if err := h.authorizeOwnerAccess(c.Request.Context(), file.OwnerID); err != nil {
			resp.Fail(c.Writer, resp.Forbidden(err.Error()))
			return
		}
	}

	result, err := h.batchService.BatchDelete(c.Request.Context(), body.IDs, body.OwnerID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error in batch delete: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to delete files"))
		return
	}

	resp.Success(c.Writer, result)
}

// GetBatchStatus handles getting batch operation status
//
// @Summary Get batch status
// @Description Get status of a batch operation
// @Tags Resource
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 200 {object} structs.BatchStatus "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/status/{job_id} [get]
// @Security Bearer
func (h *batchHandler) GetBatchStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("job_id")))
		return
	}

	status, err := h.batchService.GetBatchStatus(c.Request.Context(), jobID)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("Batch job not found"))
		return
	}

	resp.Success(c.Writer, status)
}

func (h *batchHandler) authorizeOwnerAccess(ctx context.Context, ownerID string) error {
	if ownerID == "" {
		return nil
	}

	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return fmt.Errorf("unauthorized")
	}

	if ctxutil.GetUserIsAdmin(ctx) || ownerID == userID {
		return nil
	}

	if !looksLikeSpaceOwner(ctx, ownerID) {
		return fmt.Errorf("owner access denied")
	}

	if userSpaceIDs := ctxutil.GetUserSpaceIDs(ctx); len(userSpaceIDs) > 0 {
		if utils.Contains(userSpaceIDs, ownerID) {
			return nil
		}
	}

	if h.spaceWrapper == nil || !h.spaceWrapper.HasUserSpaceService() {
		return fmt.Errorf("space service not available")
	}

	inSpace, err := h.spaceWrapper.IsUserInSpace(ctx, ownerID, userID)
	if err != nil {
		return err
	}
	if !inSpace {
		return fmt.Errorf("owner access denied")
	}

	return nil
}
