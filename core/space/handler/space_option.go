package handler

import (
	"ncobase/space/service"
	"ncobase/space/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// SpaceOptionHandlerInterface represents the space option handler interface.
type SpaceOptionHandlerInterface interface {
	AddOptionsToSpace(c *gin.Context)
	RemoveOptionsFromSpace(c *gin.Context)
	GetSpaceOption(c *gin.Context)
	CheckOptionsInSpace(c *gin.Context)
}

// spaceOptionHandler represents the space option handler.
type spaceOptionHandler struct {
	s *service.Service
}

// NewSpaceOptionHandler creates new space option handler.
func NewSpaceOptionHandler(svc *service.Service) SpaceOptionHandlerInterface {
	return &spaceOptionHandler{
		s: svc,
	}
}

// AddOptionsToSpace handles adding options to a space.
//
// @Summary Add options to space
// @Description Add options to a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.AddOptionsToSpaceRequest true "AddOptionsToSpaceRequest object"
// @Success 200 {object} structs.SpaceOption "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/options [post]
// @Security Bearer
func (h *spaceOptionHandler) AddOptionsToSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.AddOptionsToSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if options already in space
	exists, _ := h.s.SpaceOption.IsOptionsInSpace(c.Request.Context(), spaceID, req.OptionID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Options already belong to this space"))
		return
	}

	result, err := h.s.SpaceOption.AddOptionsToSpace(c.Request.Context(), spaceID, req.OptionID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveOptionsFromSpace handles removing options from a space.
//
// @Summary Remove options from space
// @Description Remove options from a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param optionsId path string true "Options ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/options/{optionsId} [delete]
// @Security Bearer
func (h *spaceOptionHandler) RemoveOptionsFromSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	optionsID := c.Param("optionsId")
	if optionsID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Options ID is required"))
		return
	}

	err := h.s.SpaceOption.RemoveOptionsFromSpace(c.Request.Context(), spaceID, optionsID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"status":    "removed",
		"space_id":  spaceID,
		"option_id": optionsID,
	})
}

// GetSpaceOption handles getting all options for a space.
//
// @Summary Get space option
// @Description Get all options for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {array} string "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/options [get]
// @Security Bearer
func (h *spaceOptionHandler) GetSpaceOption(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	optionsIDs, err := h.s.SpaceOption.GetSpaceOption(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"space_id":   spaceID,
		"option_ids": optionsIDs,
		"count":      len(optionsIDs),
	})
}

// CheckOptionsInSpace handles checking if options belong to a space.
//
// @Summary Check options in space
// @Description Check if options belong to a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param optionsId path string true "Options ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/options/{optionsId}/check [get]
// @Security Bearer
func (h *spaceOptionHandler) CheckOptionsInSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	optionsID := c.Param("optionsId")
	if optionsID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Options ID is required"))
		return
	}

	exists, err := h.s.SpaceOption.IsOptionsInSpace(c.Request.Context(), spaceID, optionsID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
