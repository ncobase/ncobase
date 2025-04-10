package handler

import (
	"ncobase/core/workflow/service"
	"ncobase/core/workflow/structs"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/resp"

	"github.com/gin-gonic/gin"
)

// DelegationHandlerInterface defines delegation handler interface
type DelegationHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Enable(c *gin.Context)
	Disable(c *gin.Context)
	GetActiveDelegations(c *gin.Context)
	CheckDelegation(c *gin.Context)
}

// DelegationHandler implements delegation operations
type DelegationHandler struct {
	delegationService service.DelegationServiceInterface
}

// NewDelegationHandler creates new delegation handler
func NewDelegationHandler(svc *service.Service) DelegationHandlerInterface {
	return &DelegationHandler{
		delegationService: svc.GetDelegation(),
	}
}

// Create handles delegation creation
// @Summary Create delegation
// @Description Create a new delegation rule
// @Tags flow
// @Accept json
// @Produce json
// @Param body body structs.DelegationBody true "Delegation creation body"
// @Success 200 {object} structs.ReadDelegation
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations [post]
// @Security Bearer
func (h *DelegationHandler) Create(c *gin.Context) {
	body := &structs.DelegationBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.delegationService.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles getting delegation details
// @Summary Get delegation
// @Description Get delegation rule details
// @Tags flow
// @Produce json
// @Param id path string true "Delegation ID"
// @Success 200 {object} structs.ReadDelegation
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations/{id} [get]
// @Security Bearer
func (h *DelegationHandler) Get(c *gin.Context) {
	params := &structs.FindDelegationParams{
		DelegatorID: c.Param("id"),
	}

	result, err := h.delegationService.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles delegation update
// @Summary Update delegation
// @Description Update delegation rule details
// @Tags flow
// @Accept json
// @Produce json
// @Param id path string true "Delegation ID"
// @Param body body structs.UpdateDelegationBody true "Delegation update body"
// @Success 200 {object} structs.ReadDelegation
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations/{id} [put]
// @Security Bearer
func (h *DelegationHandler) Update(c *gin.Context) {
	delegationID := c.Param("id")
	if delegationID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	body := &structs.UpdateDelegationBody{ID: delegationID}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.delegationService.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles delegation deletion
// @Summary Delete delegation
// @Description Delete a delegation rule
// @Tags flow
// @Produce json
// @Param id path string true "Delegation ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations/{id} [delete]
// @Security Bearer
func (h *DelegationHandler) Delete(c *gin.Context) {
	delegationID := c.Param("id")
	if delegationID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.delegationService.Delete(c.Request.Context(), delegationID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles delegation listing
// @Summary List delegations
// @Description List delegation rules with pagination
// @Tags flow
// @Produce json
// @Param params query structs.ListDelegationParams true "Delegation list parameters"
// @Success 200 {array} structs.ReadDelegation
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations [get]
// @Security Bearer
func (h *DelegationHandler) List(c *gin.Context) {
	params := &structs.ListDelegationParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.delegationService.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Enable handles delegation enabling
// @Summary Enable delegation
// @Description Enable a delegation rule
// @Tags flow
// @Produce json
// @Param id path string true "Delegation ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations/{id}/enable [post]
// @Security Bearer
func (h *DelegationHandler) Enable(c *gin.Context) {
	delegationID := c.Param("id")
	if delegationID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.delegationService.EnableDelegation(c.Request.Context(), delegationID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// Disable handles delegation disabling
// @Summary Disable delegation
// @Description Disable a delegation rule
// @Tags flow
// @Produce json
// @Param id path string true "Delegation ID"
// @Success 200 {object} resp.Exception
// @Failure 400 {object} resp.Exception
// @Router /flow/delegations/{id}/disable [post]
// @Security Bearer
func (h *DelegationHandler) Disable(c *gin.Context) {
	delegationID := c.Param("id")
	if delegationID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	if err := h.delegationService.DisableDelegation(c.Request.Context(), delegationID); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// GetActiveDelegations handles getting active delegations
// @Summary Get active delegations
// @Description Get active delegations for a user
// @Tags flow
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} structs.ReadDelegation
// @Failure 400 {object} resp.Exception
// @Router /flow/users/{user_id}/delegations [get]
// @Security Bearer
func (h *DelegationHandler) GetActiveDelegations(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("user_id")))
		return
	}

	result, err := h.delegationService.GetActiveDelegations(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// CheckDelegation handles delegation checking
// @Summary Check delegation
// @Description Check if task can be delegated to user
// @Tags flow
// @Produce json
// @Param task_id path string true "Task ID"
// @Param assignee_id query string true "Assignee ID"
// @Success 200 {string} string "Delegatee ID if delegation exists"
// @Failure 400 {object} resp.Exception
// @Router /flow/tasks/{task_id}/check-delegation [get]
// @Security Bearer
func (h *DelegationHandler) CheckDelegation(c *gin.Context) {
	taskID := c.Param("task_id")
	assigneeID := c.Query("assignee_id")

	if taskID == "" || assigneeID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Missing required parameters"))
		return
	}

	result, err := h.delegationService.CheckDelegation(c.Request.Context(), taskID, assigneeID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
