package handler

import (
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"
	"ncobase/ncore/ecode"
	"ncobase/ncore/helper"
	"ncobase/ncore/resp"

	"github.com/gin-gonic/gin"
)

// RuleHandlerInterface defines rule handler interface
type RuleHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Enable(c *gin.Context)
	Disable(c *gin.Context)
	ValidateRule(c *gin.Context)
	GetActiveRules(c *gin.Context)
	EvaluateRules(c *gin.Context)
}

// RuleHandler implements rule operations
type RuleHandler struct {
	ruleService service.RuleServiceInterface
}

// NewRuleHandler creates new rule handler
func NewRuleHandler(svc *service.Service) RuleHandlerInterface {
	return &RuleHandler{
		ruleService: svc.GetRule(),
	}
}

// Create handles rule creation
// @Summary Create rule
// @Description Create a new workflow rule
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.RuleBody true "Rule creation body"
// @Success 200 {object} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/rules [post]
// @Security Bearer
func (h *RuleHandler) Create(c *gin.Context) {
	body := &structs.RuleBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.ruleService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting rule details
// @Summary Get rule
// @Description Get workflow rule details
// @Tags flow
// @Produce json
// @Param id path string true "Rule Key"
// @Success 200 {object} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id} [get]
// @Security Bearer
func (h *RuleHandler) Get(c *gin.Context) {
	params := &structs.FindRuleParams{
		RuleKey: c.Param("id"),
	}

	result, err := h.ruleService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles rule update
// @Summary Update rule
// @Description Update workflow rule details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param body body structs.UpdateRuleBody true "Rule update body"
// @Success 200 {object} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id} [put]
// @Security Bearer
func (h *RuleHandler) Update(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateRuleBody{ID: ruleID}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.ruleService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles rule deletion
// @Summary Delete rule
// @Description Delete a workflow rule
// @Tags flow
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id} [delete]
// @Security Bearer
func (h *RuleHandler) Delete(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.ruleService.Delete(c.Request.Context(), ruleID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles rule listing
// @Summary List rules
// @Description List workflow rules with pagination
// @Tags flow
// @Produce json
// @Param params query structs.ListRuleParams true "Rule list parameters"
// @Success 200 {array} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/rules [get]
// @Security Bearer
func (h *RuleHandler) List(c *gin.Context) {
	params := &structs.ListRuleParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.ruleService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Enable handles rule enabling
// @Summary Enable rule
// @Description Enable a workflow rule
// @Tags flow
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id}/enable [post]
// @Security Bearer
func (h *RuleHandler) Enable(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.ruleService.EnableRule(c.Request.Context(), ruleID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Disable handles rule disabling
// @Summary Disable rule
// @Description Disable a workflow rule
// @Tags flow
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id}/disable [post]
// @Security Bearer
func (h *RuleHandler) Disable(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.ruleService.DisableRule(c.Request.Context(), ruleID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// ValidateRule handles rule validation
// @Summary Validate rule
// @Description Validate a workflow rule
// @Tags flow
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/{id}/validate [post]
// @Security Bearer
func (h *RuleHandler) ValidateRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.ruleService.ValidateRule(c.Request.Context(), ruleID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetActiveRules handles getting active rules
// @Summary Get active rules
// @Description Get active rules for template or node
// @Tags flow
// @Produce json
// @Param template_id query string false "Template ID"
// @Param node_key query string false "Node Key"
// @Success 200 {array} structs.ReadRule
// @Failure 400 {object} resp.Exception
// @Router /flow/rules/active [get]
// @Security Bearer
func (h *RuleHandler) GetActiveRules(c *gin.Context) {
	templateID := c.Query("template_id")
	nodeKey := c.Query("node_key")

	result, err := h.ruleService.GetActiveRules(c.Request.Context(), templateID, nodeKey)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// EvaluateRules handles rule evaluation
// @Summary Evaluate rules
// @Description Evaluate rules for a process instance
// @Tags flow
// @Accept json
// @Produce json
// @Param process_id path string true "Process ID"
// @Param data body map[string]interface{} true "Data for rule evaluation"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{process_id}/evaluate-rules [post]
// @Security Bearer
func (h *RuleHandler) EvaluateRules(c *gin.Context) {
	processID := c.Param("process_id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("process_id")))
		return
	}

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := h.ruleService.EvaluateRules(c.Request.Context(), processID, data); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}
