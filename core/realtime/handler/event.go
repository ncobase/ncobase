package handler

import (
	"ncobase/realtime/service"
	"ncobase/realtime/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// EventHandler is the interface for the event handler.
type EventHandler interface {
	Publish(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	GetHistory(c *gin.Context)
}

// eventHandler represents the event handler.
type eventHandler struct {
	event service.EventService
}

// NewEventHandler creates a new event handler.
func NewEventHandler(e service.EventService) EventHandler {
	return &eventHandler{event: e}
}

// Publish publishes a new event
//
// @Summary Publish a new event
// @Description Publish a new event
// @Tags rt
// @Accept json
// @Produce json
// @Param body body structs.CreateEvent true "Event data"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/publish [post]
// @Security Bearer
func (h *eventHandler) Publish(c *gin.Context) {
	var body structs.CreateEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.event.Publish(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get gets an event by ID
//
// @Summary Get an event by ID
// @Description Get an event by ID
// @Tags rt
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/{id} [get]
// @Security Bearer
func (h *eventHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.event.Get(c.Request.Context(), &structs.FindEvent{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete deletes an event
//
// @Summary Delete an event
// @Description Delete an event
// @Tags rt
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/{id} [delete]
// @Security Bearer
func (h *eventHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.event.Delete(c.Request.Context(), &structs.FindEvent{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all events.
//
// @Summary List all events
// @Description Retrieve a list of events based on the provided query parameters
// @Tags rt
// @Produce json
// @Param params query structs.ListEventParams true "List events parameters"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events [get]
// @Security Bearer
func (h *eventHandler) List(c *gin.Context) {
	var params structs.ListEventParams
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.event.List(c.Request.Context(), &params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// GetHistory gets event history
//
// @Summary Get event history
// @Description Get event history
// @Tags rt
// @Produce json
// @Param channel_id query string true "Channel ID"
// @Param type query string true "Event type"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/history [get]
// @Security Bearer
func (h *eventHandler) GetHistory(c *gin.Context) {
	channelID := c.Query("channel_id")
	eventType := c.Query("type")

	if channelID == "" || eventType == "" {
		resp.Fail(c.Writer, resp.BadRequest("channel_id and type are required"))
		return
	}

	result, err := h.event.GetEventHistory(c.Request.Context(), channelID, eventType)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}
