package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"ncobase/common/helper"
	"ncobase/domain/resource/service"
	"ncobase/domain/resource/structs"
	"net/http"
	"strings"

	"ncobase/common/ecode"
	"ncobase/common/log"
	"ncobase/common/resp"
	"ncobase/common/storage"
	"ncobase/common/types"
	"ncobase/common/validator"

	"github.com/gin-gonic/gin"
)

// AttachmentHandlerInterface represents the attachment handler interface.
type AttachmentHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

// attachmentHandler represents the attachment handler.
type attachmentHandler struct {
	s *service.Service
}

// NewAttachmentHandler creates a new attachment handler.
func NewAttachmentHandler(s *service.Service) AttachmentHandlerInterface {
	return &attachmentHandler{
		s: s,
	}
}

// maxAttachmentSize is the maximum allowed size of an attachment.
var maxAttachmentSize int64 = 2048 << 20 // 2048 MB

// Create handles the creation of attachments, both single and multiple.
//
// @Summary Create attachments
// @Description Create one or multiple attachments.
// @Tags attachments
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param object_id formData string false "Object ID associated with the attachment"
// @Param tenant_id formData string false "Tenant ID associated with the attachment"
// @Param extras formData string false "Additional properties associated with the attachment (JSON format)"
// @Success 200 {object} structs.ReadAttachment "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resource/attachments [post]
// @Security Bearer
func (h *attachmentHandler) Create(c *gin.Context) {
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
		return
	}
}

// handleFormDataUpload handles file upload using multipart form data
func (h *attachmentHandler) handleFormDataUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Errorf(c, "Error closing file: %v", err)
			}
		}(file)
		body, err := processFile(c, header, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateAttachmentBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Attachment.Create(c.Request.Context(), body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		resp.Success(c.Writer, result)
		return
	}

	err = c.Request.ParseMultipartForm(maxAttachmentSize) // Set maxMemory to 32MB
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to parse multipart form"))
		return
	}
	files := c.Request.MultipartForm.File["files"]
	if len(files) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("File is required"))
		return
	}
	var results []*structs.ReadAttachment
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
				log.Errorf(c, "Error closing file: %v", err)
			}
		}(file)

		body, err := processFile(c, fileHeader, file)
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		if err := h.validateAttachmentBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Attachment.Create(c.Request.Context(), body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result)
	}
	resp.Success(c.Writer, results)
}

func (h *attachmentHandler) validateAttachmentBody(body *structs.CreateAttachmentBody) error {
	if validator.IsEmpty(body.ObjectID) {
		return errors.New("belongsTo object is required")
	}
	if validator.IsEmpty(body.TenantID) {
		return errors.New("belongsTo tenant is required")
	}
	return nil
}

// // handleBlobUpload handles file upload using Blob object
// func (h *Handler) handleBlobUpload(c *gin.Context) {
// 	// Limit the maximum request body size
// 	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAttachmentSize)
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
// 	// Create a new CreateAttachmentBody instance
// 	body := &structs.CreateAttachmentBody{}
//
// 	body.File = bytes.NewReader(fileBuffer.Bytes())
// 	body.Type = fileType
//
// 	// Call service to create attachment
// 	result, err := h.svc.Create(c, body)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.InternalServer("Failed to create attachment"))
// 		return
// 	}
//
// 	resp.Success(c.Writer, result)
// }

// processFile processes file details and binds other fields from the form to the attachment body
func processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateAttachmentBody, error) {
	body := &structs.CreateAttachmentBody{}
	fileHeader := storage.GetFileHeader(header, "attachments")
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = &fileHeader.Size

	// Bind other fields from the form
	if err := bindAttachmentFields(c, body); err != nil {
		return nil, err
	}
	return body, nil
}

// bindAttachmentFields binds other fields from the form to the attachment body
func bindAttachmentFields(c *gin.Context, body *structs.CreateAttachmentBody) error {
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

// Update handles updating a attachment.
//
// @Summary Update attachment
// @Description Update an existing attachment.
// @Tags attachments
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "Slug of the attachment to update"
// @Param attachment body structs.UpdateAttachmentBody true "Attachment details"
// @Success 200 {object} structs.ReadAttachment "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resource/attachments/{slug} [put]
// @Security Bearer
func (h *attachmentHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Create a map to hold the updates
	updates := make(types.JSON)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxAttachmentSize); err != nil {
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
		fileHeader := storage.GetFileHeader(header, "attachments")
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

	result, err := h.s.Attachment.Update(c.Request.Context(), slug, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a attachment.
//
// @Summary Get attachment
// @Description Get details of a specific attachment.
// @Tags attachments
// @Produce json
// @Param slug path string true "Slug of the attachment to retrieve"
// @Param type query string false "Type of retrieval ('download' or 'stream')"
// @Success 200 {object} structs.ReadAttachment "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resource/attachments/{slug} [get]
func (h *attachmentHandler) Get(c *gin.Context) {
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
		h.attachmentStream(c)
		return
	}

	result, err := h.s.Attachment.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a attachment.
//
// @Summary Delete attachment
// @Description Delete a specific attachment.
// @Tags attachments
// @Param slug path string true "Slug of the attachment to delete"
// @Success 200 {object} structs.ReadAttachment "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resource/attachments/{slug} [delete]
// @Security Bearer
func (h *attachmentHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	if err := h.s.Attachment.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing attachments.
//
// @Summary List attachments
// @Description List attachments based on specified parameters.
// @Tags attachments
// @Produce json
// @Param params query structs.ListAttachmentParams true "List attachments parameters"
// @Success 200 {array} structs.ReadAttachment "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /resource/attachments [get]
func (h *attachmentHandler) List(c *gin.Context) {
	params := &structs.ListAttachmentParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	attachments, err := h.s.Attachment.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, attachments)
}

// downloadAttachmentHandler handles the direct download of a attachment.
func (h *attachmentHandler) download(c *gin.Context) {
	h.downloadFile(c, "attachment")
}

// attachmentStreamHandler handles the streaming of a attachment.
func (h *attachmentHandler) attachmentStream(c *gin.Context) {
	h.downloadFile(c, "inline")
}

// downloadFile handles the download or streaming of a attachment
func (h *attachmentHandler) downloadFile(c *gin.Context, dispositionType string) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	fileStream, row, err := h.s.Attachment.GetFileStream(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	filename := storage.RestoreOriginalFileName(row.Path, true)
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%s", dispositionType, filename))

	// Set the Content-Type header based on the original content t
	if row.Type == "" {
		c.Header("Content-Type", "application/octet-stream")
	}
	c.Header("Content-Type", row.Type)

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
