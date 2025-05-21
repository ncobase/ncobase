package handler

import (
	"ncobase/system/service"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// OptionsHandlerInterface represents the options handler interface.
type OptionsHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	GetByName(c *gin.Context)
	GetByType(c *gin.Context)
	BatchGetByNames(c *gin.Context)
	Delete(c *gin.Context)
	DeleteByPrefix(c *gin.Context)
	List(c *gin.Context)
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
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.OptionsBody true "OptionsBody object"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [post]
// @Security Bearer
func (h *optionsHandler) Create(c *gin.Context) {
	body := &structs.OptionsBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.UpdateOptionsBody true "UpdateOptionsBody object"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [put]
// @Security Bearer
func (h *optionsHandler) Update(c *gin.Context) {
	body := &structs.UpdateOptionsBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
// @Tags sys
// @Produce json
// @Param option path string true "Option ID or name"
// @Param params query structs.FindOptions true "FindOptions parameters"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/{option} [get]
// @Security Bearer
func (h *optionsHandler) Get(c *gin.Context) {
	params := &structs.FindOptions{Option: c.Param("option")}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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

// GetByName handles retrieving an option by name.
//
// @Summary Get option by name
// @Description Retrieve an option by its name.
// @Tags sys
// @Produce json
// @Param name path string true "Option name"
// @Success 200 {object} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/name/{name} [get]
// @Security Bearer
func (h *optionsHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	result, err := h.s.Options.GetByName(c.Request.Context(), name)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetByType handles retrieving options by type.
//
// @Summary Get options by type
// @Description Retrieve options by their type.
// @Tags sys
// @Produce json
// @Param type path string true "Option type"
// @Success 200 {array} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/type/{type} [get]
// @Security Bearer
func (h *optionsHandler) GetByType(c *gin.Context) {
	typeName := c.Param("type")

	result, err := h.s.Options.GetByType(c.Request.Context(), typeName)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// BatchGetByNames handles retrieving multiple options by their names.
//
// @Summary Batch get options
// @Description Retrieve multiple options by their names.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body []string true "Array of option names"
// @Success 200 {object} map[string]structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/batch [post]
// @Security Bearer
func (h *optionsHandler) BatchGetByNames(c *gin.Context) {
	var names []string
	if err := c.ShouldBindJSON(&names); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Options.BatchGetByNames(c.Request.Context(), names)
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
// @Tags sys
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

// DeleteByPrefix handles deleting options by prefix.
//
// @Summary Delete options by prefix
// @Description Delete options matching a prefix pattern.
// @Tags sys
// @Accept json
// @Produce json
// @Param prefix query string true "Name prefix pattern"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options/prefix [delete]
// @Security Bearer
func (h *optionsHandler) DeleteByPrefix(c *gin.Context) {
	prefix := c.Query("prefix")
	if prefix == "" {
		resp.Fail(c.Writer, resp.BadRequest("prefix is required"))
		return
	}

	err := h.s.Options.DeleteByPrefix(c.Request.Context(), prefix)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, nil)
}

// List handles listing all options.
//
// @Summary List options
// @Description Retrieve a list of options.
// @Tags sys
// @Produce json
// @Param params query structs.ListOptionsParams true "List options parameters"
// @Success 200 {array} structs.ReadOptions "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/options [get]
// @Security Bearer
func (h *optionsHandler) List(c *gin.Context) {
	params := &structs.ListOptionsParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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
