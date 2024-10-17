package handler

import (
	"ncobase/common/helper"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/feature/group/service"
	"ncobase/feature/group/structs"

	"github.com/gin-gonic/gin"
)

// GroupHandlerInterface represents the group handler interface.
type GroupHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// groupHandler represents the group handler.
type groupHandler struct {
	s *service.Service
}

// NewGroupHandler creates new group handler.
func NewGroupHandler(svc *service.Service) GroupHandlerInterface {
	return &groupHandler{
		s: svc,
	}
}

// Create handles creating a new group.
//
// @Summary Create group
// @Description Create a new group.
// @Tags group
// @Accept json
// @Produce json
// @Param body body structs.GroupBody true "GroupBody object"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /group/groups [post]
// @Security Bearer
func (h *groupHandler) Create(c *gin.Context) {
	body := &structs.CreateGroupBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a group.
//
// @Summary Update group
// @Description Update an existing group.
// @Tags group
// @Accept json
// @Produce json
// @Param body body structs.UpdateGroupBody true "UpdateGroupBody object"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /group/groups [put]
// @Security Bearer
func (h *groupHandler) Update(c *gin.Context) {
	body := types.JSON{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Update(c.Request.Context(), body["id"].(string), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a group by ID or slug.
//
// @Summary Get group
// @Description Retrieve a group by ID or slug.
// @Tags group
// @Produce json
// @Param slug path string true "Group ID or slug"
// @Param params query structs.FindGroup true "FindGroup parameters"
// @Success 200 {object} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /group/groups/{slug} [get]
// @Security Bearer
func (h *groupHandler) Get(c *gin.Context) {
	params := &structs.FindGroup{Group: c.Param("slug")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a group by ID or slug.
//
// @Summary Delete group
// @Description Delete a group by ID or slug.
// @Tags group
// @Produce json
// @Param slug path string true "Group ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /group/groups/{slug} [delete]
// @Security Bearer
func (h *groupHandler) Delete(c *gin.Context) {
	params := &structs.FindGroup{Group: c.Param("slug")}
	err := h.s.Group.Delete(c.Request.Context(), params.Group)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// List handles listing all groups.
//
// @Summary List groups
// @Description Retrieve a list or tree structure of groups.
// @Tags group
// @Produce json
// @Param params query structs.ListGroupParams true "List group parameters"
// @Success 200 {array} structs.ReadGroup "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /group/groups [get]
// @Security Bearer
func (h *groupHandler) List(c *gin.Context) {
	params := &structs.ListGroupParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Group.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// // GetTree handles retrieving the group tree.
// //
// // @Summary Get group tree
// // @Description Retrieve the group tree structure.
// // @Tags group
// // @Produce json
// // @Param params query structs.FindGroup true "FindGroup parameters"
// // @Success 200 {object} structs.ReadGroup "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /group/groups/tree [get]
// // @Security Bearer
// func (h *Handler) GetTree(c *gin.Context) {
// 	params := &structs.FindGroup{}
// 	if validationErrors, err := helper.ShouldBindAndValidateStruct(c,params); err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	} else if len(validationErrors) > 0 {
// 		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
// 		return
// 	}
//
// 	result, err := h.s.Group.GetTree(c.Request.Context(),params)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }
