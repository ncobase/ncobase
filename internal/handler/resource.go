package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/log"
	"stocms/pkg/resp"
	"stocms/pkg/storage"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// CreateResourcesHandler handles the creation of resources, both single and multiple
func (h *Handler) CreateResourcesHandler(c *gin.Context) {
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
		result, err := h.svc.CreateResourceService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		resp.Success(c.Writer, result)
		return
	}

	err = c.Request.ParseMultipartForm(32 << 20) // Set maxMemory to 32MB
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to parse multipart form"))
		return
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	var results []interface{}
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer("Failed to open file"))
			return
		}
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
		result, err := h.svc.CreateResourceService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result.Data)
	}
	resp.Success(c.Writer, &resp.Exception{Data: results})
}

// processFile processes file details and binds other fields from the form to the resource body
func processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateResourceBody, error) {
	body := &structs.CreateResourceBody{}
	fileHeader := storage.GetFileHeader(header, "resources")
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = fileHeader.Size

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
			var extraProps types.JSON
			if err := json.Unmarshal([]byte(values[0]), &extraProps); err != nil {
				return errors.New("invalid extras format")
			}
			body.ExtraProps = extraProps
		}
	}
	return nil
}

// UpdateResourceHandler handles updating an resource
func (h *Handler) UpdateResourceHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Create a map to hold the updates
	updates := make(types.JSON)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
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

// GetResourceHandler handles getting an resource
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

// DeleteResourceHandler handles deleting an resource
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

// ListResourceHandler handles listing resources
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

// DownloadResourceHandler handles the direct download of an resource
func (h *Handler) DownloadResourceHandler(c *gin.Context) {
	h.downloadFile(c, "attachment")
}

// ResourceStreamHandler handles the streaming of an resource
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

		// Set the Content-Type header based on the original content type
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
