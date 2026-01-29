package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"ncobase/plugin/resource/service"
	"ncobase/plugin/resource/structs"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils"
	"github.com/ncobase/ncore/validation"
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
	GetVersions(c *gin.Context)
	CreateVersion(c *gin.Context)
	CreateThumbnail(c *gin.Context)
	SetAccessLevel(c *gin.Context)
	GenerateShareURL(c *gin.Context)
	Download(c *gin.Context)
	GetPublic(c *gin.Context)
	GetShared(c *gin.Context)
	GetThumbnail(c *gin.Context)
	DownloadPublic(c *gin.Context)
}

type fileHandler struct {
	s *service.Service
}

func NewFileHandler(s *service.Service) FileHandlerInterface {
	return &fileHandler{s: s}
}

var maxFileSize int64 = 2048 << 20 // 2048 MB

// Create handles file creation
//
// @Summary Create files with flexible path structure
// @Description Create one or multiple files with optional path organization
// @Tags Resource
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param owner_id formData string false "Owner ID (optional for storage path)"
// @Param path_prefix formData string false "Custom path prefix (e.g., avatars, documents, public)"
// @Param access_level formData string false "Access level" Enums(public, private, shared)
// @Param is_public formData boolean false "Public access flag"
// @Param tags formData string false "Comma-separated tags"
// @Param processing_options formData string false "Processing options (JSON)"
// @Param expires_at formData integer false "Expiration timestamp"
// @Param extras formData string false "Additional properties (JSON)"
// @Success 200 {object} structs.ReadFile "File created successfully"
// @Success 200 {object} object{files=[]structs.ReadFile,total=int,success=int,failed=int,errors=[]string} "Batch upload result"
// @Failure 400 {object} resp.Exception "Bad request"
// @Failure 413 {object} resp.Exception "File too large"
// @Failure 500 {object} resp.Exception "Internal server error"
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
		logger.Errorf(c.Request.Context(), "Failed to parse multipart form: %v", err)
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse multipart form"))
		return
	}

	// Handle single file upload
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()

		if header == nil || header.Filename == "" {
			resp.Fail(c.Writer, resp.BadRequest("Invalid file or filename"))
			return
		}

		body, err := h.processFileWithPathPrefix(c, header, file)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Error processing file %s: %v", header.Filename, err)
			resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Error processing file: %v", err)))
			return
		}

		if err := h.authorizeOwnerAccess(c.Request.Context(), body.OwnerID); err != nil {
			resp.Fail(c.Writer, resp.Forbidden(err.Error()))
			return
		}

		if err = body.Validate(); err != nil {
			logger.Errorf(c.Request.Context(), "Validation error for file %s: %v", header.Filename, err)
			resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Validation error: %v", err)))
			return
		}

		result, err := h.s.File.Create(c.Request.Context(), body)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Failed to create file %s: %v", header.Filename, err)
			resp.Fail(c.Writer, resp.InternalServer(fmt.Sprintf("Failed to create file: %v", err)))
			return
		}

		resp.Success(c.Writer, result.InternalView())
		return
	}

	// Handle multiple files upload
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("No files provided"))
		return
	}

	var results []*structs.ReadFile
	var errs []string

	for i, fileHeader := range files {
		func(index int, header *multipart.FileHeader) {
			if header == nil || header.Filename == "" {
				errs = append(errs, fmt.Sprintf("File %d: invalid header", index))
				return
			}

			file, err := header.Open()
			if err != nil {
				errs = append(errs, fmt.Sprintf("File %s: failed to open: %v", header.Filename, err))
				return
			}
			defer file.Close()

			body, err := h.processFileWithPathPrefix(c, header, file)
			if err != nil {
				errs = append(errs, fmt.Sprintf("File %s: processing error: %v", header.Filename, err))
				return
			}

			if err := h.authorizeOwnerAccess(c.Request.Context(), body.OwnerID); err != nil {
				errs = append(errs, fmt.Sprintf("File %s: access denied: %v", header.Filename, err))
				return
			}

			if err = body.Validate(); err != nil {
				errs = append(errs, fmt.Sprintf("File %s: validation error: %v", header.Filename, err))
				return
			}

			result, err := h.s.File.Create(c.Request.Context(), body)
			if err != nil {
				errs = append(errs, fmt.Sprintf("File %s: creation error: %v", header.Filename, err))
				return
			}

			results = append(results, result.InternalView())
		}(i, fileHeader)
	}

	response := map[string]any{
		"files":   results,
		"total":   len(files),
		"success": len(results),
		"failed":  len(files) - len(results),
		"errors":  errs,
	}

	if len(errs) > 0 {
		logger.Warnf(c.Request.Context(), "Batch upload completed with %d errors", len(errs))
	}

	resp.Success(c.Writer, response)
}

// processFileWithPathPrefix processes file and creates CreateFileBody
func (h *fileHandler) processFileWithPathPrefix(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateFileBody, error) {
	body := &structs.CreateFileBody{}

	// Validate filename
	if header.Filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	// Extract comprehensive file info
	ext := filepath.Ext(header.Filename)
	nameWithoutExt := strings.TrimSuffix(header.Filename, ext)
	if nameWithoutExt == "" {
		nameWithoutExt = "file"
	}

	// Set basic file properties
	body.Name = nameWithoutExt
	body.OriginalName = header.Filename
	body.Path = header.Filename

	// Get content type with fallback
	body.Type = header.Header.Get("Content-Type")
	if body.Type == "" {
		body.Type = "application/octet-stream"
	}

	fileSize := int(header.Size)
	body.Size = &fileSize
	body.File = file

	return h.bindFileFields(c, body)
}

// bindFileFields binds form fields to file body
func (h *fileHandler) bindFileFields(c *gin.Context, body *structs.CreateFileBody) (*structs.CreateFileBody, error) {
	for key, values := range c.Request.Form {
		if len(values) == 0 {
			continue
		}

		switch key {
		case "owner_id":
			if values[0] != "" {
				body.OwnerID = values[0]
			}
		case "path_prefix":
			if values[0] != "" {
				body.PathPrefix = values[0]
			}
		case "access_level":
			if values[0] != "" {
				accessLevel := structs.AccessLevel(values[0])
				switch accessLevel {
				case structs.AccessLevelPublic, structs.AccessLevelPrivate, structs.AccessLevelShared:
					body.AccessLevel = accessLevel
				default:
					return nil, fmt.Errorf("invalid access level: %s", values[0])
				}
			}
		case "is_public":
			body.IsPublic = values[0] == "true" || values[0] == "1"
		case "tags":
			if values[0] != "" {
				tagList := strings.Split(values[0], ",")
				cleanTags := make([]string, 0, len(tagList))
				for _, tag := range tagList {
					tag = strings.TrimSpace(tag)
					if tag != "" {
						cleanTags = append(cleanTags, tag)
					}
				}
				body.Tags = cleanTags
			}
		case "processing_options":
			if values[0] != "" {
				var options structs.ProcessingOptions
				if err := json.Unmarshal([]byte(values[0]), &options); err != nil {
					return nil, fmt.Errorf("invalid processing options format: %w", err)
				}
				// Validate compression quality
				if options.CompressionQuality <= 0 || options.CompressionQuality > 100 {
					options.CompressionQuality = 80
				}
				body.ProcessingOptions = &options
			}
		case "expires_at":
			if values[0] != "" {
				if expiresAtInt, err := strconv.ParseInt(values[0], 10, 64); err == nil {
					body.ExpiresAt = &expiresAtInt
				} else {
					return nil, fmt.Errorf("invalid expires_at format: %s", values[0])
				}
			}
		case "extras":
			if values[0] != "" {
				var extras types.JSON
				if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
					return nil, fmt.Errorf("invalid extras format: %w", err)
				}
				body.Extras = &extras
			}
		case "category":
			if values[0] != "" {
				// Allow manual category override
				category := structs.FileCategory(values[0])
				// Validate category
				validCategories := []structs.FileCategory{
					structs.FileCategoryImage,
					structs.FileCategoryDocument,
					structs.FileCategoryVideo,
					structs.FileCategoryAudio,
					structs.FileCategoryArchive,
					structs.FileCategoryOther,
				}
				isValid := false
				for _, valid := range validCategories {
					if category == valid {
						isValid = true
						break
					}
				}
				if isValid {
					// Add to extras for processing
					if body.Extras == nil {
						body.Extras = &types.JSON{}
					}
					(*body.Extras)["category_override"] = string(category)
				}
			}
		}
	}

	return body, nil
}

// Get handles file retrieval (internal API)
//
// @Summary Get file
// @Description Get details of a specific file
// @Tags Resource
// @Produce json
// @Param slug path string true "File slug"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug} [get]
func (h *fileHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	resp.Success(c.Writer, result.InternalView())
}

// GetPublic handles public file viewing
//
// @Summary Get public file
// @Description Get details of a public file by slug
// @Tags Resource Public
// @Produce json
// @Param slug path string true "File slug"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "file not found or not public"
// @Router /res/view/{slug} [get]
func (h *fileHandler) GetPublic(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	file, err := h.s.File.GetPublic(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("File not found or not public"))
		return
	}

	resp.Success(c.Writer, file.PublicView())
}

// GetShared handles shared file access
//
// @Summary Get shared file
// @Description Access a shared file using share token
// @Tags Resource Public
// @Produce json
// @Param token path string true "Share token"
// @Success 200 {object} structs.ReadFile "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "invalid or expired share token"
// @Router /res/share/{token} [get]
func (h *fileHandler) GetShared(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("token")))
		return
	}

	file, err := h.s.File.GetByShareToken(c.Request.Context(), token)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("Invalid or expired share token"))
		return
	}

	resp.Success(c.Writer, file.PublicView())
}

// Update handles file updates
//
// @Summary Update file
// @Description Update an existing file
// @Tags Resource
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "File slug"
// @Param name formData string false "File name"
// @Param file formData file false "New file content"
// @Param access_level formData string false "Access level (public, private, shared)"
// @Param is_public formData boolean false "Public access flag"
// @Param tags formData string false "Comma-separated tags"
// @Param processing_options formData string false "Processing options (JSON)"
// @Param extras formData string false "Additional properties (JSON)"
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

	existing, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), existing); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	updates := make(types.JSON)

	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to parse form for update: %v", err)
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse form"))
		return
	}

	// Bind form values with comprehensive validation
	for key, values := range c.Request.Form {
		if len(values) > 0 && values[0] != "" {
			switch key {
			case "name":
				updates["name"] = values[0]
			case "original_name":
				updates["original_name"] = values[0]
			case "access_level":
				accessLevel := structs.AccessLevel(values[0])
				switch accessLevel {
				case structs.AccessLevelPublic, structs.AccessLevelPrivate, structs.AccessLevelShared:
					updates["access_level"] = accessLevel
				default:
					resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Invalid access level: %s", values[0])))
					return
				}
			case "category":
				category := structs.FileCategory(values[0])
				validCategories := []structs.FileCategory{
					structs.FileCategoryImage,
					structs.FileCategoryDocument,
					structs.FileCategoryVideo,
					structs.FileCategoryAudio,
					structs.FileCategoryArchive,
					structs.FileCategoryOther,
				}
				isValid := false
				for _, valid := range validCategories {
					if category == valid {
						isValid = true
						break
					}
				}
				if isValid {
					updates["category"] = category
				} else {
					resp.Fail(c.Writer, resp.BadRequest(fmt.Sprintf("Invalid category: %s", values[0])))
					return
				}
			case "expires_at":
				if expiresAt, err := strconv.ParseInt(values[0], 10, 64); err == nil {
					updates["expires_at"] = expiresAt
				} else {
					resp.Fail(c.Writer, resp.BadRequest("Invalid expires_at format"))
					return
				}
			case "tags":
				if values[0] != "" {
					tagList := strings.Split(values[0], ",")
					cleanTags := make([]string, 0, len(tagList))
					for _, tag := range tagList {
						tag = strings.TrimSpace(tag)
						if tag != "" {
							cleanTags = append(cleanTags, tag)
						}
					}
					updates["tags"] = cleanTags
				} else {
					updates["tags"] = []string{} // Clear tags
				}
			case "is_public":
				updates["is_public"] = values[0] == "true" || values[0] == "1"
			case "extras":
				var extras types.JSON
				if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
					resp.Fail(c.Writer, resp.BadRequest("Invalid extras format"))
					return
				}
				updates["extras"] = extras
			case "hash":
				// Allow manual hash update (admin only - should add permission check)
				updates["hash"] = values[0]
			default:
				// Log unknown fields
				logger.Debugf(c.Request.Context(), "Unknown update field: %s = %s", key, values[0])
			}
		}
	}

	// Handle file upload with complete metadata update
	if fileHeaders, ok := c.Request.MultipartForm.File["file"]; ok && len(fileHeaders) > 0 {
		header := fileHeaders[0]

		if header.Filename == "" {
			resp.Fail(c.Writer, resp.BadRequest("Filename cannot be empty"))
			return
		}

		file, err := header.Open()
		if err != nil {
			logger.Errorf(c.Request.Context(), "Failed to open uploaded file: %v", err)
			resp.Fail(c.Writer, resp.BadRequest("Failed to open uploaded file"))
			return
		}
		defer file.Close()

		// Set comprehensive file update fields
		ext := filepath.Ext(header.Filename)
		nameWithoutExt := strings.TrimSuffix(header.Filename, ext)

		updates["name"] = nameWithoutExt
		updates["original_name"] = header.Filename
		updates["size"] = int(header.Size)
		updates["type"] = header.Header.Get("Content-Type")
		if updates["type"] == "" {
			updates["type"] = "application/octet-stream"
		}

		// Update category based on new file type
		newCategory := structs.GetFileCategory(ext)
		updates["category"] = newCategory

		updates["file"] = file

		// Handle processing options for new file
		if processingOptionsStr := c.PostForm("processing_options"); processingOptionsStr != "" {
			var options structs.ProcessingOptions
			if err := json.Unmarshal([]byte(processingOptionsStr), &options); err != nil {
				resp.Fail(c.Writer, resp.BadRequest("Invalid processing options format"))
				return
			}
			if options.CompressionQuality <= 0 || options.CompressionQuality > 100 {
				options.CompressionQuality = 80
			}
			updates["processing_options"] = &options
		}
	}

	result, err := h.s.File.Update(c.Request.Context(), slug, updates)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to update file %s: %v", slug, err)
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result.InternalView())
}

// Delete handles file deletion
//
// @Summary Delete file
// @Description Delete a specific file
// @Tags Resource
// @Param slug path string true "File slug"
// @Success 200 "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug} [delete]
// @Security Bearer
func (h *fileHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
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
// @Summary List files with flexible filtering
// @Description List files based on specified parameters
// @Tags Resource
// @Produce json
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Number of items per page (max 100)" maximum(100)
// @Param direction query string false "Pagination direction" Enums(forward, backward)
// @Param owner_id query string true "Owner ID for filtering"
// @Param user query string false "Created by user filter"
// @Param type query string false "Content type filter"
// @Param storage query string false "Storage provider filter"
// @Param category query string false "File category filter"
// @Param tags query string false "Comma-separated tags filter"
// @Param access_level query string false "Access level filter"
// @Param path_prefix query string false "Path prefix filter"
// @Param created_after query integer false "Created after timestamp"
// @Param created_before query integer false "Created before timestamp"
// @Param size_min query integer false "Minimum file size in bytes"
// @Param size_max query integer false "Maximum file size in bytes"
// @Param is_public query boolean false "Public flag filter"
// @Param q query string false "Search query"
// @Success 200 {object} structs.Result[structs.ReadFile] "Paginated file list"
// @Failure 400 {object} resp.Exception "Bad request"
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

	if err := h.authorizeOwnerAccess(c.Request.Context(), params.OwnerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	files, err := h.s.File.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	for _, file := range files.Items {
		*file = *file.InternalView()
	}

	resp.Success(c.Writer, files)
}

// Search handles file searching
//
// @Summary Search files
// @Description Search files by various criteria
// @Tags Resource
// @Produce json
// @Param owner_id query string true "Owner ID"
// @Param q query string false "Search query"
// @Param category query string false "File category"
// @Param tags query string false "Comma-separated tags"
// @Param is_public query boolean false "Public flag"
// @Param limit query int false "Number of results"
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

	if params.OwnerID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("owner_id")))
		return
	}
	if err := h.authorizeOwnerAccess(c.Request.Context(), params.OwnerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	// Tag-based search
	if params.Tags != "" {
		tags := strings.Split(params.Tags, ",")
		limit := 100
		if params.Limit > 0 {
			limit = params.Limit
		}

		results, err := h.s.File.SearchByTags(c.Request.Context(), params.OwnerID, tags, limit)
		if err != nil {
			logger.Errorf(c.Request.Context(), "Error searching files by tags: %v", err)
			resp.Fail(c.Writer, resp.InternalServer("Failed to search files"))
			return
		}

		internalResults := make([]*structs.ReadFile, len(results))
		for i, file := range results {
			internalResults[i] = file.InternalView()
		}

		resp.Success(c.Writer, internalResults)
		return
	}

	// Standard list with filtering
	results, err := h.s.File.List(c.Request.Context(), &params)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error searching files: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to search files"))
		return
	}

	for _, file := range results.Items {
		*file = *file.InternalView()
	}

	resp.Success(c.Writer, results)
}

// ListCategories handles listing file categories
//
// @Summary List file categories
// @Description List all available file categories
// @Tags Resource
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
// @Description List all tags used in files for an owner
// @Tags Resource
// @Produce json
// @Param owner_id query string true "Owner ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/tags [get]
// @Security Bearer
func (h *fileHandler) ListTags(c *gin.Context) {
	ownerID := c.Query("owner_id")
	if ownerID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("owner_id")))
		return
	}
	if err := h.authorizeOwnerAccess(c.Request.Context(), ownerID); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	tags, err := h.s.File.GetTagsByOwner(c.Request.Context(), ownerID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to get tags"))
		return
	}

	resp.Success(c.Writer, tags)
}

// GetVersions handles file version retrieval
//
// @Summary Get file versions
// @Description Get all versions of a file
// @Tags Resource
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

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve file"))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	versions, err := h.s.File.GetVersions(c.Request.Context(), slug)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Error retrieving versions: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve versions"))
		return
	}

	internalVersions := make([]*structs.ReadFile, len(versions))
	for i, version := range versions {
		internalVersions[i] = version.InternalView()
	}

	resp.Success(c.Writer, internalVersions)
}

// CreateVersion handles file version creation
//
// @Summary Create file version
// @Description Create a new version of an existing file
// @Tags Resource
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

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve file"))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
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

	resp.Success(c.Writer, version.InternalView())
}

// CreateThumbnail handles thumbnail creation
//
// @Summary Create thumbnail
// @Description Create a thumbnail for an image file
// @Tags Resource
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

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve file"))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
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

	resp.Success(c.Writer, file.InternalView())
}

// SetAccessLevel handles access level setting
//
// @Summary Set access level
// @Description Set the access level for a file
// @Tags Resource
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body object{access_level=string} true "Access level request"
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

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve file"))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
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

	resp.Success(c.Writer, file.InternalView())
}

// GenerateShareURL handles share URL generation
//
// @Summary Generate share URL
// @Description Generate a shareable URL for a file
// @Tags Resource
// @Accept json
// @Produce json
// @Param slug path string true "File slug"
// @Param body body object{expiration_hours=int} true "Expiration settings"
// @Success 200 {object} object{url=string,expires_in=string,expires_at=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug}/share [post]
// @Security Bearer
func (h *fileHandler) GenerateShareURL(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	result, err := h.s.File.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to retrieve file"))
		return
	}
	if err := h.authorizeFileAccess(c.Request.Context(), result); err != nil {
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
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

// Download handles file download
//
// @Summary Download file
// @Description Download a file
// @Tags Resource
// @Param slug path string true "File slug"
// @Success 200 "File content"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /res/{slug}/download [get]
// @Security Bearer
func (h *fileHandler) Download(c *gin.Context) {
	h.downloadFile(c, "attachment")
}

// GetThumbnail handles thumbnail viewing
//
// @Summary Get file thumbnail
// @Description Get thumbnail image for a file
// @Tags Resource Public
// @Produce image/jpeg
// @Param slug path string true "File slug"
// @Success 200 "Thumbnail image"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "thumbnail not found"
// @Router /res/thumb/{slug} [get]
func (h *fileHandler) GetThumbnail(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	file, err := h.s.File.GetPublic(c.Request.Context(), slug)
	if err != nil || file == nil {
		resp.Fail(c.Writer, resp.NotFound("File not found or not public"))
		return
	}

	h.serveThumbnail(c, slug)
}

// DownloadPublic handles public file download
//
// @Summary Download public file
// @Description Download a public file directly
// @Tags Resource Public
// @Produce application/octet-stream
// @Param slug path string true "File slug"
// @Success 200 "File content"
// @Failure 400 {object} resp.Exception "bad request"
// @Failure 404 {object} resp.Exception "file not found or not public"
// @Router /res/dl/{slug} [get]
func (h *fileHandler) DownloadPublic(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Check if file is public
	file, err := h.s.File.GetPublic(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("File not found or not public"))
		return
	}

	h.serveFileContent(c, file, "attachment")
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
	if err := h.authorizeFileAccess(c.Request.Context(), row); err != nil {
		fileStream.Close()
		resp.Fail(c.Writer, resp.Forbidden(err.Error()))
		return
	}

	filename := row.GetFilename()
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

// serveFileContent serves file content
func (h *fileHandler) serveFileContent(c *gin.Context, file *structs.ReadFile, disposition string) {
	fileStream, err := h.s.File.GetFileStreamByID(c.Request.Context(), file.ID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to get file stream"))
		return
	}
	defer fileStream.Close()

	filename := file.GetFilename()
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%s", disposition, filename))
	c.Header("Content-Type", file.Type)

	io.Copy(c.Writer, fileStream)
}

// serveThumbnail serves thumbnail image
func (h *fileHandler) serveThumbnail(c *gin.Context, slug string) {
	thumbnail, err := h.s.File.GetThumbnail(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("Thumbnail not found"))
		return
	}

	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 1 day

	io.Copy(c.Writer, thumbnail)
	thumbnail.Close()
}

// authorizeOwnerAccess checks if the user has access to the specified owner ID
func (h *fileHandler) authorizeOwnerAccess(ctx context.Context, ownerID string) error {
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

	if h.s.Space == nil || !h.s.Space.HasUserSpaceService() {
		return fmt.Errorf("space service not available")
	}

	inSpace, err := h.s.Space.IsUserInSpace(ctx, ownerID, userID)
	if err != nil {
		return err
	}
	if !inSpace {
		return fmt.Errorf("owner access denied")
	}

	return nil
}

// authorizeFileAccess checks if the user has access to the specified file
func (h *fileHandler) authorizeFileAccess(ctx context.Context, file *structs.ReadFile) error {
	if file == nil {
		return fmt.Errorf("file not found")
	}
	return h.authorizeOwnerAccess(ctx, file.OwnerID)
}
