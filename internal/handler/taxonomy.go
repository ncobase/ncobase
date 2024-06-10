package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/ecode"
	"stocms/pkg/resp"
	"stocms/pkg/types"

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
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /taxonomy [post]
func (h *Handler) CreateTaxonomyHandler(c *gin.Context) {
	var body *structs.CreateTaxonomyBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
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
// @Param body body object true "Update data"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /taxonomy/{slug} [put]
func (h *Handler) UpdateTaxonomyHandler(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	var updates types.JSON
	if err := c.ShouldBind(&updates); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateTaxonomyService(c, slug, updates)
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
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /taxonomy/{slug} [get]
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
// @Router /taxonomy/{slug} [delete]
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
// @Param category query string false "Category filter"
// @Param limit query int false "Result limit"
// @Param offset query int false "Result offset"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /taxonomy [get]
func (h *Handler) ListTaxonomyHandler(c *gin.Context) {
	var params *structs.ListTaxonomyParams
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	taxonomies, err := h.svc.ListTaxonomiesService(c, params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, taxonomies)
}
