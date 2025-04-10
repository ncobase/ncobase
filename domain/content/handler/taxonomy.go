package handler

import (
	"ncobase/domain/content/service"
	"ncobase/domain/content/structs"

	"github.com/ncobase/ncore/pkg/helper"

	"github.com/ncobase/ncore/pkg/ecode"
	"github.com/ncobase/ncore/pkg/resp"
	"github.com/ncobase/ncore/pkg/types"

	"github.com/gin-gonic/gin"
)

// TaxonomyHandlerInterface is the interface for the handler.
type TaxonomyHandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

// taxonomyHandler represents the handler.
type taxonomyHandler struct {
	s *service.Service
}

// NewTaxonomyHandler creates a new handler.
func NewTaxonomyHandler(svc *service.Service) TaxonomyHandlerInterface {
	return &taxonomyHandler{
		s: svc,
	}
}

// Create handles the creation of a taxonomy.
//
// @Summary Create taxonomy
// @Description Create a new taxonomy.
// @Tags cms
// @Accept json
// @Produce json
// @Param body body structs.CreateTaxonomyBody true "CreateTaxonomyBody object"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/taxonomies [post]
// @Security Bearer
func (h *taxonomyHandler) Create(c *gin.Context) {
	body := &structs.CreateTaxonomyBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Taxonomy.Create(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a taxonomy.
//
// @Summary Update taxonomy
// @Description Update an existing taxonomy.
// @Tags cms
// @Accept json
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Param body body structs.UpdateTaxonomyBody true "UpdateTaxonomyBody object"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/taxonomies/{slug} [put]
// @Security Bearer
func (h *taxonomyHandler) Update(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	updates := &types.JSON{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Taxonomy.Update(c.Request.Context(), slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get handles getting a taxonomy.
//
// @Summary Get taxonomy
// @Description Retrieve details of a taxonomy.
// @Tags cms
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/taxonomies/{slug} [get]
func (h *taxonomyHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Taxonomy.Get(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete handles deleting a taxonomy.
//
// @Summary Delete taxonomy
// @Description Delete an existing taxonomy.
// @Tags cms
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/taxonomies/{slug} [delete]
// @Security Bearer
func (h *taxonomyHandler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	if err := h.s.Taxonomy.Delete(c.Request.Context(), slug); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing taxonomies.
//
// @Summary List taxonomies
// @Description Retrieve a list of taxonomies.
// @Tags cms
// @Produce json
// @Param params query structs.ListTaxonomyParams true "ListTaxonomyParams object"
// @Success 200 {array} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /cms/taxonomies [get]
func (h *taxonomyHandler) List(c *gin.Context) {
	params := &structs.ListTaxonomyParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	taxonomies, err := h.s.Taxonomy.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, taxonomies)
}
