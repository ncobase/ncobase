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
	"strconv"
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

// FileHandlerInterface defines file handler methods
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

type fileHandler struct {
	s *service.Service
}

// NewFileHandler creates new file handler
func NewFileHandler(s *service.Service) FileHandlerInterface {
	return &fileHandler{s: s}
}

var maxFileSize int64 = 2048 << 20 // 2048 MB

// Create handles file creation
//
// @Summary Create files
// @Description Create one or multiple files
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param owner_id formData string false "Owner ID"
// @Param space_id formData string false "Space ID"
// @Param extras formData string false "Additional properties (JSON)"
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
	if strings.HasPrefix(contentType, "multipart/") {
		h.handleFormDataUpload(c)
	} else {
		resp.Fail(c.Writer, resp.BadRequest("Unsupported content type"))
	}
}

// handleFormDataUpload handles multipart form data upload
func (h *fileHandler) handleFormDataUpload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse multipart form"))
		return
	}

	// Handle single file upload
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()

		body, err := h.processFile(c, header, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Error processing file: %v", err)))
			return
		}

		if err := h.validateFileBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Validation error: %v", err)))
			return
		}

		result, err := h.s.File.Create(c.Request.Context(), body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to create file: %v", err)))
			return
		}

		resp.Success(c.Writer, result)
		return
	}

	// Handle multiple files
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No files provided"))
		return
	}

	var results []*structs.ReadFile
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to open file %s: %v", fileHeader.Filename, err)))
			return
		}

		func(file multipart.File) {
			defer file.Close()

			body, err := h.processFile(c, fileHeader, file)
			if err != nil {
				resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Error processing file %s: %v", fileHeader.Filename, err)))
				return
			}

			if err := h.validateFileBody(body); err != nil {
				resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Validation error for file %s: %v", fileHeader.Filename, err)))
				return
			}

			result, err := h.s.File.Create(c.Request.Context(), body)
			if err != nil {
				resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to create file %s: %v", fileHeader.Filename, err)))
				return
			}

			results = append(results, result)
		}(file)
	}

	resp.Success(c.Writer, results)
}

// validateFileBody validates file body
func (h *fileHandler) validateFileBody(body *structs.CreateFileBody) error {
	if validator.IsEmpty(body.OwnerID) {
		return errors.New("owner_id is required")
	}

	if validator.IsEmpty(body.SpaceID) {
		return errors.New("space_id is required")
	}

	// Validate access level
	if body.AccessLevel != "" {
		switch body.AccessLevel {
		case structs.AccessLevelPublic, structs.AccessLevelPrivate, structs.AccessLevelShared:
			// Valid
		default:
			return fmt.Errorf("invalid access_level: %s", body.AccessLevel)
		}
	}

	// Validate processing options
	if body.ProcessingOptions != nil {
		if body.ProcessingOptions.CompressionQuality < 1 || body.ProcessingOptions.CompressionQuality > 100 {
			return errors.New("compression_quality must be between 1 and 100")
		}
		if body.ProcessingOptions.MaxWidth < 0 || body.ProcessingOptions.MaxHeight < 0 {
			return errors.New("max_width and max_height must be non-negative")
		}
	}

	return nil
}

// processFile processes file and binds form fields
func (h *fileHandler) processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateFileBody, error) {
	body := &structs.CreateFileBody{}

	folderPath := c.PostForm("folder_path")
	fileHeader := storage.GetFileHeader(header, folderPath)

	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = &fileHeader.Size

	return h.bindFileFields(c, body)
}

// bindFileFields binds form fields to file body
func (h *fileHandler) bindFileFields(c *gin.Context, body *structs.CreateFileBody) (*structs.CreateFileBody, error) {
	for key, values := range c.Request.Form {
		if len(values) == 0 || (key != "file" && values[0] == "") {
			continue
		}

		switch key {
		case "owner_id":
			body.OwnerID = values[0]
		case "space_id":
			body.SpaceID = values[0]
		case "folder_path":
			body.FolderPath = values[0]
		case "access_level":
			body.AccessLevel = structs.AccessLevel(values[0])
		case "is_public":
			body.IsPublic = values[0] == "true" || values[0] == "1"
		case "tags":
			tagList := strings.Split(values[0], ",")
			cleanTags := make([]string, 0, len(tagList))
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					cleanTags = append(cleanTags, tag)
				}
			}
			body.Tags = cleanTags
		case "processing_options":
			var options structs.ProcessingOptions
			if err := json.Unmarshal([]byte(values[0]), &options); err != nil {
				return nil, fmt.Errorf("invalid processing options format: %w", err)
			}
			if options.CompressionQuality <= 0 || options.CompressionQuality > 100 {
				options.CompressionQuality = 80
			}
			body.ProcessingOptions = &options
		case "expires_at":
			if expiresAtInt, err := strconv.ParseInt(values[0], 10, 64); err == nil {
				body.ExpiresAt = &expiresAtInt
			}
		case "extras":
			var extras types.JSON
			if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
				return nil, fmt.Errorf("invalid extras format: %w", err)
			}
			body.Extras = &extras
		}
	}

	return body, nil
}

// Update handles file updates
//
// @Summary Update file
// @Description Update an existing file
// @Tags res
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "File slug"
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

	updates := make(types.JSON)

	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse form"))
		return
	}

	// Bind form values
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
				updates["is_public"] = values[0] == "true" || values[0] == "1"
			default:
				updates[key] = values[0]
			}
		}
	}

	// Handle file upload
	if fileHeaders, ok := c.Request.MultipartForm.File["file"]; ok && len(fileHeaders) > 0 {
		header := fileHeaders[0]
		folderPath := c.PostForm("folder_path")
		fileHeader := storage.GetFileHeader(header, folderPath)

		updates["name"] = fileHeader.Name
		updates["size"] = fileHeader.Size
		updates["type"] = fileHeader.Type
		updates["path"] = fileHeader.Path

		file, err := header.Open()
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Failed to open uploaded file"))
			return
		}
		defer file.Close()

		updates["file"] = file

		// Handle processing options
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

// Get handles file retrieval
//
// @Summary Get file
// @Description Get details of a specific file
// @Tags res
// @Produce json
// @Param slug path string true "File slug"
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

// Delete handles file deletion
//
// @Summary Delete file
// @Description Delete a specific file
// @Tags res
// @Param slug path string true "File slug"
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

// List handles file listing
//
// @Summary List files
// @Description List files based on specified parameters
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

// download handles file download
func (h *fileHandler) download(c *gin.Context) {
	h.downloadFile(c, "file")
}

// fileStream handles file streaming
func (h *fileHandler) fileStream(c *gin.Context) {
	h.downloadFile(c, "inline")
}

// downloadFile handles file download/streaming
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

	defer func(file io.ReadCloser) {
		if file != nil {
			file.Close()
		}
	}(fileStream)
}

// Search handles file searching
//
// @Summary Search files
// @Description Search files by various criteria
// @Tags res
// @Accept json
// @Produce json
// @Param space_id query string true "Space ID"
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

	if params.SpaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	// Tag-based search
	if params.Tags != "" {
		tags := strings.Split(params.Tags, ",")
		limit := 100
		if params.Limit > 0 {
			limit = params.Limit
		}

		results, err := h.s.File.SearchByTags(c.Request.Context(), params.SpaceID, tags, limit)
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

// ListCategories handles listing file categories
//
// @Summary List file categories
// @Description List all available file categories
// @Tags res
// @Produce json
// @Success 200 {array} string "success"
// @Router /res/categories [get]
// @Security Bearer
func (h *fileHandler) ListCategories(c *gin.Context) {
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

// ListTags handles listing file tags
//
// @Summary List file tags
// @Description List all tags used in files for a space
// @Tags res
// @Produce json
// @Param space_id query string true "Space ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/tags [get]
// @Security Bearer
func (h *fileHandler) ListTags(c *gin.Context) {
	spaceID := c.Query("space_id")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	// Sample tags - in real implementation, query database
	tags := []string{
		"document", "image", "important", "archive", "shared", "draft", "final",
	}

	resp.Success(c.Writer, tags)
}

// GetStorageStats handles storage statistics retrieval
//
// @Summary Get storage statistics
// @Description Get storage usage statistics for a space
// @Tags res
// @Produce json
// @Param space_id query string true "Space ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/stats [get]
// @Security Bearer
func (h *fileHandler) GetStorageStats(c *gin.Context) {
	spaceID := c.Query("space_id")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("space_id")))
		return
	}

	usage, err := h.s.Quota.GetUsage(c.Request.Context(), spaceID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting storage usage: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get storage statistics"))
		return
	}

	quota, err := h.s.Quota.GetQuota(c.Request.Context(), spaceID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error getting storage quota: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to get storage statistics"))
		return
	}

	usagePercent := 0.0
	if quota > 0 {
		usagePercent = float64(usage) / float64(quota) * 100
	}

	stats := map[string]any{
		"total_bytes":     usage,
		"quota_bytes":     quota,
		"usage_percent":   usagePercent,
		"formatted_usage": formatSize(usage),
		"formatted_quota": formatSize(quota),
	}

	resp.Success(c.Writer, stats)
}

// GetVersions handles file version retrieval
//
// @Summary Get file versions
// @Description Get all versions of a file
// @Tags res
// @Produce json
// @Param slug path string true "File slug"
// @Success 200 {array} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
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

// CreateVersion handles file version creation
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
// @Router /res/{slug}/versions [post]
// @Security Bearer
func (h *fileHandler) CreateVersion(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	defer file.Close()

	version, err := h.s.File.CreateVersion(c.Request.Context(), slug, file, header.Filename)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error creating version: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create version"))
		return
	}

	resp.Success(c.Writer, version)
}

// CreateThumbnail handles thumbnail creation
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

	options.CreateThumbnail = true

	file, err := h.s.File.CreateThumbnail(c.Request.Context(), slug, &options)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error creating thumbnail: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create thumbnail"))
		return
	}

	resp.Success(c.Writer, file)
}

// SetAccessLevel handles access level setting
//
// @Summary Set access level
// @Description Set the access level for a file
// @Tags res
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body map[string]string true "Access level"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
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

	accessLevel := structs.AccessLevel(body.AccessLevel)
	if accessLevel != structs.AccessLevelPublic &&
		accessLevel != structs.AccessLevelPrivate &&
		accessLevel != structs.AccessLevelShared {
		resp.Fail(c.Writer, resp.BadRequest("Invalid access level"))
		return
	}

	file, err := h.s.File.SetAccessLevel(c.Request.Context(), slug, accessLevel)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error setting access level: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to set access level"))
		return
	}

	resp.Success(c.Writer, file)
}

// GenerateShareURL handles share URL generation
//
// @Summary Generate share URL
// @Description Generate a shareable URL for a file
// @Tags res
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body map[string]int true "Expiration time in hours"
// @Success 200 {object} map[string]string "success"
// @Failure 400 {object} resp.Exception "bad request"
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
		body.ExpirationHours = 24 // Default to 24 hours
	}

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
