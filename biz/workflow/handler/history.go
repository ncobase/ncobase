package handler

import (
	"ncobase/workflow/service"
	"ncobase/workflow/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// HistoryHandlerInterface defines history handler interface
type HistoryHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	GetProcessHistory(c *gin.Context)
	GetTaskHistory(c *gin.Context)
	GetOperatorHistory(c *gin.Context)
}

// HistoryHandler implements history operations
type HistoryHandler struct {
	historyService service.HistoryServiceInterface
}

// NewHistoryHandler creates new history handler
func NewHistoryHandler(svc *service.Service) HistoryHandlerInterface {
	return &HistoryHandler{
		historyService: svc.GetHistory(),
	}
}

// Create handles history creation
// @Summary Create history
// @Description Create a new history record
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.HistoryBody true "History creation body"
// @Success 200 {object} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/histories [post]
// @Security Bearer
func (h *HistoryHandler) Create(c *gin.Context) {
	body := &structs.HistoryBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.historyService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting history details
// @Summary Get history
// @Description Get history record details
// @Tags flow
// @Produce json
// @Param id path string true "History ID"
// @Success 200 {object} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/histories/{id} [get]
// @Security Bearer
func (h *HistoryHandler) Get(c *gin.Context) {
	params := &structs.FindHistoryParams{
		ProcessID: c.Param("id"),
	}

	result, err := h.historyService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// List handles history listing
// @Summary List histories
// @Description List history records
// @Tags flow
// @Produce json
// @Param params query structs.ListHistoryParams true "History list parameters"
// @Success 200 {array} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/histories [get]
// @Security Bearer
func (h *HistoryHandler) List(c *gin.Context) {
	params := &structs.ListHistoryParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.historyService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetProcessHistory handles getting process history
// @Summary Get process history
// @Description Get complete history of a process instance
// @Tags flow
// @Produce json
// @Param process_id path string true "Process ID"
// @Success 200 {array} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/processes/{process_id}/histories [get]
// @Security Bearer
func (h *HistoryHandler) GetProcessHistory(c *gin.Context) {
	processID := c.Param("process_id")
	if processID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing process ID"))
		return
	}

	result, err := h.historyService.GetProcessHistory(c.Request.Context(), processID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetTaskHistory handles getting task history
// @Summary Get task history
// @Description Get complete history of a task
// @Tags flow
// @Produce json
// @Param task_id path string true "Task ID"
// @Success 200 {array} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{task_id}/histories [get]
// @Security Bearer
func (h *HistoryHandler) GetTaskHistory(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing task ID"))
		return
	}

	result, err := h.historyService.GetTaskHistory(c.Request.Context(), taskID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetOperatorHistory handles getting operator history
// @Summary Get operator history
// @Description Get complete history of operations by an operator
// @Tags flow
// @Produce json
// @Param operator path string true "Operator ID"
// @Success 200 {array} structs.ReadHistory
// @Failure 400 {object} resp.Exception
// @Router /flow/operators/{operator}/histories [get]
// @Security Bearer
func (h *HistoryHandler) GetOperatorHistory(c *gin.Context) {
	operator := c.Param("operator")
	if operator == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing operator ID"))
		return
	}

	result, err := h.historyService.GetOperatorHistory(c.Request.Context(), operator)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
