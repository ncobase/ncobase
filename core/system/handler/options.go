package handler

import (
	"ncobase/common/helper"
	"ncobase/common/resp"
	"ncobase/core/system/service"
	"ncobase/core/system/structs"

	"github.com/gin-gonic/gin"
)

// OptionsHandlerInterface represents the options handler interface.
type OptionsHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Initialize(c *gin.Context)
}

// optionsHandler represents the options handler.
type optionsHandler struct {
	s *service.Service
}

// NewOptionsHandler creates new options handler.
func NewOptionsHandler(svc *service.Service) OptionsHandlerInterface {
	return &optionsHandler{
		s: svc,
	}
}

// Create handles creating a new option.
//
// @Summary Create option
// @Description Create a new option.
// @Tags options
// @Accept json
// @Produce json
// @Param body body structs.OptionsBody true "OptionsBody object"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [post]
// @Security Bearer
func (h *optionsHandler) Create(c *gin.Context) {
	body := &structs.OptionsBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Options.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating an option.
//
// @Summary Update option
// @Description Update an existing option.
// @Tags options
// @Accept json
// @Produce json
// @Param body body structs.UpdateOptionsBody true "UpdateOptionsBody object"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [put]
// @Security Bearer
func (h *optionsHandler) Update(c *gin.Context) {
	body := &structs.UpdateOptionsBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Options.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving an option by ID or name.
//
// @Summary Get option
// @Description Retrieve an option by ID or name.
// @Tags options
// @Produce json
// @Param option path string true "Option ID or name"
// @Param params query structs.FindOptions true "FindOptions parameters"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/{option} [get]
// @Security Bearer
func (h *optionsHandler) Get(c *gin.Context) {
	params := &structs.FindOptions{Option: c.Param("option")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Options.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting an option by ID or name.
//
// @Summary Delete option
// @Description Delete an option by ID or name.
// @Tags options
// @Produce json
// @Param option path string true "Option ID or name"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/{option} [delete]
// @Security Bearer
func (h *optionsHandler) Delete(c *gin.Context) {
	params := &structs.FindOptions{Option: c.Param("option")}
	err := h.s.Options.Delete(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// List handles listing all options.
//
// @Summary List options
// @Description Retrieve a list of options.
// @Tags options
// @Produce json
// @Param params query structs.ListOptionsParams true "List options parameters"
// @Success 200 {array} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [get]
// @Security Bearer
func (h *optionsHandler) List(c *gin.Context) {
	params := &structs.ListOptionsParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Options.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Initialize initializes the system with default options
//
// @Summary Initialize
// @Description Initialize the system with default options
// @Tags options
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/initialize [post]
// @Security Bearer
func (h *optionsHandler) Initialize(c *gin.Context) {
	err := h.s.Options.Initialize(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}
