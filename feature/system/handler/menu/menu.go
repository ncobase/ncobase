package menu

import (
	"ncobase/common/resp"
	"ncobase/feature/system/service"
	"ncobase/feature/system/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

type HandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}
type Handler struct {
	s *service.Service
}

func New(svc *service.Service) HandlerInterface {
	return &Handler{
		s: svc,
	}
}

// Create handles creating a new menu.
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
func (h *Handler) Create(c *gin.Context) {
	body := &structs.MenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.CreateMenuService(c.Request.Context(), body)
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
// @Tags menu
// @Accept json
// @Produce json
// @Param body body structs.UpdateMenuBody true "UpdateMenuBody object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus [put]
// @Security Bearer
func (h *Handler) Update(c *gin.Context) {
	body := &structs.UpdateMenuBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.UpdateMenuService(c.Request.Context(), body)
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
// @Tags menu
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Param params query structs.FindMenu true "FindMenu parameters"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus/{slug} [get]
// @Security Bearer
func (h *Handler) Get(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.GetMenuService(c.Request.Context(), params)
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
// @Tags menu
// @Produce json
// @Param slug path string true "Menu ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus/{slug} [delete]
// @Security Bearer
func (h *Handler) Delete(c *gin.Context) {
	params := &structs.FindMenu{Menu: c.Param("slug")}
	result, err := h.s.Menu.DeleteMenuService(c.Request.Context(), params)
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
// @Tags menu
// @Produce json
// @Param params query structs.ListMenuParams true "List menu parameters"
// @Success 200 {array} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/menus [get]
// @Security Bearer
func (h *Handler) List(c *gin.Context) {
	params := &structs.ListMenuParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Menu.ListMenusService(c.Request.Context(), params)
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
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c,params); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, err := h.s.Menu.GetMenuTreeService(c.Request.Context(),params)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
