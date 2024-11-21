package handler

import (
	"ncobase/common/ecode"
	"ncobase/common/helper"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"

	"github.com/gin-gonic/gin"
)

// TaskHandlerInterface defines task handler interface
type TaskHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Complete(c *gin.Context)
	Delegate(c *gin.Context)
	Transfer(c *gin.Context)
	Withdraw(c *gin.Context)
	Urge(c *gin.Context)
	Claim(c *gin.Context)
}

// TaskHandler implements task operations
type TaskHandler struct {
	taskService service.TaskServiceInterface
}

// NewTaskHandler creates new task handler
func NewTaskHandler(svc *service.Service) TaskHandlerInterface {
	return &TaskHandler{
		taskService: svc.GetTask(),
	}
}

// Create handles task creation
// @Summary Create task
// @Description Create a new task
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.TaskBody true "Task creation body"
// @Success 200 {object} structs.ReadTask
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks [post]
// @Security Bearer
func (h *TaskHandler) Create(c *gin.Context) {
	body := &structs.TaskBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.taskService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting task details
// @Summary Get task
// @Description Get task details
// @Tags flow
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} structs.ReadTask
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id} [get]
// @Security Bearer
func (h *TaskHandler) Get(c *gin.Context) {
	params := &structs.FindTaskParams{
		ProcessID: c.Param("id"),
	}

	result, err := h.taskService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles task update
// @Summary Update task
// @Description Update task details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.UpdateTaskBody true "Task update body"
// @Success 200 {object} structs.ReadTask
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id} [put]
// @Security Bearer
func (h *TaskHandler) Update(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateTaskBody{ID: taskID}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.taskService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles task deletion
// @Summary Delete task
// @Description Delete a task
// @Tags flow
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id} [delete]
// @Security Bearer
func (h *TaskHandler) Delete(c *gin.Context) {
	params := &structs.FindTaskParams{
		ProcessID: c.Param("id"),
	}

	if err := h.taskService.Delete(c.Request.Context(), params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles task listing
// @Summary List tasks
// @Description List tasks with pagination
// @Tags flow
// @Produce json
// @Param params query structs.ListTaskParams true "Task list parameters"
// @Success 200 {array} structs.ReadTask
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks [get]
// @Security Bearer
func (h *TaskHandler) List(c *gin.Context) {
	params := &structs.ListTaskParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.taskService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Complete handles task completion
// @Summary Complete task
// @Description Complete a task
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.CompleteTaskRequest true "Task completion request"
// @Success 200 {object} structs.CompleteTaskResponse
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/complete [post]
// @Security Bearer
func (h *TaskHandler) Complete(c *gin.Context) {
	req := &structs.CompleteTaskRequest{
		TaskID: c.Param("id"),
	}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.taskService.Complete(c.Request.Context(), req)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delegate handles task delegation
// @Summary Delegate task
// @Description Delegate a task to another user
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.DelegateTaskRequest true "Task delegation request"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/delegate [post]
// @Security Bearer
func (h *TaskHandler) Delegate(c *gin.Context) {
	req := &structs.DelegateTaskRequest{
		TaskID: c.Param("id"),
	}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.taskService.Delegate(c.Request.Context(), req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Transfer handles task transfer
// @Summary Transfer task
// @Description Transfer a task to another user
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.TransferTaskRequest true "Task transfer request"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/transfer [post]
// @Security Bearer
func (h *TaskHandler) Transfer(c *gin.Context) {
	req := &structs.TransferTaskRequest{
		TaskID: c.Param("id"),
	}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.taskService.Transfer(c.Request.Context(), req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Withdraw handles task withdrawal
// @Summary Withdraw task
// @Description Withdraw a completed task
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.WithdrawTaskRequest true "Task withdrawal request"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/withdraw [post]
// @Security Bearer
func (h *TaskHandler) Withdraw(c *gin.Context) {
	req := &structs.WithdrawTaskRequest{
		TaskID: c.Param("id"),
	}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.taskService.Withdraw(c.Request.Context(), req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Urge handles task urging
// @Summary Urge task
// @Description Send urge request for a task
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param body body structs.UrgeTaskRequest true "Task urge request"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/urge [post]
// @Security Bearer
func (h *TaskHandler) Urge(c *gin.Context) {
	req := &structs.UrgeTaskRequest{
		TaskID: c.Param("id"),
	}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.taskService.Urge(c.Request.Context(), req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Claim handles task claiming
// @Summary Claim task
// @Description Claim an unassigned task
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param assignees body types.JSONArray true "Task assignees"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{id}/claim [post]
// @Security Bearer
func (h *TaskHandler) Claim(c *gin.Context) {
	taskID := c.Param("id")
	var assignees types.JSONArray
	if err := c.ShouldBindJSON(&assignees); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if err := h.taskService.Claim(c.Request.Context(), taskID, &assignees); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}
