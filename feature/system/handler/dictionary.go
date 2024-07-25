package handler

import (
	"ncobase/common/resp"
	"ncobase/feature/system/service"
	"ncobase/feature/system/structs"
	"ncobase/helper"

	"github.com/gin-gonic/gin"
)

// DictionaryHandlerInterface represents the dictionary handler interface.
type DictionaryHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// dictionaryHandler represents the dictionary handler.
type dictionaryHandler struct {
	s *service.Service
}

// NewDictionaryHandler creates new dictionary handler.
func NewDictionaryHandler(svc *service.Service) DictionaryHandlerInterface {
	return &dictionaryHandler{
		s: svc,
	}
}

// Create handles creating a new dictionary.
//
// @Summary Create dictionary
// @Description Create a new dictionary.
// @Tags dictionary
// @Accept json
// @Produce json
// @Param body body structs.DictionaryBody true "DictionaryBody object"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dictionarys [post]
// @Security Bearer
func (h *dictionaryHandler) Create(c *gin.Context) {
	body := &structs.DictionaryBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Dictionary.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Update handles updating a dictionary.
//
// @Summary Update dictionary
// @Description Update an existing dictionary.
// @Tags dictionary
// @Accept json
// @Produce json
// @Param body body structs.UpdateDictionaryBody true "UpdateDictionaryBody object"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dictionarys [put]
// @Security Bearer
func (h *dictionaryHandler) Update(c *gin.Context) {
	body := &structs.UpdateDictionaryBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Dictionary.Update(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Get handles retrieving a dictionary by ID or slug.
//
// @Summary Get dictionary
// @Description Retrieve a dictionary by ID or slug.
// @Tags dictionary
// @Produce json
// @Param slug path string true "Dictionary ID or slug"
// @Param params query structs.FindDictionary true "FindDictionary parameters"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dictionarys/{slug} [get]
// @Security Bearer
func (h *dictionaryHandler) Get(c *gin.Context) {
	params := &structs.FindDictionary{Dictionary: c.Param("slug")}

	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Dictionary.Get(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// Delete handles deleting a dictionary by ID or slug.
//
// @Summary Delete dictionary
// @Description Delete a dictionary by ID or slug.
// @Tags dictionary
// @Produce json
// @Param slug path string true "Dictionary ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dictionarys/{slug} [delete]
// @Security Bearer
func (h *dictionaryHandler) Delete(c *gin.Context) {
	params := &structs.FindDictionary{Dictionary: c.Param("slug")}
	result, err := h.s.Dictionary.Delete(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// List handles listing all dictionarys.
//
// @Summary List dictionarys
// @Description Retrieve a list or tree structure of dictionarys.
// @Tags dictionary
// @Produce json
// @Param params query structs.ListDictionaryParams true "List dictionary parameters"
// @Success 200 {array} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/dictionarys [get]
// @Security Bearer
func (h *dictionaryHandler) List(c *gin.Context) {
	params := &structs.ListDictionaryParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Dictionary.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
