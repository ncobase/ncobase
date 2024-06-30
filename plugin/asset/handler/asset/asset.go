package asset

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"ncobase/helper"
	"ncobase/plugin/asset/service"
	"ncobase/plugin/asset/structs"
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

type HandlerInterface interface {
	CreateAssetsHandler(c *gin.Context)
	UpdateAssetHandler(c *gin.Context)
	GetAssetHandler(c *gin.Context)
	ListAssetsHandler(c *gin.Context)
	DeleteAssetHandler(c *gin.Context)
}

type Handler struct {
	s *service.Service
}

func New(s *service.Service) HandlerInterface {
	return &Handler{
		s: s,
	}
}

// maxAssetSize is the maximum allowed size of an asset.
var maxAssetSize int64 = 2048 << 20 // 2048 MB

// CreateAssetsHandler handles the creation of assets, both single and multiple.
//
// @Summary Create assets
// @Description Create one or multiple assets.
// @Tags assets
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param object_id formData string false "Object ID associated with the asset"
// @Param tenant_id formData string false "Tenant ID associated with the asset"
// @Param extras formData string false "Additional properties associated with the asset (JSON format)"
// @Success 200 {object} structs.ReadAsset "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/assets [post]
// @Security Bearer
func (h *Handler) CreateAssetsHandler(c *gin.Context) {
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
		if err := h.validateAssetBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Asset.CreateAssetService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		resp.Success(c.Writer, result)
		return
	}

	err = c.Request.ParseMultipartForm(maxAssetSize) // Set maxMemory to 32MB
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
		if err := h.validateAssetBody(body); err != nil {
			resp.Fail(c.Writer, resp.BadRequest(err.Error()))
			return
		}
		result, err := h.s.Asset.CreateAssetService(c, body)
		if err != nil {
			resp.Fail(c.Writer, resp.InternalServer(err.Error()))
			return
		}
		results = append(results, result.Data)
	}
	resp.Success(c.Writer, &resp.Exception{Data: results})
}

func (h *Handler) validateAssetBody(body *structs.CreateAssetBody) error {
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
// 	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAssetSize)
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
// 	// Create a new CreateAssetBody instance
// 	body := &structs.CreateAssetBody{}
//
// 	body.File = bytes.NewReader(fileBuffer.Bytes())
// 	body.Type = fileType
//
// 	// Call service to create asset
// 	result, err := h.svc.CreateAssetService(c, body)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.InternalServer("Failed to create asset"))
// 		return
// 	}
//
// 	resp.Success(c.Writer, result)
// }

// processFile processes file details and binds other fields from the form to the asset body
func processFile(c *gin.Context, header *multipart.FileHeader, file multipart.File) (*structs.CreateAssetBody, error) {
	body := &structs.CreateAssetBody{}
	fileHeader := storage.GetFileHeader(header, "assets")
	body.Path = fileHeader.Path
	body.File = file
	body.Type = fileHeader.Type
	body.Name = fileHeader.Name
	body.Size = &fileHeader.Size

	// Bind other fields from the form
	if err := bindAssetFields(c, body); err != nil {
		return nil, err
	}
	return body, nil
}

// bindAssetFields binds other fields from the form to the asset body
func bindAssetFields(c *gin.Context, body *structs.CreateAssetBody) error {
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

// UpdateAssetHandler handles updating a asset.
//
// @Summary Update asset
// @Description Update an existing asset.
// @Tags assets
// @Accept multipart/form-data
// @Produce json
// @Param slug path string true "Slug of the asset to update"
// @Param asset body structs.UpdateAssetBody true "Asset details"
// @Success 200 {object} structs.ReadAsset "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/assets/{slug} [put]
// @Security Bearer
func (h *Handler) UpdateAssetHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	// Create a map to hold the updates
	updates := make(types.JSON)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxAssetSize); err != nil {
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
		fileHeader := storage.GetFileHeader(header, "assets")
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

	result, err := h.s.Asset.UpdateAssetService(c, slug, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetAssetHandler handles getting a asset.
//
// @Summary Get asset
// @Description Get details of a specific asset.
// @Tags assets
// @Produce json
// @Param slug path string true "Slug of the asset to retrieve"
// @Param type query string false "Type of retrieval ('download' or 'stream')"
// @Success 200 {object} structs.ReadAsset "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/assets/{slug} [get]
func (h *Handler) GetAssetHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	if c.Query("type") == "download" {
		h.downloadAssetHandler(c)
		return
	}

	if c.Query("type") == "stream" {
		h.assetStreamHandler(c)
		return
	}

	result, err := h.s.Asset.GetAssetService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteAssetHandler handles deleting a asset.
//
// @Summary Delete asset
// @Description Delete a specific asset.
// @Tags assets
// @Param slug path string true "Slug of the asset to delete"
// @Success 200 {object} structs.ReadAsset "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/assets/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteAssetHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("file")))
		return
	}

	result, err := h.s.Asset.DeleteAssetService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListAssetsHandler handles listing assets.
//
// @Summary List assets
// @Description List assets based on specified parameters.
// @Tags assets
// @Produce json
// @Param params query structs.ListAssetParams true "List assets parameters"
// @Success 200 {array} structs.ReadAsset "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/assets [get]
func (h *Handler) ListAssetsHandler(c *gin.Context) {
	params := &structs.ListAssetParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	assets, err := h.s.Asset.ListAssetsService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, assets)
}

// downloadAssetHandler handles the direct download of a asset.
func (h *Handler) downloadAssetHandler(c *gin.Context) {
	h.downloadFile(c, "attachment")
}

// assetStreamHandler handles the streaming of a asset.
func (h *Handler) assetStreamHandler(c *gin.Context) {
	h.downloadFile(c, "inline")
}

// downloadFile handles the download or streaming of a asset
func (h *Handler) downloadFile(c *gin.Context, dispositionType string) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}

	fileStream, exception := h.s.Asset.GetFileStream(c, slug)
	if exception != nil {
		if exception.Code != 0 {
			resp.Fail(c.Writer, exception)
			return
		}
		row := exception.Data.(*structs.ReadAsset)
		filename := storage.RestoreOriginalFileName(row.Path, true)
		c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%s", dispositionType, filename))

		// Set the Content-Type header based on the original content t
		if row.Type == "" {
			c.Header("Content-Type", "application/octet-stream")
		}
		c.Header("Content-Type", row.Type)

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
