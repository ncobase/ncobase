package handler

import (
	"github.com/ncobase/ncore/pkg/helper"
	"github.com/ncobase/ncore/pkg/resp"
	"ncobase/core/system/service"
	"ncobase/core/system/structs"

	"github.com/gin-gonic/gin"
)

// MenuHandlerInterface represents the menu handler interface.
type MenuHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// menuHandler represents the menu handler.
type menuHandler struct {
	s *service.Service
}

// NewMenuHandler creates new menu handler.
func NewMenuHandler(svc *service.Service) MenuHandlerInterface {
	return &menuHandler{
		s: svc,
	}
}

// Create handles creating a new menu.
//
// @Summary Create menu
// @Description Create a new menu.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.MenuBody true "MenuBody object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus [post]
// @Security Bearer
func (h *menuHandler) Create(c *gin.Context) {
	body := &structs.MenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a menu.
//
// @Summary Update menu
// @Description Update an existing menu.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.UpdateMenuBody true "UpdateMenuBody object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus [put]
// @Security Bearer
func (h *menuHandler) Update(c *gin.Context) {
	body := &structs.UpdateMenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a menu by ID or slug.
//
// @Summary Get menu
// @Description Retrieve a menu by ID or slug.
// @Tags sys
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Param params query structs.FindMenu true "FindMenu parameters"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/{slug} [get]
// @Security Bearer
func (h *menuHandler) Get(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a menu by ID or slug.
//
// @Summary Delete menu
// @Description Delete a menu by ID or slug.
// @Tags sys
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/{slug} [delete]
// @Security Bearer
func (h *menuHandler) Delete(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}
	result, err := h.s.Menu.Delete(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// List handles listing all menus.
//
// @Summary List menus
// @Description Retrieve a list or tree structure of menus.
// @Tags sys
// @Produce json
// @Param params query structs.ListMenuParams true "List menu parameters"
// @Success 200 {array} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus [get]
// @Security Bearer
func (h *menuHandler) List(c *gin.Context) {
	params := &structs.ListMenuParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// // GetTree handles retrieving the menu tree.
// //
// // @Summary Get menu tree
// // @Description Retrieve the menu tree structure.
// // @Tags sys
// // @Produce json
// // @Param params query structs.FindMenu true "FindMenu parameters"
// // @Success 200 {object} structs.ReadMenu "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /sys/menus/tree [get]
// // @Security Bearer
// func (h *Handler) GetTree(c *gin.Context) {
// 	params := &structs.FindMenu{}
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c,params); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, err := h.s.Menu.GetTree(c.Request.Context(),params)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
