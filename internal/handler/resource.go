package handler

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"stocms/internal/data/ent"
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/storage"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// CreateResourceHandler handles the creation of an resource
func (h *Handler) CreateResourceHandler(c *gin.Context) {

	// Handle file upload
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		}
	}(file)

	var body structs.CreateResourceBody

	// Set the file details in the request body
	fileHeader := storage.GetFileHeader(header, "resources")
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = fileHeader.Size

	// Manually bind other fields from the form
	for key, values := range c.Request.Form {
		if len(values) == 0 || (key != "file" && values[0] == "") {
			continue
		}
		switch key {
		case "name":
			body.Name = values[0]
		case "storage":
			body.Storage = values[0]
		case "object_id":
			body.ObjectID = values[0]
		case "domain_id":
			body.DomainID = values[0]
		case "extras":
			var extraProps types.JSON
			if err := json.Unmarshal([]byte(values[0]), &extraProps); err != nil {
				resp.Fail(c.Writer, resp.BadRequest("Invalid extras format"))
				return
			}
			body.ExtraProps = extraProps
		}
	}

	result, err := h.svc.CreateResourceService(c, &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
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
		if updates["name"] == nil {
			updates["name"] = fileHeader.Name
		}
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

// DownloadResourceHandler handles downloading an resource
func (h *Handler) DownloadResourceHandler(c *gin.Context) {
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
		c.Header("Content-Disposition", "resource; filename="+filename)

		_, err := io.Copy(c.Writer, fileStream)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
	}

	// close file stream
	defer func(file io.ReadCloser) {
		err := file.Close()
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		}
	}(fileStream)
}
