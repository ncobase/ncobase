package handler

import (
	"ncobase/app/data/structs"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// CreateCasbinRuleHandler handles the creation of a Casbin rule.
//
// @Summary Create Casbin rule
// @Description Create a new Casbin rule.
// @Tags casbin
// @Accept json
// @Produce json
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/policies [post]
// @Security Bearer
func (h *Handler) CreateCasbinRuleHandler(c *gin.Context) {
	body := &structs.CasbinRuleBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateCasbinRuleService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateCasbinRuleHandler handles updating a Casbin rule.
//
// @Summary Update Casbin rule
// @Description Update an existing Casbin rule, either fully or partially.
// @Tags casbin
// @Accept json
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/policies/{id} [put]
// @Security Bearer
func (h *Handler) UpdateCasbinRuleHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.UpdateCasbinRuleService(c, id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetCasbinRuleHandler handles getting a Casbin rule.
//
// @Summary Get Casbin rule
// @Description Retrieve details of a Casbin rule.
// @Tags casbin
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/policies/{id} [get]
// @Security Bearer
func (h *Handler) GetCasbinRuleHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.svc.GetCasbinRuleService(c, id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteCasbinRuleHandler handles deleting a Casbin rule.
//
// @Summary Delete Casbin rule
// @Description Delete an existing Casbin rule.
// @Tags casbin
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/policies/{id} [delete]
// @Security Bearer
func (h *Handler) DeleteCasbinRuleHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.svc.DeleteCasbinRuleService(c, id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListCasbinRuleHandler handles listing Casbin rules.
//
// @Summary List Casbin rules
// @Description Retrieve a list of Casbin rules.
// @Tags casbin
// @Produce json
// @Param params query structs.ListCasbinRuleParams true "ListCasbinRuleParams object"
// @Success 200 {array} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/policies [get]
// @Security Bearer
func (h *Handler) ListCasbinRuleHandler(c *gin.Context) {
	params := &structs.ListCasbinRuleParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	casbinRules, err := h.svc.ListCasbinRulesService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, casbinRules)
}
