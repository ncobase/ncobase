package handler

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/resp"
	"github.com/ncobase/ncore/pkg/types"

	"github.com/gin-gonic/gin"
)

// ProcessDesignHandlerInterface defines process design handler interface
type ProcessDesignHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	SaveDraft(c *gin.Context)
	PublishDraft(c *gin.Context)
	ValidateDesign(c *gin.Context)
	ImportDesign(c *gin.Context)
	ExportDesign(c *gin.Context)
}

// ProcessDesignHandler implements process design operations
type ProcessDesignHandler struct {
	processDesignService service.ProcessDesignServiceInterface
}

// NewProcessDesignHandler creates new process design handler
func NewProcessDesignHandler(svc *service.Service) ProcessDesignHandlerInterface {
	return &ProcessDesignHandler{
		processDesignService: svc.GetProcessDesign(),
	}
}

// Create handles process design creation
// @Summary Create process design
// @Description Create a new process design
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.ProcessDesignBody true "Process design creation body"
// @Success 200 {object} structs.ReadProcessDesign
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs [post]
// @Security Bearer
func (h *ProcessDesignHandler) Create(c *gin.Context) {
	body := &structs.ProcessDesignBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processDesignService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting process design details
// @Summary Get process design
// @Description Get process design details
// @Tags flow
// @Produce json
// @Param id path string true "Process Design ID"
// @Success 200 {object} structs.ReadProcessDesign
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id} [get]
// @Security Bearer
func (h *ProcessDesignHandler) Get(c *gin.Context) {
	params := &structs.FindProcessDesignParams{
		TemplateID: c.Param("id"),
	}

	result, err := h.processDesignService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles process design update
// @Summary Update process design
// @Description Update process design details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Process Design ID"
// @Param body body structs.UpdateProcessDesignBody true "Process design update body"
// @Success 200 {object} structs.ReadProcessDesign
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id} [put]
// @Security Bearer
func (h *ProcessDesignHandler) Update(c *gin.Context) {
	designID := c.Param("id")
	if designID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateProcessDesignBody{ID: designID}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processDesignService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles process design deletion
// @Summary Delete process design
// @Description Delete a process design
// @Tags flow
// @Produce json
// @Param id path string true "Process Design ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id} [delete]
// @Security Bearer
func (h *ProcessDesignHandler) Delete(c *gin.Context) {
	designID := c.Param("id")
	if designID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.processDesignService.Delete(c.Request.Context(), designID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles process design listing
// @Summary List process designs
// @Description List process designs with pagination
// @Tags flow
// @Produce json
// @Param params query structs.ListProcessDesignParams true "Process design list parameters"
// @Success 200 {array} structs.ReadProcessDesign
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs [get]
// @Security Bearer
func (h *ProcessDesignHandler) List(c *gin.Context) {
	params := &structs.ListProcessDesignParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processDesignService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// SaveDraft handles saving process design draft
// @Summary Save design draft
// @Description Save process design as draft
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Process Design ID"
// @Param design body []byte true "Process design data"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id}/drafts [post]
// @Security Bearer
func (h *ProcessDesignHandler) SaveDraft(c *gin.Context) {
	designID := c.Param("id")
	if designID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	designBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	var design types.JSON
	if err := json.Unmarshal(designBytes, &design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid JSON: "+err.Error()))
		return
	}

	if err := h.processDesignService.SaveDraft(c.Request.Context(), designID, design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// PublishDraft handles publishing process design draft
// @Summary Publish design draft
// @Description Publish process design draft as official version
// @Tags flow
// @Produce json
// @Param id path string true "Process Design ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id}/publish [post]
// @Security Bearer
func (h *ProcessDesignHandler) PublishDraft(c *gin.Context) {
	designID := c.Param("id")
	if designID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.processDesignService.PublishDraft(c.Request.Context(), designID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// ValidateDesign handles process design validation
// @Summary Validate design
// @Description Validate process design structure and rules
// @Tags flow
// @Accept json
// @Produce json
// @Param design body []byte true "Process design data"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/validate [post]
// @Security Bearer
func (h *ProcessDesignHandler) ValidateDesign(c *gin.Context) {
	designBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	var design types.JSON
	if err := json.Unmarshal(designBytes, &design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid JSON: "+err.Error()))
		return
	}

	if err := h.processDesignService.ValidateDesign(c.Request.Context(), design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// ImportDesign handles importing process design
// @Summary Import design
// @Description Import process design from file
// @Tags flow
// @Accept multipart/form-data
// @Produce json
// @Param template_id path string true "Template ID"
// @Param file formData file true "Process design file"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{template_id}/designs/import [post]
// @Security Bearer
func (h *ProcessDesignHandler) ImportDesign(c *gin.Context) {
	templateID := c.Param("template_id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("template_id")))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to read file"))
		return
	}

	f, err := file.Open()
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to open file"))
		return
	}

	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			resp.Fail(c.Writer, resp.BadRequest("Failed to close file"))
		}
	}(f)

	designBytes, err := io.ReadAll(f)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to read file content"))
		return
	}

	var design types.JSON
	if err := json.Unmarshal(designBytes, &design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid JSON: "+err.Error()))
		return
	}

	if err := h.processDesignService.ImportDesign(c.Request.Context(), templateID, design); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// ExportDesign handles exporting process design
// @Summary Export design
// @Description Export process design to file
// @Tags flow
// @Produce application/octet-stream
// @Param id path string true "Process Design ID"
// @Success 200 {file} octet-stream
// @Failure 400 {object} resp.Exception
// @Router /flow/process-designs/{id}/export [get]
// @Security Bearer
func (h *ProcessDesignHandler) ExportDesign(c *gin.Context) {
	designID := c.Param("id")
	if designID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	design, err := h.processDesignService.ExportDesign(c.Request.Context(), designID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	designBytes, err := json.Marshal(design)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Failed to marshal JSON: "+err.Error()))
		return
	}

	// Set headers for file download
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=process-design.json")
	c.Data(200, "application/octet-stream", designBytes)
}
