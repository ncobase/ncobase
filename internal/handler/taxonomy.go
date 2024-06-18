package handler

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/internal/data/structs"
	"ncobase/internal/helper"

	"github.com/gin-gonic/gin"
)

// CreateTaxonomyHandler handles the creation of a taxonomy.
//
// @Summary Create taxonomy
// @Description Create a new taxonomy.
// @Tags taxonomy
// @Accept json
// @Produce json
// @Param body body structs.CreateTaxonomyBody true "CreateTaxonomyBody object"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies [post]
// @Security Bearer
func (h *Handler) CreateTaxonomyHandler(c *gin.Context) {
	body := &structs.CreateTaxonomyBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.svc.CreateTaxonomyService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// UpdateTaxonomyHandler handles updating a taxonomy.
//
// @Summary Update taxonomy
// @Description Update an existing taxonomy.
// @Tags taxonomy
// @Accept json
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Param body body structs.UpdateTaxonomyBody true "UpdateTaxonomyBody object"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies/{slug} [put]
// @Security Bearer
func (h *Handler) UpdateTaxonomyHandler(c *gin.Context) {
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

	result, err := h.svc.UpdateTaxonomyService(c, slug, *updates)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetTaxonomyHandler handles getting a taxonomy.
//
// @Summary Get taxonomy
// @Description Retrieve details of a taxonomy.
// @Tags taxonomy
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies/{slug} [get]
func (h *Handler) GetTaxonomyHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.GetTaxonomyService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteTaxonomyHandler handles deleting a taxonomy.
//
// @Summary Delete taxonomy
// @Description Delete an existing taxonomy.
// @Tags taxonomy
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies/{slug} [delete]
// @Security Bearer
func (h *Handler) DeleteTaxonomyHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.svc.DeleteTaxonomyService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListTaxonomyHandler handles listing taxonomies.
//
// @Summary List taxonomies
// @Description Retrieve a list of taxonomies.
// @Tags taxonomy
// @Produce json
// @Param params query structs.ListTaxonomyParams true "ListTaxonomyParams object"
// @Success 200 {array} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies [get]
func (h *Handler) ListTaxonomyHandler(c *gin.Context) {
	params := &structs.ListTaxonomyParams{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	taxonomies, err := h.svc.ListTaxonomiesService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, taxonomies)
}
