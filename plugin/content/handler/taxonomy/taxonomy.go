package taxonomy

import (
	"ncobase/helper"
	"ncobase/plugin/content/service"
	"ncobase/plugin/content/structs"

	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"

	"github.com/gin-gonic/gin"
)

// HandlerInterface is the interface for the taxonomy handler.
type HandlerInterface interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
}

type Handler struct {
	s *service.Service
}

func New(svc *service.Service) HandlerInterface {
	return &Handler{
		s: svc,
	}
}

// Create handles the creation of a taxonomy.
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
func (h *Handler) Create(c *gin.Context) {
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
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update handles updating a taxonomy.
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
func (h *Handler) Update(c *gin.Context) {
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
// @Tags taxonomy
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies/{slug} [get]
func (h *Handler) Get(c *gin.Context) {
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
// @Tags taxonomy
// @Produce json
// @Param slug path string true "Taxonomy slug"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies/{slug} [delete]
// @Security Bearer
func (h *Handler) Delete(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("slug / id")))
		return
	}

	result, err := h.s.Taxonomy.Delete(c.Request.Context(), slug)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// List handles listing taxonomies.
//
// @Summary List taxonomies
// @Description Retrieve a list of taxonomies.
// @Tags taxonomy
// @Produce json
// @Param params query structs.ListTaxonomyParams true "ListTaxonomyParams object"
// @Success 200 {array} structs.ReadTaxonomy "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/taxonomies [get]
func (h *Handler) List(c *gin.Context) {
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
