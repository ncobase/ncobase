package handler

import (
	"ncobase/plugin/ai/service"
	"ncobase/plugin/ai/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

type RunHandlerInterface interface {
	Get(c *gin.Context)
	List(c *gin.Context)
	Usage(c *gin.Context)
}

type runHandler struct {
	svc service.RunServiceInterface
}

func NewRunHandler(svc service.RunServiceInterface) RunHandlerInterface {
	return &runHandler{svc: svc}
}

func (h *runHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}
	result, err := h.svc.GetByID(c.Request.Context(), id, canManageAI(c))
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound("AI run not found"))
		return
	}
	resp.Success(c.Writer, result)
}

func (h *runHandler) List(c *gin.Context) {
	var query structs.RunQuery
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	result, err := h.svc.List(c.Request.Context(), &query, canManageAI(c))
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *runHandler) Usage(c *gin.Context) {
	var query structs.UsageQuery
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	result, err := h.svc.Usage(c.Request.Context(), &query, canManageAI(c))
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}
