package handler

import (
	"net/http"

	"ncobase/plugin/sample/service"
	"ncobase/plugin/sample/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/net/resp"
)

// Handler provides HTTP handlers for the sample plugin.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new Handler.
func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers all routes for the sample plugin.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/samples")
	{
		g.GET("", h.List)
		g.GET("/:id", h.Get)
		g.POST("", h.Create)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}

// List godoc
//
// @Summary      List samples
// @Description  Get a list of all samples.
// @Tags         samples
// @Produce      json
// @Success      200  {array}   structs.Sample    "success"
// @Failure      500  {object}  resp.Exception    "internal server error"
// @Router       /samples [get]
func (h *Handler) List(c *gin.Context) {
	samples, err := h.service.Sample.List(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, samples)
}

// Get godoc
//
// @Summary      Get sample
// @Description  Get a single sample by ID.
// @Tags         samples
// @Produce      json
// @Param        id   path      string          true  "Sample ID"
// @Success      200  {object}  structs.Sample  "success"
// @Failure      404  {object}  resp.Exception  "not found"
// @Failure      500  {object}  resp.Exception  "internal server error"
// @Router       /samples/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	sample, err := h.service.Sample.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		resp.Fail(c.Writer, resp.NotFound(err.Error()))
		return
	}
	resp.Success(c.Writer, sample)
}

// Create godoc
//
// @Summary      Create sample
// @Description  Create a new sample.
// @Tags         samples
// @Accept       json
// @Produce      json
// @Param        body  body      structs.CreateSampleInput  true  "Sample data"
// @Success      201   {object}  structs.Sample             "created"
// @Failure      400   {object}  resp.Exception             "bad request"
// @Failure      500   {object}  resp.Exception             "internal server error"
// @Router       /samples [post]
func (h *Handler) Create(c *gin.Context) {
	var body structs.CreateSampleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	sample, err := h.service.Sample.Create(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.WithStatusCode(c.Writer, http.StatusCreated, sample)
}

// Update godoc
//
// @Summary      Update sample
// @Description  Update an existing sample.
// @Tags         samples
// @Accept       json
// @Produce      json
// @Param        id    path      string                     true  "Sample ID"
// @Param        body  body      structs.UpdateSampleInput  true  "Sample data"
// @Success      200   {object}  structs.Sample             "success"
// @Failure      400   {object}  resp.Exception             "bad request"
// @Failure      404   {object}  resp.Exception             "not found"
// @Failure      500   {object}  resp.Exception             "internal server error"
// @Router       /samples/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var body structs.UpdateSampleInput
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	sample, err := h.service.Sample.Update(c.Request.Context(), c.Param("id"), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.Success(c.Writer, sample)
}

// Delete godoc
//
// @Summary      Delete sample
// @Description  Delete a sample by ID.
// @Tags         samples
// @Produce      json
// @Param        id   path      string          true  "Sample ID"
// @Success      204  "no content"
// @Failure      404  {object}  resp.Exception  "not found"
// @Failure      500  {object}  resp.Exception  "internal server error"
// @Router       /samples/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Sample.Delete(c.Request.Context(), c.Param("id")); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}
	resp.WithStatusCode(c.Writer, http.StatusNoContent)
}
