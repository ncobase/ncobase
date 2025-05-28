package handler

import (
	"ncobase/workflow/service"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// ProcessHandlerInterface defines process handler interface
type ProcessHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Start(c *gin.Context)
	Complete(c *gin.Context)
	Terminate(c *gin.Context)
	Suspend(c *gin.Context)
	Resume(c *gin.Context)
}

// ProcessHandler implements process operations
type ProcessHandler struct {
	processService service.ProcessServiceInterface
}

// NewProcessHandler creates new process handler
func NewProcessHandler(svc *service.Service) ProcessHandlerInterface {
	return &ProcessHandler{
		processService: svc.GetProcess(),
	}
}

// Create handles process creation
// @Summary Create process
// @Description Create a new process instance
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.ProcessBody true "Process creation body"
// @Success 200 {object} structs.ReadProcess
// @Failure 400 {object} resp.Exception
// @Router /flow/processes [post]
// @Security Bearer
func (h *ProcessHandler) Create(c *gin.Context) {
	body := &structs.ProcessBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting process details
// @Summary Get process
// @Description Get process instance details
// @Tags flow
// @Produce json
// @Param id path string true "Process Key"
// @Success 200 {object} structs.ReadProcess
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id} [get]
// @Security Bearer
func (h *ProcessHandler) Get(c *gin.Context) {
	params := &structs.FindProcessParams{
		ProcessKey: c.Param("id"),
	}

	result, err := h.processService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles process update
// @Summary Update process
// @Description Update process instance details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Process ID"
// @Param body body structs.UpdateProcessBody true "Process update body"
// @Success 200 {object} structs.ReadProcess
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id} [put]
// @Security Bearer
func (h *ProcessHandler) Update(c *gin.Context) {
	processID := c.Param("id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateProcessBody{ID: processID}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles process deletion
// @Summary Delete process
// @Description Delete process instance
// @Tags flow
// @Produce json
// @Param id path string true "Process Key"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id} [delete]
// @Security Bearer
func (h *ProcessHandler) Delete(c *gin.Context) {
	params := &structs.FindProcessParams{
		ProcessKey: c.Param("id"),
	}

	if err := h.processService.Delete(c.Request.Context(), params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles process listing
// @Summary List processes
// @Description List process instances with pagination
// @Tags flow
// @Produce json
// @Param params query structs.ListProcessParams true "Process list parameters"
// @Success 200 {array} structs.ReadProcess
// @Failure 400 {object} resp.Exception
// @Router /flow/processes [get]
// @Security Bearer
func (h *ProcessHandler) List(c *gin.Context) {
	params := &structs.ListProcessParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Start handles starting a process
// @Summary Start process
// @Description Start a new process instance
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.StartProcessRequest true "Process start request"
// @Success 200 {object} structs.StartProcessResponse
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/start [post]
// @Security Bearer
func (h *ProcessHandler) Start(c *gin.Context) {
	req := &structs.StartProcessRequest{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.processService.Start(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Complete handles completing a process
// @Summary Complete process
// @Description Complete a process instance
// @Tags flow
// @Produce json
// @Param id path string true "Process ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id}/complete [post]
// @Security Bearer
func (h *ProcessHandler) Complete(c *gin.Context) {
	processID := c.Param("id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.processService.Complete(c.Request.Context(), processID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Terminate handles terminating a process
// @Summary Terminate process
// @Description Terminate a process instance
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Process ID"
// @Param body body structs.TerminateProcessRequest true "Process termination request"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id}/terminate [post]
// @Security Bearer
func (h *ProcessHandler) Terminate(c *gin.Context) {
	req := &structs.TerminateProcessRequest{
		ProcessID: c.Param("id"),
	}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.processService.Terminate(c.Request.Context(), req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Suspend handles suspending a process
// @Summary Suspend process
// @Description Suspend a process instance
// @Tags flow
// @Produce json
// @Param id path string true "Process ID"
// @Param reason query string false "Suspension reason"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id}/suspend [post]
// @Security Bearer
func (h *ProcessHandler) Suspend(c *gin.Context) {
	processID := c.Param("id")
	reason := c.Query("reason")

	if err := h.processService.Suspend(c.Request.Context(), processID, reason); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Resume handles resuming a suspended process
// @Summary Resume process
// @Description Resume a suspended process instance
// @Tags flow
// @Produce json
// @Param id path string true "Process ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{id}/resume [post]
// @Security Bearer
func (h *ProcessHandler) Resume(c *gin.Context) {
	processID := c.Param("id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.processService.Resume(c.Request.Context(), processID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}
