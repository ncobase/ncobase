package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/log"
	"stocms/pkg/resp"
	"stocms/pkg/storage"
	"stocms/pkg/types"
	"stocms/pkg/validator"
	"strings"

	"github.com/gin-gonic/gin"
)

var maxResourceSize int64 = 2048 << 20 // 2048 MB

// CreateResourcesHandler handles the creation of resources, both single and multiple.
//
// @Summary Create resources
// @Description Create one or multiple resources.
// @Tags resources
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param object_id formData string false "Object ID associated with the resource"
// @Param domain_id formData string false "Domain ID associated with the resource"
// @Param extras formData string false "Additional properties associated with the resource (JSON format)"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources [post]
func (h *Handler) CreateResourcesHandler(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		resp.Fail(c.Writer, resp.NotAllowed("Method not allowed"))
		return
	}
	contentType := c.ContentType()
	switch {
	case strings.HasPrefix(contentType, "multipart/"):
		h.handleFormDataUpload(c)
	// case contentType == "application/octet-stream":
	// 	h.handleBlobUpload(c)
	default:
		resp.Fail(c.Writer, resp.BadRequest("Unsupported content type"))
	}
}

// handleFormDataUpload handles file upload using multipart form data
func (h *Handler) handleFormDataUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Errorf(c, "Error closing file: %v\n", err)
			}
		}(file)
		body, err := processFile(c, header, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateResourceBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.svc.CreateResourceService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		resp.Success(c.Writer, result)
		return
	}

	err = c.Request.ParseMultipartForm(maxResourceSize) // Set maxMemory to 32MB
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to parse multipart form"))
		return
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	var results []any
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
				log.Errorf(c, "Error closing file: %v\n", err)
			}
		}(file)

		body, err := processFile(c, fileHeader, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateResourceBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.svc.CreateResourceService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result.Data)
	}
	resp.Success(c.Writer, &resp.Exception{Data: results})
}

func (h *Handler) validateResourceBody(body *structs.CreateResourceBody) error {
	if validator.IsEmpty(body.ObjectID) {
		return errors.New("belongsTo object is required")
	}
	if validator.IsEmpty(body.DomainID) {
		return errors.New("belongsTo domain is required")
	}
	return nil
}

// // handleBlobUpload handles file upload using Blob object
// func (h *Handler) handleBlobUpload(c *gin.Context) {
// 	// Limit the maximum request body size
// 	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxResourceSize)
//
// 	// Create a buffer to store the file data
// 	fileBuffer := bytes.Buffer{}
//
// 	// Copy data from request body to buffer
// 	_, err := io.Copy(&fileBuffer, c.Request.Body)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.InternalServer("Failed to read request body"))
// 		return
// 	}
//
// 	// Check if file is empty
// 	if fileBuffer.Len() == 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("No file provided"))
// 		return
// 	}
//
// 	// Determine file type
// 	fileType := http.DetectContentType(fileBuffer.Bytes())
//
// 	// Create a new CreateResourceBody instance
// 	body := &structs.CreateResourceBody{}
//
// 	body.File = bytes.NewReader(fileBuffer.Bytes())
// 	body.Type = fileType
//
// 	// Call service to create resource
// 	result, err := h.svc.CreateResourceService(c, body)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.InternalServer("Failed to create resource"))
// 		return
// 	}
//
// 	resp.Success(c.Writer, result)
// }

// processFile processes file details and binds other fields from the form to the resource body
func processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateResourceBody, error) {
	body := &structs.CreateResourceBody{}
	fileHeader := storage.GetFileHeader(header, "resources")
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = &fileHeader.Size

	// Bind other fields from the form
	if err := bindResourceFields(c, body); err != nil {
		return nil, err
	}
	return body, nil
}

// bindResourceFields binds other fields from the form to the resource body
func bindResourceFields(c *gin.Context, body *structs.CreateResourceBody) error {
	// Manually bind other fields from the form
	for key, values := range c.Request.Form {
		if len(values) == 0 || (key != "file" && values[0] == "") {
			continue
		}
		switch key {
		case "object_id":
			body.ObjectID = values[0]
		case "domain_id":
			body.DomainID = values[0]
		case "extras":
			var extras types.JSON
			if err := json.Unmarshal([]byte(values[0]), &extras); err != nil {
				return errors.New("invalid extras format")
			}
			body.Extras = &extras
		}
	}
	return nil
}

// UpdateResourceHandler handles updating a resource.
//
// @Summary Update resource
// @Description Update an existing resource.
// @Tags resources
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "Slug of the resource to update"
// @Param file formData file false "File to upload (optional)"
// @Param object_id formData string false "Object ID associated with the resource"
// @Param domain_id formData string false "Domain ID associated with the resource"
// @Param extras formData string false "Additional properties associated with the resource (JSON format)"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources/{slug} [put]
func (h *Handler) UpdateResourceHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Create a map to hold the updates
	updates := make(types.JSON)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxResourceSize); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to parse form"))
		return
	}

	// Bind form values to updates
	for key, values := range c.Request.Form {
		if len(values) > 0 && values[0] != "" {
			updates[key] = values[0]
		}
	}

	// Check if the file is included in the request
	if fileHeaders, ok := c.Request.MultipartForm.File["file"]; ok && len(fileHeaders) > 0 {
		// Fetch file header from request
		header := fileHeaders[0]
		// Get file data
		fileHeader := storage.GetFileHeader(header, "resources")
		// Add file header data to updates
		// if updates["name"] == nil {
		updates["name"] = fileHeader.Name
		// }
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
	}

	result, err := h.svc.UpdateResourceService(c, slug, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetResourceHandler handles getting a resource.
//
// @Summary Get resource
// @Description Get details of a specific resource.
// @Tags resources
// @Produce json
// @Param slug path string true "Slug of the resource to retrieve"
// @Param type query string false "Type of retrieval ('download' or 'stream')"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources/{slug} [get]
func (h *Handler) GetResourceHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	if c.Query("type") == "download" {
		h.DownloadResourceHandler(c)
		return
	}

	if c.Query("type") == "stream" {
		h.ResourceStreamHandler(c)
		return
	}

	result, err := h.svc.GetResourceService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteResourceHandler handles deleting a resource.
//
// @Summary Delete resource
// @Description Delete a specific resource.
// @Tags resources
// @Param slug path string true "Slug of the resource to delete"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources/{slug} [delete]
func (h *Handler) DeleteResourceHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	result, err := h.svc.DeleteResourceService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListResourceHandler handles listing resources.
//
// @Summary List resources
// @Description List resources based on specified parameters.
// @Tags resources
// @Produce json
// @Param page query integer false "Page number"
// @Param page_size query integer false "Page size"
// @Param sort_by query string false "Sort by field"
// @Param order query string false "Sort order ('asc' or 'desc')"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources [get]
func (h *Handler) ListResourceHandler(c *gin.Context) {
	params := &structs.ListResourceParams{}
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resources, err := h.svc.ListResourcesService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, resources)
}

// DownloadResourceHandler handles the direct download of a resource.
//
// @Summary Download resource
// @Description Download a specific resource.
// @Tags resources
// @Produce octet-stream
// @Param slug path string true "Slug of the resource to download"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources/{slug}/download [get]
func (h *Handler) DownloadResourceHandler(c *gin.Context) {
	h.downloadFile(c, "attachment")
}

// ResourceStreamHandler handles the streaming of a resource.
//
// @Summary Stream resource
// @Description Stream a specific resource.
// @Tags resources
// @Produce octet-stream
// @Param slug path string true "Slug of the resource to stream"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resources/{slug}/stream [get]
func (h *Handler) ResourceStreamHandler(c *gin.Context) {
	h.downloadFile(c, "inline")
}

// downloadFile handles the download or streaming of a resource
func (h *Handler) downloadFile(c *gin.Context, dispositionType string) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	fileStream, exception := h.svc.GetFileStream(c, slug)
	if exception != nil {
		if exception.Code != 0 {
			resp.Fail(c.Writer, exception)
			return
		}
		resource := exception.Data.(*ent.Resource)
		filename := storage.RestoreOriginalFileName(resource.Name, true)
		c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%s", dispositionType, filename))

		// Set the Content-Type header based on the original content t
		if resource.Type == "" {
			c.Header("Content-Type", "application/octet-stream")
		}
		c.Header("Content-Type", resource.Type)

		_, err := io.Copy(c.Writer, fileStream)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
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
