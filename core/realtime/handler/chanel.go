package handler

import (
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/core/realtime/service"
	"ncobase/core/realtime/structs"

	"github.com/gin-gonic/gin"
)

// ChannelHandler is the interface for the channel handler.
type ChannelHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	Subscribe(c *gin.Context)
	Unsubscribe(c *gin.Context)
	GetSubscribers(c *gin.Context)
	GetUserChannels(c *gin.Context)
}

// channelHandler represents the channel handler.
type channelHandler struct {
	channel service.ChannelService
}

// NewChannelHandler creates a new channel handler.
func NewChannelHandler(ch service.ChannelService) ChannelHandler {
	return &channelHandler{channel: ch}
}

// Create creates a new channel
//
// @Summary Create a new channel
// @Description Create a new channel
// @Tags rt
// @Accept json
// @Produce json
// @Param body body structs.CreateChannel true "Channel data"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels [post]
// @Security Bearer
func (h *channelHandler) Create(c *gin.Context) {
	var body structs.CreateChannel
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.channel.Create(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get gets a channel by ID
//
// @Summary Get a channel by ID
// @Description Get a channel by ID
// @Tags rt
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id} [get]
// @Security Bearer
func (h *channelHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.channel.Get(c.Request.Context(), &structs.FindChannel{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update updates a channel
//
// @Summary Update a channel
// @Description Update a channel
// @Tags rt
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body structs.UpdateChannel true "Channel data"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id} [put]
// @Security Bearer
func (h *channelHandler) Update(c *gin.Context) {
	var body structs.UpdateChannel
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	body.ID = c.Param("id")
	if body.ID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.channel.Update(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete deletes a channel
//
// @Summary Delete a channel
// @Description Delete a channel
// @Tags rt
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id} [delete]
// @Security Bearer
func (h *channelHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.channel.Delete(c.Request.Context(), &structs.FindChannel{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all channels.
//
// @Summary List all channels
// @Description Retrieve a list of channels based on the provided query parameters
// @Tags rt
// @Produce json
// @Param params query structs.ListChannelParams true "List channels parameters"
// @Success 200 {array} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels [get]
// @Security Bearer
func (h *channelHandler) List(c *gin.Context) {
	var params structs.ListChannelParams
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.channel.List(c.Request.Context(), &params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Subscribe subscribes a user to a channel
//
// @Summary Subscribe to a channel
// @Description Subscribe to a channel
// @Tags rt
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Param body body structs.CreateSubscription true "Subscription data"
// @Success 200 {object} structs.ReadSubscription "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id}/subscribe [post]
// @Security Bearer
func (h *channelHandler) Subscribe(c *gin.Context) {
	var body structs.CreateSubscription
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Set channel ID from path
	body.Subscription.ChannelID = c.Param("id")
	if body.Subscription.ChannelID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("channel_id")))
		return
	}

	// Set user ID from auth context if not provided
	if body.Subscription.UserID == "" {
		body.Subscription.UserID = c.GetString("user_id")
		if body.Subscription.UserID == "" {
			resp.Fail(c.Writer, resp.UnAuthorized("user not authenticated"))
			return
		}
	}

	result, err := h.channel.Subscribe(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Unsubscribe unsubscribes a user from a channel
//
// @Summary Unsubscribe from a channel
// @Description Unsubscribe from a channel
// @Tags rt
// @Accept json
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {object} map[string]any{message=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id}/unsubscribe [post]
// @Security Bearer
func (h *channelHandler) Unsubscribe(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("channel_id")))
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("user not authenticated"))
		return
	}

	err := h.channel.Unsubscribe(c.Request.Context(), userID, channelID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// GetSubscribers gets subscribers of a channel
//
// @Summary Get subscribers of a channel
// @Description Get subscribers of a channel
// @Tags rt
// @Produce json
// @Param id path string true "Channel ID"
// @Success 200 {array} structs.ReadUser "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/{id}/subscribers [get]
// @Security Bearer
func (h *channelHandler) GetSubscribers(c *gin.Context) {
	channelID := c.Param("id")
	if channelID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("channel_id")))
		return
	}

	subscribers, err := h.channel.GetSubscribers(c.Request.Context(), channelID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, subscribers)
}

// GetUserChannels gets channels of a user
//
// @Summary Get channels of a user
// @Description Get channels of a user
// @Tags rt
// @Produce json
// @Success 200 {array} structs.ReadChannel "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/channels/user [get]
// @Security Bearer
func (h *channelHandler) GetUserChannels(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("user not authenticated"))
		return
	}

	channels, err := h.channel.GetUserChannels(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, channels)
}
