package handler

import (
	"stocms/internal/data/structs"
	"stocms/pkg/resp"

	"github.com/gin-gonic/gin"
)

// CreateTaxonomyHandler handles creation of a taxonomy
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

// UpdateTaxonomyHandler handles updating a taxonomy
func (h *Handler) UpdateTaxonomyHandler(c *gin.Context) {
	var body *structs.UpdateTaxonomyBody
	if err := c.ShouldBind(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.svc.UpdateTaxonomyService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetTaxonomyHandler handles getting a taxonomy
func (h *Handler) GetTaxonomyHandler(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.svc.GetTaxonomyService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// DeleteTaxonomyHandler handles deleting a taxonomy
func (h *Handler) DeleteTaxonomyHandler(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.svc.DeleteTaxonomyService(c, slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// ListTaxonomyHandler handles listing taxonomies
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
