package handler

import (
	"ncobase/plugin/ai/service"
	"ncobase/plugin/ai/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
)

type ActionHandlerInterface interface {
	List(c *gin.Context)
	Run(c *gin.Context)
}

type actionHandler struct {
	svc service.AIServiceInterface
}

func NewActionHandler(svc service.AIServiceInterface) ActionHandlerInterface {
	return &actionHandler{svc: svc}
}

func (h *actionHandler) List(c *gin.Context) {
	result, err := h.svc.Actions(c.Request.Context())
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}

func (h *actionHandler) Run(c *gin.Context) {
	action := c.Param("action")
	if action == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("action")))
		return
	}
	var input structs.ActionRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request body", err.Error()))
		return
	}
	result, err := h.svc.RunAction(c.Request.Context(), action, &input)
	if err != nil {
		handleAIError(c, err)
		return
	}
	resp.Success(c.Writer, result)
}
