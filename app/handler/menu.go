package handler

import (
	"ncobase/app/data/structs"
	"ncobase/common/resp"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// CreateMenuHandler handles creating a new menu.
//
// @Summary Create menu
// @Description Create a new menu.
// @Tags menu
// @Accept json
// @Produce json
// @Param body body structs.MenuBody true "MenuBody object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus [post]
// @Security Bearer
func (h *Handler) CreateMenuHandler(c *gin.Context) {
	body := &structs.MenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateMenuService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UpdateMenuHandler handles updating a menu.
//
// @Summary Update menu
// @Description Update an existing menu.
// @Tags menu
// @Accept json
// @Produce json
// @Param body body structs.UpdateMenuBody true "UpdateMenuBody object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus [put]
// @Security Bearer
func (h *Handler) UpdateMenuHandler(c *gin.Context) {
	body := &structs.UpdateMenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.UpdateMenuService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMenuHandler handles retrieving a menu by ID or slug.
//
// @Summary Get menu
// @Description Retrieve a menu by ID or slug.
// @Tags menu
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Param params query structs.FindMenu true "FindMenu parameters"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus/{slug} [get]
// @Security Bearer
func (h *Handler) GetMenuHandler(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.GetMenuService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DeleteMenuHandler handles deleting a menu by ID or slug.
//
// @Summary Delete menu
// @Description Delete a menu by ID or slug.
// @Tags menu
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteMenuHandler(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}
	result, err := h.svc.DeleteMenuService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ListMenusHandler handles listing all menus.
//
// @Summary List menus
// @Description Retrieve a list or tree structure of menus.
// @Tags menu
// @Produce json
// @Param params query structs.ListMenuParams true "List menu parameters"
// @Success 200 {array} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus [get]
// @Security Bearer
func (h *Handler) ListMenusHandler(c *gin.Context) {
	params := &structs.ListMenuParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.ListMenusService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// // GetMenuTreeHandler handles retrieving the menu tree.
// //
// // @Summary Get menu tree
// // @Description Retrieve the menu tree structure.
// // @Tags menu
// // @Produce json
// // @Param params query structs.FindMenu true "FindMenu parameters"
// // @Success 200 {object} structs.ReadMenu "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /v1/menus/tree [get]
// // @Security Bearer
// func (h *Handler) GetMenuTreeHandler(c *gin.Context) {
// 	params := &structs.FindMenu{}
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, err := h.svc.GetMenuTreeService(c, params)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
