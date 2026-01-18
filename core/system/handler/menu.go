package handler

import (
	"ncobase/core/system/service"
	"ncobase/core/system/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// MenuHandlerInterface represents the menu handler interface.
type MenuHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	GetBySlug(c *gin.Context)
	GetNavigationMenus(c *gin.Context)
	GetMenuTree(c *gin.Context)
	GetUserAuthorizedMenus(c *gin.Context)
	MoveMenu(c *gin.Context)
	ReorderMenus(c *gin.Context)
	ToggleMenuStatus(c *gin.Context)
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
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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

// GetBySlug handles retrieving a menu by slug.
//
// @Summary Get menu by slug
// @Description Retrieve a menu by its slug.
// @Tags sys
// @Produce json
// @Param slug path string true "Menu slug"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/slug/{slug} [get]
// @Security Bearer
func (h *menuHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.s.Menu.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetNavigationMenus handles retrieving the system navigation menu orgs.
//
// @Summary Get navigation menus
// @Description Retrieve the system navigation menu orgs organized by type.
// @Tags sys
// @Produce json
// @Param sort_by query string false "Sort by field (order, created_at, name)" default(order)
// @Success 200 {object} structs.NavigationMenus "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/navigation [get]
// @Security Bearer
func (h *menuHandler) GetNavigationMenus(c *gin.Context) {
	sortBy := c.Query("sort_by")
	if sortBy == "" {
		sortBy = structs.SortByOrder
	}

	// Validate sort parameter
	validSortOptions := []string{structs.SortByOrder, structs.SortByCreatedAt, structs.SortByName}
	isValidSort := false
	for _, validOption := range validSortOptions {
		if sortBy == validOption {
			isValidSort = true
			break
		}
	}

	if !isValidSort {
		resp.Fail(c.Writer, resp.BadRequest("Invalid sort_by parameter. Valid options: order, created_at, name"))
		return
	}

	result, err := h.s.Menu.GetNavigationMenus(c.Request.Context(), sortBy)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetMenuTree handles retrieving the menu tree.
//
// @Summary Get menu tree
// @Description Retrieve the menu tree structure.
// @Tags sys
// @Produce json
// @Param params query structs.FindMenu true "FindMenu parameters"
// @Success 200 {object} []structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/tree [get]
// @Security Bearer
func (h *menuHandler) GetMenuTree(c *gin.Context) {
	params := structs.FindMenu{}
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Always include children for tree structure
	params.Children = true

	result, err := h.s.Menu.GetMenuTree(c.Request.Context(), &params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetUserAuthorizedMenus handles retrieving menus that a user is authorized to access.
//
// @Summary Get user authorized menus
// @Description Retrieve menus that a user is authorized to access.
// @Tags sys
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {array} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/authorized/{user_id} [get]
// @Security Bearer
func (h *menuHandler) GetUserAuthorizedMenus(c *gin.Context) {
	userID := c.Param("user_id")

	result, err := h.s.Menu.GetUserAuthorizedMenus(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// MoveMenu handles moving a menu to a new parent and/or changing its order.
//
// @Summary Move menu
// @Description Move a menu to a new parent and/or change its order.
// @Tags sys
// @Accept json
// @Produce json
// @Param id path string true "Menu ID"
// @Param body body object true "MoveMenuParams object"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/{id}/move [put]
// @Security Bearer
func (h *menuHandler) MoveMenu(c *gin.Context) {
	id := c.Param("id")

	var params struct {
		ParentID string `json:"parent_id"`
		Order    int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Menu.MoveMenu(c.Request.Context(), id, params.ParentID, params.Order)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ReorderMenus handles reordering a set of menus.
//
// @Summary Reorder menus
// @Description Reorder a set of sibling menus.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body []string true "Array of menu IDs in desired order"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/reorder [post]
// @Security Bearer
func (h *menuHandler) ReorderMenus(c *gin.Context) {
	var menuIDs []string
	if err := c.ShouldBindJSON(&menuIDs); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	err := h.s.Menu.ReorderMenus(c.Request.Context(), menuIDs)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, nil)
}

// ToggleMenuStatus handles toggling menu status (enable/disable/show/hide).
//
// @Summary Toggle menu status
// @Description Toggle menu status with specified action.
// @Tags sys
// @Produce json
// @Param id path string true "Menu ID"
// @Param action path string true "Action (enable/disable/show/hide)"
// @Success 200 {object} structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus/{id}/{action} [put]
// @Security Bearer
func (h *menuHandler) ToggleMenuStatus(c *gin.Context) {
	id := c.Param("id")
	action := c.Param("action")

	// Validate action
	validActions := []string{"enable", "disable", "show", "hide"}
	isValidAction := false
	for _, validAction := range validActions {
		if action == validAction {
			isValidAction = true
			break
		}
	}

	if !isValidAction {
		resp.Fail(c.Writer, resp.BadRequest("Invalid action. Valid actions: enable, disable, show, hide"))
		return
	}

	result, err := h.s.Menu.ToggleStatus(c.Request.Context(), id, action)
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
// @Success 200 {object} []structs.ReadMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/menus [get]
// @Security Bearer
func (h *menuHandler) List(c *gin.Context) {
	params := &structs.ListMenuParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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
