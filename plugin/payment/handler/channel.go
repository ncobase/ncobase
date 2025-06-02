package handler

import (
	"ncobase/payment/service"
	"ncobase/payment/structs"

	"github.com/gin-gonic/gin"
	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"
)

// ChannelHandlerInterface defines the interface for channel handler operations
type ChannelHandlerInterface interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	ChangeStatus(c *gin.Context)
	List(c *gin.Context)
}

// channelHandler handles payment channel-related requests
type channelHandler struct {
	svc service.ChannelServiceInterface
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(svc service.ChannelServiceInterface) ChannelHandlerInterface {
	return &channelHandler{svc: svc}
}

// Create handles the creation of a new payment channel
//
// @Summary Create payment channel
// @Description Create a new payment channel
// @Tags payment,channels
// @Accept json
// @Produce json
// @Param body body structs.CreateChannelInput true "Channel data"
// @Success 200 {object} resp.Exception{data=structs.Channel} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels [post]
// @Security Bearer
func (h *channelHandler) Create(c *gin.Context) {
	var input structs.CreateChannelInput
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Create channel
	channel, err := h.svc.Create(c.Request.Context(), &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to create payment channel: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to create payment channel", err))
		return
	}

	resp.Success(c.Writer, channel)
}

// Get handles getting a payment channel by ID
//
// @Summary Get payment channel
// @Description Get a payment channel by ID
// @Tags payment,channels
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} resp.Exception{data=structs.Channel} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 404 {object} resp.Exception "not found"
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels/{id} [get]
// @Security Bearer
func (h *channelHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Get channel
	channel, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to get payment channel: %v", err)
		resp.Fail(c.Writer, resp.NotFound("Payment channel not found"))
		return
	}

	resp.Success(c.Writer, channel)
}

// Update handles updating a payment channel
//
// @Summary Update payment channel
// @Description Update an existing payment channel
// @Tags payment,channels
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body structs.UpdateChannelInput true "Updates"
// @Success 200 {object} resp.Exception{data=structs.Channel} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels/{id} [put]
// @Security Bearer
func (h *channelHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var input structs.UpdateChannelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request data", err))
		return
	}

	// Set ID from path parameter
	input.ID = id

	// Update channel
	channel, err := h.svc.Update(c.Request.Context(), id, &input)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to update payment channel: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to update payment channel", err))
		return
	}

	resp.Success(c.Writer, channel)
}

// Delete handles deleting a payment channel
//
// @Summary Delete payment channel
// @Description Delete a payment channel by ID
// @Tags payment,channels
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} resp.Exception{data=nil} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels/{id} [delete]
// @Security Bearer
func (h *channelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	// Delete channel
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		logger.Errorf(c.Request.Context(), "Failed to delete payment channel: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to delete payment channel", err))
		return
	}

	resp.Success(c.Writer, nil)
}

// ChangeStatus handles changing the status of a payment channel
//
// @Summary Change channel status
// @Description Change the status of a payment channel
// @Tags payment,channels
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body structs.ChannelStatus true "Status data"
// @Success 200 {object} resp.Exception{data=structs.Channel} "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels/{id}/status [put]
// @Security Bearer
func (h *channelHandler) ChangeStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var input struct {
		Status structs.ChannelStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid request data", err))
		return
	}

	// Change status
	channel, err := h.svc.ChangeStatus(c.Request.Context(), id, input.Status)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to change payment channel status: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to change payment channel status", err))
		return
	}

	resp.Success(c.Writer, channel)
}

// List handles listing payment channels
//
// @Summary List payment channels
// @Description Get a paginated list of payment channels
// @Tags payment,channels
// @Produce json
// @Param provider query string false "Filter by provider"
// @Param status query string false "Filter by status"
// @Param tenant_id query string false "Filter by tenant ID"
// @Param cursor query string false "Cursor for pagination"
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} structs.Channel "success"
// @Failure 400 {object} resp.Exception "bad request
// @Failure 500 {object} resp.Exception "internal server error"
// @Router /pay/channels [get]
// @Security Bearer
func (h *channelHandler) List(c *gin.Context) {
	var query structs.ChannelQuery
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, &query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Set default page size if not provided
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// List channels
	result, err := h.svc.List(c.Request.Context(), &query)
	if err != nil {
		logger.Errorf(c.Request.Context(), "Failed to list payment channels: %v", err)
		resp.Fail(c.Writer, resp.InternalServer("Failed to list payment channels", err))
		return
	}

	resp.Success(c.Writer, result)
}
