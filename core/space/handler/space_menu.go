package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// SpaceMenuHandlerInterface represents the space menu handler interface.
type SpaceMenuHandlerInterface interface {
	AddMenuToSpace(c *gin.Context)
	RemoveMenuFromSpace(c *gin.Context)
	GetSpaceMenus(c *gin.Context)
	CheckMenuInSpace(c *gin.Context)
}

// spaceMenuHandler represents the space menu handler.
type spaceMenuHandler struct {
	s *service.Service
}

// NewSpaceMenuHandler creates new space menu handler.
func NewSpaceMenuHandler(svc *service.Service) SpaceMenuHandlerInterface {
	return &spaceMenuHandler{
		s: svc,
	}
}

// AddMenuToSpace handles adding a menu to a space.
//
// @Summary Add menu to space
// @Description Add a menu to a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.AddMenuToSpaceRequest true "AddMenuToSpaceRequest object"
// @Success 200 {object} structs.SpaceMenu "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menus [post]
// @Security Bearer
func (h *spaceMenuHandler) AddMenuToSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.AddMenuToSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if menu already in space
	exists, _ := h.s.SpaceMenu.IsMenuInSpace(c.Request.Context(), spaceID, req.MenuID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Menu already belongs to this space"))
		return
	}

	result, err := h.s.SpaceMenu.AddMenuToSpace(c.Request.Context(), spaceID, req.MenuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveMenuFromSpace handles removing a menu from a space.
//
// @Summary Remove menu from space
// @Description Remove a menu from a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param menuId path string true "Menu ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menus/{menuId} [delete]
// @Security Bearer
func (h *spaceMenuHandler) RemoveMenuFromSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	menuID := c.Param("menuId")
	if menuID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Menu ID is required"))
		return
	}

	err := h.s.SpaceMenu.RemoveMenuFromSpace(c.Request.Context(), spaceID, menuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"status":   "removed",
		"space_id": spaceID,
		"menu_id":  menuID,
	})
}

// GetSpaceMenus handles getting all menus for a space.
//
// @Summary Get space menus
// @Description Get all menus for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menus [get]
// @Security Bearer
func (h *spaceMenuHandler) GetSpaceMenus(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	menuIDs, err := h.s.SpaceMenu.GetSpaceMenus(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"space_id": spaceID,
		"menu_ids": menuIDs,
		"count":    len(menuIDs),
	})
}

// CheckMenuInSpace handles checking if a menu belongs to a space.
//
// @Summary Check menu in space
// @Description Check if a menu belongs to a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param menuId path string true "Menu ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menus/{menuId}/check [get]
// @Security Bearer
func (h *spaceMenuHandler) CheckMenuInSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	menuID := c.Param("menuId")
	if menuID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Menu ID is required"))
		return
	}

	exists, err := h.s.SpaceMenu.IsMenuInSpace(c.Request.Context(), spaceID, menuID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
