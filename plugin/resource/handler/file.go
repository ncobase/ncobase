package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"ncobase/resource/service"
	"ncobase/resource/structs"
	"net/http"
	"strings"
	"time"

	"github.com/ncobase/ncore/validation"

	"github.com/ncobase/ncore/data/storage"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation/validator"

	"github.com/gin-gonic/gin"
)

// FileHandlerInterface represents the file handler interface.
type FileHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
	Search(c *gin.Context)
	ListCategories(c *gin.Context)
	ListTags(c *gin.Context)
	GetStorageStats(c *gin.Context)
	GetVersions(c *gin.Context)
	CreateVersion(c *gin.Context)
	CreateThumbnail(c *gin.Context)
	SetAccessLevel(c *gin.Context)
	GenerateShareURL(c *gin.Context)
}

// fileHandler represents the file handler.
type fileHandler struct {
	s *service.Service
}

// NewFileHandler creates a new file handler.
func NewFileHandler(s *service.Service) FileHandlerInterface {
	return &fileHandler{
		s: s,
	}
}

// maxFileSize is the maximum allowed size of an file.
var maxFileSize int64 = 2048 << 20 // 2048 MB

// Create handles the creation of files, both single and multiple.
//
// @Summary Create files
// @Description Create one or multiple files.
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param object_id formData string false "Object ID associated with the file"
// @Param tenant_id formData string false "Tenant ID associated with the file"
// @Param extras formData string false "Additional properties associated with the file (JSON format)"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res [post]
// @Security Bearer
func (h *fileHandler) Create(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		resp.Fail(c.Writer, resp.NotAllowed("Method not allowed"))
		return
	}
	contentType := c.ContentType()
	switch {
	case strings.HasPrefix(contentType, "multipart/"):
		h.handleFormDataUpload(c)
	default:
		resp.Fail(c.Writer, resp.BadRequest("Unsupported content type"))
		return
	}
}

// handleFormDataUpload handles file upload using multipart form data
func (h *fileHandler) handleFormDataUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				logger.Errorf(c, "Error closing file: %v", err)
			}
		}(file)
		body, err := processFile(c, header, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateFileBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.File.Create(c.Request.Context(), body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		resp.Success(c.Writer, result)
		return
	}

	err = c.Request.ParseMultipartForm(maxFileSize) // Set maxMemory to 32MB
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to parse multipart form"))
		return
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	var results []*structs.ReadFile
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to open file"))
			return
		}
		//goland:noinspection ALL
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				logger.Errorf(c, "Error closing file: %v", err)
			}
		}(file)

		body, err := processFile(c, fileHeader, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateFileBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.File.Create(c.Request.Context(), body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result)
	}
	resp.Success(c.Writer, results)
}

func (h *fileHandler) validateFileBody(body *structs.CreateFileBody) error {
	if validator.IsEmpty(body.ObjectID) {
		return errors.New("belongsTo object is required")
	}
	return nil
}

// processFile processes file details and binds other fields from the form to the file body
func processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateFileBody, error) {
	body := &structs.CreateFileBody{}
	fileHeader := storage.GetFileHeader(header)
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = &fileHeader.Size

	// Bind other fields from the form
	if err := bindFileFields(c, body); err != nil {
		return nil, err
	}
	return body, nil
}

// bindFileFields binds other fields from the form to the file body
func bindFileFields(c *gin.Context, body *structs.CreateFileBody) error {
	// Manually bind other fields from the form
	for key, values := range c.Request.Form {
		if len(values) == 0 || (key != "file" && values[0] == "") {
			continue
		}
		switch key {
		case "object_id":
			body.ObjectID = values[0]
		case "tenant_id":
			body.TenantID = values[0]
		case "folder_path":
			body.FolderPath = values[0]
		case "extras":
			var extras types.JSON
			if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
				return errors.New("invalid extras format")
			}
			body.Extras = &extras
		case "access_level":
			body.AccessLevel = structs.AccessLevel(values[0])
		case "tags":
			body.Tags = strings.Split(values[0], ",")
		case "is_public":
			isPublic := values[0] == "true" || values[0] == "1"
			body.IsPublic = isPublic
		case "processing_options":
			var options structs.ProcessingOptions
			if err := json.Unmarshal([]byte(values[0]), &options); err != nil {
				return errors.New("invalid processing options format")
			}
			body.ProcessingOptions = &options
		}
	}
	return nil
}

// Update handles updating a file.
//
// @Summary Update file
// @Description Update an existing file.
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "Slug of the file to update"
// @Param file body structs.UpdateFileBody true "File details"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug} [put]
// @Security Bearer
func (h *fileHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Create a map to hold the updates
	updates := make(types.JSON)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse form"))
		return
	}

	// Bind form values to updates
	for key, values := range c.Request.Form {
		if len(values) > 0 && values[0] != "" {
			switch key {
			case "folder_path":
				updates["folder_path"] = values[0]
			case "access_level":
				updates["access_level"] = structs.AccessLevel(values[0])
			case "extras":
				var extras types.JSON
				if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
					resp.Fail(c.Writer, resp.BadRequest("Invalid extras format"))
					return
				}
				updates["extras"] = extras
			case "tags":
				updates["tags"] = strings.Split(values[0], ",")
			case "is_public":
				isPublic := values[0] == "true" || values[0] == "1"
				updates["is_public"] = isPublic
			default:
				updates[key] = values[0]
			}
		}
	}

	// Check if the file is included in the request
	if fileHeaders, ok := c.Request.MultipartForm.File["file"]; ok && len(fileHeaders) > 0 {
		// Fetch file header from request
		header := fileHeaders[0]
		// Get file data
		fileHeader := storage.GetFileHeader(header)
		// Add file header data to updates
		updates["name"] = fileHeader.Name
		updates["size"] = fileHeader.Size
		updates["type"] = fileHeader.Type
		updates["path"] = fileHeader.Path

		// Open the file
		file, err := header.Open()
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Failed to open uploaded file"))
			return
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			}
		}(file)
		updates["file"] = file

		// Handle processing options if provided
		if processingOptionsStr := c.PostForm("processing_options"); processingOptionsStr != "" {
			var options structs.ProcessingOptions
			if err := json.Unmarshal([]byte(processingOptionsStr), &options); err != nil {
				resp.Fail(c.Writer, resp.BadRequest("Invalid processing options format"))
				return
			}
			updates["processing_options"] = &options
		}
	}

	result, err := h.s.File.Update(c.Request.Context(), slug, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a file.
//
// @Summary Get file
// @Description Get details of a specific file.
// @Tags res
// @Produce json
// @Param slug path string true "Slug of the file to retrieve"
// @Param type query string false "Type of retrieval ('download' or 'stream')"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug} [get]
func (h *fileHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	if c.Query("type") == "download" {
		h.download(c)
		return
	}

	if c.Query("type") == "stream" {
		h.fileStream(c)
		return
	}

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a file.
//
// @Summary Delete file
// @Description Delete a specific file.
// @Tags res
// @Param slug path string true "Slug of the file to delete"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug} [delete]
// @Security Bearer
func (h *fileHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	if err := h.s.File.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing files.
//
// @Summary List files
// @Description List files based on specified parameters.
// @Tags res
// @Produce json
// @Param params query structs.ListFileParams true "List files parameters"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res [get]
func (h *fileHandler) List(c *gin.Context) {
	params := &structs.ListFileParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	files, err := h.s.File.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, files)
}

// downloadFileHandler handles the direct download of a file.
func (h *fileHandler) download(c *gin.Context) {
	h.downloadFile(c, "file")
}

// fileStreamHandler handles the streaming of a file.
func (h *fileHandler) fileStream(c *gin.Context) {
	h.downloadFile(c, "inline")
}

// downloadFile handles the download or streaming of a file
func (h *fileHandler) downloadFile(c *gin.Context, dispositionType string) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	fileStream, row, err := h.s.File.GetFileStream(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	filename := storage.RestoreOriginalFileName(row.Path, true)
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%s", dispositionType, filename))

	// Set the Content-Type header based on the original content type
	if row.Type == "" {
		c.Header("Content-Type", "application/octet-stream")
	} else {
		c.Header("Content-Type", row.Type)
	}

	_, err = io.Copy(c.Writer, fileStream)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	// close file stream
	defer func(file io.ReadCloser) {
		if file != nil {
			err := file.Close()
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			}
		}
	}(fileStream)
}

// Search handles searching for files
//
// @Summary Search files
// @Description Search files by various criteria
// @Tags res
// @Accept json
// @Produce json
// @Param tenant query string true "Tenant ID"
// @Param q query string false "Search query"
// @Param category query string false "File category"
// @Param tags query string false "Comma-separated tags"
// @Param folder_path query string false "Folder path"
// @Param is_public query boolean false "Public flag"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/search [get]
// @Security Bearer
func (h *fileHandler) Search(c *gin.Context) {
	var params structs.ListFileParams
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if params.Tenant == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant")))
		return
	}

	// For tag-based search
	if params.Tags != "" {
		tags := strings.Split(params.Tags, ",")
		limit := 100 // Default limit
		if params.Limit > 0 {
			limit = params.Limit
		}

		results, err := h.s.File.SearchByTags(c.Request.Context(), params.Tenant, tags, limit)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Error searching files by tags: %v", err)
			resp.Fail(c.Writer, resp.InternalServer("Failed to search files"))
			return
		}

		resp.Success(c.Writer, results)
		return
	}

	// Standard list with filtering
	results, err := h.s.File.List(c.Request.Context(), &params)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error searching files: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to search files"))
		return
	}

	resp.Success(c.Writer, results)
}

// ListCategories handles listing all available file categories
//
// @Summary List file categories
// @Description List all available file categories
// @Tags res
// @Produce json
// @Success 200 {array} string "success"
// @Router /res/categories [get]
// @Security Bearer
func (h *fileHandler) ListCategories(c *gin.Context) {
	// Return all available categories
	categories := []string{
		string(structs.FileCategoryImage),
		string(structs.FileCategoryDocument),
		string(structs.FileCategoryVideo),
		string(structs.FileCategoryAudio),
		string(structs.FileCategoryArchive),
		string(structs.FileCategoryOther),
	}

	resp.Success(c.Writer, categories)
}

// ListTags handles listing all tags used in files
//
// @Summary List file tags
// @Description List all tags used in files for a tenant
// @Tags res
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/tags [get]
// @Security Bearer
func (h *fileHandler) ListTags(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	// For simplicity, we're just returning a sample list
	// In a real implementation, you'd query the database for all unique tags
	tags := []string{
		"document", "image", "important", "archive", "shared", "draft", "final",
	}

	resp.Success(c.Writer, tags)
}

// GetStorageStats handles retrieving storage statistics
//
// @Summary Get storage statistics
// @Description Get storage usage statistics for a tenant
// @Tags res
// @Produce json
// @Param tenant_id query string true "Tenant ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/stats [get]
// @Security Bearer
func (h *fileHandler) GetStorageStats(c *gin.Context) {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("tenant_id")))
		return
	}

	// Get usage from quota service
	usage, err := h.s.Quota.GetUsage(c.Request.Context(), tenantID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting storage usage: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get storage statistics"))
		return
	}

	// Get quota
	quota, err := h.s.Quota.GetQuota(c.Request.Context(), tenantID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting storage quota: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get storage statistics"))
		return
	}

	// Calculate usage percentage
	usagePercent := 0.0
	if quota > 0 {
		usagePercent = float64(usage) / float64(quota) * 100
	}

	// Build response with storage statistics
	stats := map[string]any{
		"total_bytes":     usage,
		"quota_bytes":     quota,
		"usage_percent":   usagePercent,
		"formatted_usage": formatSize(usage),
		"formatted_quota": formatSize(quota),
	}

	resp.Success(c.Writer, stats)
}

// GetVersions handles retrieving all versions of an file
//
// @Summary Get file versions
// @Description Get all versions of an file
// @Tags res
// @Produce json
// @Param slug path string true "File slug"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "not found"
// @Router /res/{slug}/versions [get]
// @Security Bearer
func (h *fileHandler) GetVersions(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	versions, err := h.s.File.GetVersions(c.Request.Context(), slug)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error retrieving versions: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve versions"))
		return
	}

	resp.Success(c.Writer, versions)
}

// CreateVersion handles creating a new version of an file
//
// @Summary Create file version
// @Description Create a new version of an existing file
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "File slug"
// @Param file formData file true "New version file"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "not found"
// @Router /res/{slug}/versions [post]
// @Security Bearer
func (h *fileHandler) CreateVersion(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	defer file.Close()

	// Create new version
	version, err := h.s.File.CreateVersion(c.Request.Context(), slug, file, header.Filename)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error creating version: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create version"))
		return
	}

	resp.Success(c.Writer, version)
}

// CreateThumbnail handles creating a thumbnail for an image file
//
// @Summary Create thumbnail
// @Description Create a thumbnail for an image file
// @Tags res
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param options body structs.ProcessingOptions true "Processing options"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "not found"
// @Router /res/{slug}/thumbnail [post]
// @Security Bearer
func (h *fileHandler) CreateThumbnail(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	var options structs.ProcessingOptions
	if err := c.ShouldBindJSON(&options); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid processing options"))
		return
	}

	// Ensure thumbnail creation is enabled
	options.CreateThumbnail = true

	// Create thumbnail
	file, err := h.s.File.CreateThumbnail(c.Request.Context(), slug, &options)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error creating thumbnail: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create thumbnail"))
		return
	}

	resp.Success(c.Writer, file)
}

// SetAccessLevel handles setting the access level for an file
//
// @Summary Set access level
// @Description Set the access level for an file
// @Tags res
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body map[string]string true "Access level"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "not found"
// @Router /res/{slug}/access [put]
// @Security Bearer
func (h *fileHandler) SetAccessLevel(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	var body struct {
		AccessLevel string `json:"access_level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body"))
		return
	}

	// Validate access level
	accessLevel := structs.AccessLevel(body.AccessLevel)
	if accessLevel != structs.AccessLevelPublic &&
		accessLevel != structs.AccessLevelPrivate &&
		accessLevel != structs.AccessLevelShared {
		resp.Fail(c.Writer, resp.BadRequest("Invalid access level"))
		return
	}

	// Set access level
	file, err := h.s.File.SetAccessLevel(c.Request.Context(), slug, accessLevel)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error setting access level: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to set access level"))
		return
	}

	resp.Success(c.Writer, file)
}

// GenerateShareURL handles generating a shareable URL for an file
//
// @Summary Generate share URL
// @Description Generate a shareable URL for an file
// @Tags res
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body map[string]int true "Expiration time in hours"
// @Success 200 {object} map[string]string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "not found"
// @Router /res/{slug}/share [post]
// @Security Bearer
func (h *fileHandler) GenerateShareURL(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	var body struct {
		ExpirationHours int `json:"expiration_hours"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		// Default to 24 hours if not specified
		body.ExpirationHours = 24
	}

	// Generate share URL
	shareURL, err := h.s.File.GeneratePublicURL(c.Request.Context(), slug, body.ExpirationHours)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error generating share URL: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to generate share URL"))
		return
	}

	resp.Success(c.Writer, map[string]string{
		"url":        shareURL,
		"expires_in": fmt.Sprintf("%d hours", body.ExpirationHours),
		"expires_at": time.Now().Add(time.Duration(body.ExpirationHours) * time.Hour).Format(time.RFC3339),
	})
}

// formatSize formats a size in bytes to a human-readable format
func formatSize(sizeInBytes int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	size := float64(sizeInBytes)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", size/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", size/KB)
	default:
		return fmt.Sprintf("%d bytes", sizeInBytes)
	}
}
