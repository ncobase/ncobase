package handler

import (
	"ncobase/plugin/sample/service"
	"ncobase/plugin/sample/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/response"
)

// Handler provides HTTP handlers for the sample plugin
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all routes for the sample plugin
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	samples := r.Group("/samples")
	{
		samples.GET("", h.List)
		samples.GET("/:id", h.Get)
		samples.POST("", h.Create)
		samples.PUT("/:id", h.Update)
		samples.DELETE("/:id", h.Delete)
	}
}

// List godoc
// @Summary List samples
// @Description Get a list of all samples
// @Tags samples
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]structs.Sample}
// @Failure 500 {object} response.Response
// @Router /samples [get]
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	samples, err := h.service.Sample.List(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to list samples: %v", err)
		response.Error(c, err)
		return
	}

	response.Success(c, samples)
}

// Get godoc
// @Summary Get sample by ID
// @Description Get a single sample by its ID
// @Tags samples
// @Accept json
// @Produce json
// @Param id path string true "Sample ID"
// @Success 200 {object} response.Response{data=structs.Sample}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /samples/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	sample, err := h.service.Sample.GetByID(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "Failed to get sample: %v", err)
		response.Error(c, err)
		return
	}

	response.Success(c, sample)
}

// Create godoc
// @Summary Create sample
// @Description Create a new sample
// @Tags samples
// @Accept json
// @Produce json
// @Param input body structs.CreateSampleInput true "Sample data"
// @Success 201 {object} response.Response{data=structs.Sample}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /samples [post]
func (h *Handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	var input structs.CreateSampleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	sample, err := h.service.Sample.Create(ctx, &input)
	if err != nil {
		logger.Errorf(ctx, "Failed to create sample: %v", err)
		response.Error(c, err)
		return
	}

	response.Created(c, sample)
}

// Update godoc
// @Summary Update sample
// @Description Update an existing sample
// @Tags samples
// @Accept json
// @Produce json
// @Param id path string true "Sample ID"
// @Param input body structs.UpdateSampleInput true "Sample data"
// @Success 200 {object} response.Response{data=structs.Sample}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /samples/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var input structs.UpdateSampleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err)
		return
	}

	sample, err := h.service.Sample.Update(ctx, id, &input)
	if err != nil {
		logger.Errorf(ctx, "Failed to update sample: %v", err)
		response.Error(c, err)
		return
	}

	response.Success(c, sample)
}

// Delete godoc
// @Summary Delete sample
// @Description Delete a sample by ID
// @Tags samples
// @Accept json
// @Produce json
// @Param id path string true "Sample ID"
// @Success 204 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /samples/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	if err := h.service.Sample.Delete(ctx, id); err != nil {
		logger.Errorf(ctx, "Failed to delete sample: %v", err)
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
