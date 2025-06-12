package handler

import (
	"ncobase/space/service"
	"ncobase/space/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// SpaceHandlerInterface is the interface for the handler.
type SpaceHandlerInterface interface {
	Create(c *gin.Context)
	UserOwn(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	GetMenus(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	ListAttachments(c *gin.Context)
	ListRoles(c *gin.Context)
	GetSetting(c *gin.Context)
	ListUsers(c *gin.Context)
	ListOrganizations(c *gin.Context)
}

// SpaceHandler represents the handler.
type SpaceHandler struct {
	s *service.Service
}

// NewSpaceHandler creates a new handler.
func NewSpaceHandler(svc *service.Service) SpaceHandlerInterface {
	return &SpaceHandler{
		s: svc,
	}
}

// Create handles creating a space.
//
// @Summary Create space
// @Description Create a new space.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.CreateSpaceBody true "CreateSpaceBody object"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces [post]
// @Security Bearer
func (h *SpaceHandler) Create(c *gin.Context) {
	body := &structs.CreateSpaceBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// UserOwn handles reading a user's space.
//
// @Summary Get user owned space
// @Description Retrieve the space associated with the specified user.
// @Tags sys
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/users/{username}/space [get]
// @Security Bearer
func (h *SpaceHandler) UserOwn(c *gin.Context) {
	result, err := h.s.Space.UserOwn(c.Request.Context(), c.Param("username"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a space.
//
// @Summary Update space
// @Description Update the space information.
// @Tags sys
// @Accept json
// @Produce json
// @Param spaceId path string true "Space ID"
// @Param body body structs.UpdateSpaceBody true "UpdateSpaceBody object"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [put]
// @Security Bearer
func (h *SpaceHandler) Update(c *gin.Context) {
	slug := c.Param("spaceId")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug")))
		return
	}
	body := &structs.UpdateSpaceBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles reading space information.
//
// @Summary Get space
// @Description Retrieve information about a specific space.
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} structs.ReadSpace "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [get]
// @Security Bearer
func (h *SpaceHandler) Get(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetMenus handles reading space menu.
//
// @Summary Get space menu
// @Description Retrieve the menu associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/menu [get]
// @Security Bearer
func (h *SpaceHandler) GetMenus(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// GetSpaceSetting handles reading space setting.
//
// @Summary Get space setting
// @Description Retrieve the settings associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/setting [get]
// @Security Bearer
func (h *SpaceHandler) GetSpaceSetting(c *gin.Context) {
	result, err := h.s.Space.Get(c.Request.Context(), c.Param("spaceId"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a space.
//
// @Summary Delete space
// @Description Delete a specific space.
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId} [delete]
// @Security Bearer
func (h *SpaceHandler) Delete(c *gin.Context) {
	if err := h.s.Space.Delete(c.Request.Context(), c.Param("spaceId")); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing spaces.
//
// @Summary List spaces
// @Description Retrieve a list of spaces.
// @Tags sys
// @Produce json
// @Param params query structs.ListSpaceParams true "List space parameters"
// @Success 200 {array} structs.ReadSpace"success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces [get]
// @Security Bearer
func (h *SpaceHandler) List(c *gin.Context) {
	params := &structs.ListSpaceParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Space.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListAttachments handles listing space attachments.
// TODO: implement this
// @Summary List space attachments
// @Description Retrieve a list of attachments associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/attachments [get]
// @Security Bearer
func (h *SpaceHandler) ListAttachments(c *gin.Context) {
	// result, err := h.svc.ListAttachmentss(c.Request.Context(),c.Param("spaceId"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListRoles handles listing space roles.
// TODO: implement this
// @Summary List space roles
// @Description Retrieve a list of roles associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/roles [get]
// @Security Bearer
func (h *SpaceHandler) ListRoles(c *gin.Context) {
	// result, err := h.svc.ListRoless(c.Request.Context(),c.Param("spaceId"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// GetSetting handles listing space settings.
// TODO: implement this
// @Summary List space settings
// @Description Retrieve a list of settings associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/settings [get]
// @Security Bearer
func (h *SpaceHandler) GetSetting(c *gin.Context) {
	// result, err := h.svc.GetSettings(c.Request.Context(),c.Param("spaceId"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListUsers handles listing space users.
// TODO: implement this
// @Summary List space users
// @Description Retrieve a list of users associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/users [get]
// @Security Bearer
func (h *SpaceHandler) ListUsers(c *gin.Context) {
	// result, err := h.svc.ListUserss(c.Request.Context(),c.Param("spaceId"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}

// ListOrganizations handles listing space orgs.
// TODO: implement this
// @Summary List space orgs
// @Description Retrieve a list of orgs associated
// @Tags sys
// @Produce json
// @Param spaceId path string true "Space ID"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/spaces/{spaceId}/orgs [get]
// @Security Bearer
func (h *SpaceHandler) ListOrganizations(c *gin.Context) {
	// result, err := h.svc.ListOrganizationss(c.Request.Context(),c.Param("spaceId"))
	// if err != nil {
	// 	resp.Fail(c.Writer, resp.BadRequest(err.Error()))
	// 	return
	// }
	resp.Success(c.Writer)
}
