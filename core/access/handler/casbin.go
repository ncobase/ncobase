package handler

import (
	"ncobase/core/access/service"
	"ncobase/core/access/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// CasbinHandlerInterface is the interface for the handler.
type CasbinHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// casbinHandler represents the handler.
type casbinHandler struct {
	s *service.Service
}

// NewCasbinHandler creates a new handler.
func NewCasbinHandler(svc *service.Service) CasbinHandlerInterface {
	return &casbinHandler{
		s: svc,
	}
}

// Create handles the creation of a Casbin rule.
//
// @Summary Create Casbin rule
// @Description Create a new Casbin rule.
// @Tags iam
// @Accept json
// @Produce json
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/policies [post]
// @Security Bearer
func (h *casbinHandler) Create(c *gin.Context) {
	body := &structs.CasbinRuleBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Casbin.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a Casbin rule.
//
// @Summary Update Casbin rule
// @Description Update an existing Casbin rule, either fully or partially.
// @Tags iam
// @Accept json
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Param body body structs.CasbinRuleBody true "CasbinRuleBody object"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/policies/{id} [put]
// @Security Bearer
func (h *casbinHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Casbin.Update(c.Request.Context(), id, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a Casbin rule.
//
// @Summary Get Casbin rule
// @Description Retrieve details of a Casbin rule.
// @Tags iam
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} structs.ReadCasbinRule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/policies/{id} [get]
// @Security Bearer
func (h *casbinHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.s.Casbin.Get(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a Casbin rule.
//
// @Summary Delete Casbin rule
// @Description Delete an existing Casbin rule.
// @Tags iam
// @Produce json
// @Param id path string true "Casbin rule ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/policies/{id} [delete]
// @Security Bearer
func (h *casbinHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.s.Casbin.Delete(c.Request.Context(), id); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing Casbin rules.
//
// @Summary List Casbin rules
// @Description Retrieve a list of Casbin rules.
// @Tags iam
// @Produce json
// @Param params query structs.ListCasbinRuleParams true "ListCasbinRuleParams object"
// @Success 200 {array} structs.CasbinRuleBody "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/policies [get]
// @Security Bearer
func (h *casbinHandler) List(c *gin.Context) {
	params := &structs.ListCasbinRuleParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	casbinRules, err := h.s.Casbin.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, casbinRules)
}
