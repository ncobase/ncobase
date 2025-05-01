package handler

import (
	"encoding/json"
	"ncobase/domain/resource/service"
	"ncobase/domain/resource/structs"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
)

// BatchHandlerInterface defines the interface for batch handler operations
type BatchHandlerInterface interface {
	BatchUpload(c *gin.Context)
	BatchProcess(c *gin.Context)
}

// batchHandler handles batch operations
type batchHandler struct {
	fileService  service.FileServiceInterface
	batchService service.BatchServiceInterface
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(fileService service.FileServiceInterface, batchService service.BatchServiceInterface) BatchHandlerInterface {
	return &batchHandler{
		fileService:  fileService,
		batchService: batchService,
	}
}

// BatchUpload handles uploading multiple files in a batch
//
// @Summary Batch upload
// @Description Upload multiple files in a batch
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "Files to upload"
// @Param tenant_id formData string true "Tenant ID"
// @Param object_id formData string true "Object ID"
// @Param folder_path formData string false "Virtual folder path"
// @Param access_level formData string false "Access level"
// @Param tags formData string false "Comma-separated tags"
// @Param extras formData string false "Additional metadata (JSON)"
// @Success 200 {object} structs.BatchUploadResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/batch/upload [post]
// @Security Bearer
func (h *batchHandler) BatchUpload(c *gin.Context) {
	// Parse form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse multipart form"))
		return
	}

	// Get files
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No files provided"))
		return
	}

	// Get form parameters
	tenantID := c.PostForm("tenant_id")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	objectID := c.PostForm("object_id")
	if objectID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("object_id")))
		return
	}

	// Create batch upload params
	params := &structs.BatchUploadParams{
		TenantID:   tenantID,
		ObjectID:   objectID,
		FolderPath: c.PostForm("folder_path"),
	}

	// Set access level if provided
	if accessLevel := c.PostForm("access_level"); accessLevel != "" {
		params.AccessLevel = structs.AccessLevel(accessLevel)
	}

	// Set tags if provided
	if tags := c.PostForm("tags"); tags != "" {
		params.Tags = strings.Split(tags, ",")
	}

	// Set extras if provided
	if extrasStr := c.PostForm("extras"); extrasStr != "" {
		var extras types.JSON
		if err := json.Unmarshal([]byte(extrasStr), &extras); err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Invalid extras format"))
			return
		}
		params.Extras = &extras
	}

	// Set processing options
	params.ProcessingOptions = &structs.ProcessingOptions{
		CreateThumbnail: true,
		MaxWidth:        300,
		MaxHeight:       300,
	}

	// Perform batch upload
	result, err := h.batchService.BatchUpload(c.Request.Context(), files, params)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error in batch upload: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to upload files"))
		return
	}

	resp.Success(c.Writer, result)
}

// BatchProcess handles processing multiple files in a batch
//
// @Summary Batch process
// @Description Process multiple files in a batch
// @Tags res
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Processing parameters"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/batch/process [post]
// @Security Bearer
func (h *batchHandler) BatchProcess(c *gin.Context) {
	var body struct {
		IDs      []string                   `json:"ids" binding:"required"`
		TenantID string                     `json:"tenant_id" binding:"required"`
		Options  *structs.ProcessingOptions `json:"options"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body"))
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

		files = append(files, file)
	}

	if len(files) == 0 {
		resp.Fail(c.Writer, resp.NotFound("No files found"))
		return
	}

	// Process files
	processed, err := h.batchService.ProcessImages(c.Request.Context(), files, body.Options)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error processing files: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to process files"))
		return
	}

	resp.Success(c.Writer, processed)
}
