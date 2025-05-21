package handler

import (
	"ncobase/system/service"
	"ncobase/system/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

// DictionaryHandlerInterface represents the dictionary handler interface.
type DictionaryHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	GetBySlug(c *gin.Context)
	GetEnumOptions(c *gin.Context)
	ValidateEnumValue(c *gin.Context)
	BatchGetBySlug(c *gin.Context)
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
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.DictionaryBody true "DictionaryBody object"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionarys [post]
// @Security Bearer
func (h *dictionaryHandler) Create(c *gin.Context) {
	body := &structs.DictionaryBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
// @Tags sys
// @Accept json
// @Produce json
// @Param body body structs.UpdateDictionaryBody true "UpdateDictionaryBody object"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionarys [put]
// @Security Bearer
func (h *dictionaryHandler) Update(c *gin.Context) {
	body := &structs.UpdateDictionaryBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
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
// @Tags sys
// @Produce json
// @Param slug path string true "Dictionary ID or slug"
// @Param params query structs.FindDictionary true "FindDictionary parameters"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionarys/{slug} [get]
// @Security Bearer
func (h *dictionaryHandler) Get(c *gin.Context) {
	params := &structs.FindDictionary{Dictionary: c.Param("slug")}

	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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

// GetBySlug handles retrieving a dictionary by slug.
//
// @Summary Get dictionary by slug
// @Description Retrieve a dictionary by its slug.
// @Tags sys
// @Produce json
// @Param slug path string true "Dictionary slug"
// @Success 200 {object} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionaries/slug/{slug} [get]
// @Security Bearer
func (h *dictionaryHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.s.Dictionary.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetEnumOptions handles retrieving dictionary options for frontend select components.
//
// @Summary Get dictionary enum options
// @Description Retrieve dictionary options formatted for frontend select components.
// @Tags sys
// @Produce json
// @Param slug path string true "Dictionary slug"
// @Success 200 {array} map[string]interface{} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionaries/options/{slug} [get]
// @Security Bearer
func (h *dictionaryHandler) GetEnumOptions(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.s.Dictionary.GetEnumOptions(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ValidateEnumValue handles validating if a value exists in a dictionary enum.
//
// @Summary Validate enum value
// @Description Check if a value exists in a dictionary enum.
// @Tags sys
// @Produce json
// @Param slug path string true "Dictionary slug"
// @Param value query string true "Value to validate"
// @Success 200 {object} map[string]bool "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionaries/validate/{slug} [get]
// @Security Bearer
func (h *dictionaryHandler) ValidateEnumValue(c *gin.Context) {
	slug := c.Param("slug")
	value := c.Query("value")

	valid, err := h.s.Dictionary.ValidateEnumValue(c.Request.Context(), slug, value)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, map[string]bool{"valid": valid})
}

// BatchGetBySlug handles retrieving multiple dictionaries by their slugs.
//
// @Summary Batch get dictionaries
// @Description Retrieve multiple dictionaries by their slugs.
// @Tags sys
// @Accept json
// @Produce json
// @Param body body []string true "Array of dictionary slugs"
// @Success 200 {object} map[string]structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionaries/batch [post]
// @Security Bearer
func (h *dictionaryHandler) BatchGetBySlug(c *gin.Context) {
	var slugs []string
	if err := c.ShouldBindJSON(&slugs); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.s.Dictionary.BatchGetBySlug(c.Request.Context(), slugs)
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
// @Tags sys
// @Produce json
// @Param slug path string true "Dictionary ID or slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionarys/{slug} [delete]
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
// @Tags sys
// @Produce json
// @Param params query structs.ListDictionaryParams true "List dictionary parameters"
// @Success 200 {array} structs.ReadDictionary "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sys/dictionarys [get]
// @Security Bearer
func (h *dictionaryHandler) List(c *gin.Context) {
	params := &structs.ListDictionaryParams{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, params); err != nil {
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
