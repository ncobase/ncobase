package handler

import (
	"ncobase/core/space/service"
	"ncobase/core/space/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// SpaceDictionaryHandlerInterface represents the space dictionary handler interface.
type SpaceDictionaryHandlerInterface interface {
	AddDictionaryToSpace(c *gin.Context)
	RemoveDictionaryFromSpace(c *gin.Context)
	GetSpaceDictionaries(c *gin.Context)
	CheckDictionaryInSpace(c *gin.Context)
}

// spaceDictionaryHandler represents the space dictionary handler.
type spaceDictionaryHandler struct {
	s *service.Service
}

// NewSpaceDictionaryHandler creates new space dictionary handler.
func NewSpaceDictionaryHandler(svc *service.Service) SpaceDictionaryHandlerInterface {
	return &spaceDictionaryHandler{
		s: svc,
	}
}

// AddDictionaryToSpace handles adding a dictionary to a space.
//
// @Summary Add dictionary to space
// @Description Add a dictionary to a space
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.AddDictionaryToSpaceRequest true "AddDictionaryToSpaceRequest object"
// @Success 200 {object} structs.SpaceDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/dictionaries [post]
// @Security Bearer
func (h *spaceDictionaryHandler) AddDictionaryToSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	var req structs.AddDictionaryToSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Check if dictionary already in space
	exists, _ := h.s.SpaceDictionary.IsDictionaryInSpace(c.Request.Context(), spaceID, req.DictionaryID)
	if exists {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary already belongs to this space"))
		return
	}

	result, err := h.s.SpaceDictionary.AddDictionaryToSpace(c.Request.Context(), spaceID, req.DictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RemoveDictionaryFromSpace handles removing a dictionary from a space.
//
// @Summary Remove dictionary from space
// @Description Remove a dictionary from a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param dictionaryId path string true "Dictionary ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/dictionaries/{dictionaryId} [delete]
// @Security Bearer
func (h *spaceDictionaryHandler) RemoveDictionaryFromSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	dictionaryID := c.Param("dictionaryId")
	if dictionaryID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary ID is required"))
		return
	}

	err := h.s.SpaceDictionary.RemoveDictionaryFromSpace(c.Request.Context(), spaceID, dictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"status":        "removed",
		"space_id":      spaceID,
		"dictionary_id": dictionaryID,
	})
}

// GetSpaceDictionaries handles getting all dictionaries for a space.
//
// @Summary Get space dictionaries
// @Description Get all dictionaries for a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/dictionaries [get]
// @Security Bearer
func (h *spaceDictionaryHandler) GetSpaceDictionaries(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	dictionaryIDs, err := h.s.SpaceDictionary.GetSpaceDictionaries(c.Request.Context(), spaceID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]any{
		"space_id":       spaceID,
		"dictionary_ids": dictionaryIDs,
		"count":          len(dictionaryIDs),
	})
}

// CheckDictionaryInSpace handles checking if a dictionary belongs to a space.
//
// @Summary Check dictionary in space
// @Description Check if a dictionary belongs to a space
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param dictionaryId path string true "Dictionary ID"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/dictionaries/{dictionaryId}/check [get]
// @Security Bearer
func (h *spaceDictionaryHandler) CheckDictionaryInSpace(c *gin.Context) {
	spaceID := c.Param("spaceId")
	if spaceID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Space ID is required"))
		return
	}

	dictionaryID := c.Param("dictionaryId")
	if dictionaryID == "" {
		resp.Fail(c.Writer, resp.BadRequest("Dictionary ID is required"))
		return
	}

	exists, err := h.s.SpaceDictionary.IsDictionaryInSpace(c.Request.Context(), spaceID, dictionaryID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"exists": exists})
}
