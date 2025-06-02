package handler

import (
	"ncobase/workflow/service"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// TemplateHandlerInterface defines template handler interface
type TemplateHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	CreateVersion(c *gin.Context)
	SetLatestVersion(c *gin.Context)
	Enable(c *gin.Context)
	Disable(c *gin.Context)
	ValidateTemplate(c *gin.Context)
	GetDesigns(c *gin.Context)
	GetRules(c *gin.Context)
}

// TemplateHandler implements template operations
type TemplateHandler struct {
	templateService      service.TemplateServiceInterface
	processDesignService service.ProcessDesignServiceInterface
	ruleService          service.RuleServiceInterface
}

// NewTemplateHandler creates new template handler
func NewTemplateHandler(svc *service.Service) TemplateHandlerInterface {
	return &TemplateHandler{
		templateService:      svc.GetTemplate(),
		processDesignService: svc.GetProcessDesign(),
		ruleService:          svc.GetRule(),
	}
}

// Create handles template creation
// @Summary Create template
// @Description Create a new workflow template
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.TemplateBody true "Template creation body"
// @Success 200 {object} structs.ReadTemplate
// @Failure 400 {object} resp.Exception
// @Router /flow/templates [post]
// @Security Bearer
func (h *TemplateHandler) Create(c *gin.Context) {
	body := &structs.TemplateBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.templateService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting template details
// @Summary Get template
// @Description Get workflow template details
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} structs.ReadTemplate
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id} [get]
// @Security Bearer
func (h *TemplateHandler) Get(c *gin.Context) {
	params := &structs.FindTemplateParams{
		Code: c.Param("id"),
	}

	result, err := h.templateService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles template update
// @Summary Update template
// @Description Update workflow template details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param body body structs.UpdateTemplateBody true "Template update body"
// @Success 200 {object} structs.ReadTemplate
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id} [put]
// @Security Bearer
func (h *TemplateHandler) Update(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateTemplateBody{ID: templateID}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.templateService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles template deletion
// @Summary Delete template
// @Description Delete a workflow template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id} [delete]
// @Security Bearer
func (h *TemplateHandler) Delete(c *gin.Context) {
	params := &structs.FindTemplateParams{
		Code: c.Param("id"),
	}

	if err := h.templateService.Delete(c.Request.Context(), params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles template listing
// @Summary List templates
// @Description List workflow templates
// @Tags flow
// @Produce json
// @Param params query structs.ListTemplateParams true "Template list parameters"
// @Success 200 {array} structs.ReadTemplate
// @Failure 400 {object} resp.Exception
// @Router /flow/templates [get]
// @Security Bearer
func (h *TemplateHandler) List(c *gin.Context) {
	params := &structs.ListTemplateParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.templateService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// CreateVersion handles creating new template version
// @Summary Create template version
// @Description Create a new version of existing template
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Template ID"
// @Param version query string true "New Version"
// @Success 200 {object} structs.ReadTemplate
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/versions [post]
// @Security Bearer
func (h *TemplateHandler) CreateVersion(c *gin.Context) {
	templateID := c.Param("id")
	version := c.Query("version")

	if templateID == "" || version == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	result, err := h.templateService.CreateVersion(c.Request.Context(), templateID, version)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// SetLatestVersion handles setting a version as latest
// @Summary Set latest version
// @Description Set a template version as the latest version
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/versions/latest [put]
// @Security Bearer
func (h *TemplateHandler) SetLatestVersion(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.templateService.SetLatestVersion(c.Request.Context(), templateID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Enable handles enabling a template
// @Summary Enable template
// @Description Enable a workflow template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/enable [post]
// @Security Bearer
func (h *TemplateHandler) Enable(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.templateService.Enable(c.Request.Context(), templateID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Disable handles disabling a template
// @Summary Disable template
// @Description Disable a workflow template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/disable [post]
// @Security Bearer
func (h *TemplateHandler) Disable(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.templateService.Disable(c.Request.Context(), templateID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// ValidateTemplate handles template validation
// @Summary Validate template
// @Description Validate a workflow template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/validate [post]
// @Security Bearer
func (h *TemplateHandler) ValidateTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.templateService.ValidateTemplate(c.Request.Context(), templateID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetDesigns handles getting template process designs
// @Summary Get template designs
// @Description Get process designs associated with a template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {array} structs.ReadProcessDesign
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/designs [get]
// @Security Bearer
func (h *TemplateHandler) GetDesigns(c *gin.Context) {
	params := &structs.ListProcessDesignParams{
		TemplateID: c.Param("id"),
	}

	result, err := h.processDesignService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetRules handles getting template rules
// @Summary Get template rules
// @Description Get rules associated with a template
// @Tags flow
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {array} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/templates/{id}/rules [get]
// @Security Bearer
func (h *TemplateHandler) GetRules(c *gin.Context) {
	params := &structs.ListRuleParams{
		TemplateID: c.Param("id"),
	}

	result, err := h.ruleService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
