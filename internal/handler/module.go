package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/types"

	"github.com/gin-gonic/gin"
)

// CreateModuleHandler handles the creation of a module.
//
// @Summary Create a new module
// @Description Create a new module with the provided data
// @Tags modules
// @Accept json
// @Produce json
// @Param body body structs.CreateModuleBody true "Module data"
// @Success 200 {object} structs.ReadModule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/modules [post]
// @Security Bearer
func (h *Handler) CreateModuleHandler(c *gin.Context) {
	var body *structs.CreateModuleBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.CreateModuleService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateModuleHandler handles updating a module (full and partial).
//
// @Summary Update an existing module
// @Description Update an existing module with the provided data
// @Tags modules
// @Accept json
// @Produce json
// @Param slug path string true "Module slug or ID"
// @Param body body structs.UpdateModuleBody true "Module data"
// @Success 200 {object} structs.ReadModule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/modules/{slug} [put]
// @Security Bearer
func (h *Handler) UpdateModuleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	var updates types.JSON
	if err := c.ShouldBind(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateModuleService(c, slug, updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetModuleHandler handles getting a module.
//
// @Summary Get a module by slug or ID
// @Description Retrieve a module by its slug or ID
// @Tags modules
// @Produce json
// @Param slug path string true "Module slug or ID"
// @Success 200 {object} structs.ReadModule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/modules/{slug} [get]
// @Security Bearer
func (h *Handler) GetModuleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.GetModuleService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteModuleHandler handles deleting a module.
//
// @Summary Delete a module by slug or ID
// @Description Delete a module by its slug or ID
// @Tags modules
// @Produce json
// @Param slug path string true "Module slug or ID"
// @Success 200 {object} structs.ReadModule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/modules/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteModuleHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.DeleteModuleService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListModuleHandler handles listing modules.
//
// @Summary List all modules
// @Description Retrieve a list of modules based on the provided query parameters
// @Tags modules
// @Produce json
// @Param params query structs.ListModuleParams true "List modules parameters"
// @Success 200 {array} structs.ReadModule "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/modules [get]
// @Security Bearer
func (h *Handler) ListModuleHandler(c *gin.Context) {
	params := &structs.ListModuleParams{}
	if err := c.ShouldBind(params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	validationErrors := structs.Validate(params)
	if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	modules, err := h.svc.ListModulesService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, modules)
}
