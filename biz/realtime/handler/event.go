package handler

import (
	"ncobase/realtime/service"
	"ncobase/realtime/structs"
	"strconv"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/net/resp"

	"github.com/gin-gonic/gin"
)

// EventHandler is the interface for the event handler
type EventHandler interface {
	Publish(c *gin.Context)
	Get(c *gin.Context)
	Search(c *gin.Context)
	RealtimeStats(c *gin.Context)
	Retry(c *gin.Context)
	List(c *gin.Context)
	PublishExtended(c *gin.Context)
	Delete(c *gin.Context)
	GetHistory(c *gin.Context)
	PublishBatch(c *gin.Context)
	GetFailedEvents(c *gin.Context)
	ProcessPendingEvents(c *gin.Context)
	UpdateEventStatus(c *gin.Context)
	GetEventTypes(c *gin.Context)
	GetEventSources(c *gin.Context)
}

// eventHandler represents the event handler
type eventHandler struct {
	event service.EventService
}

// NewEventHandler creates a new event handler
func NewEventHandler(e service.EventService) EventHandler {
	return &eventHandler{event: e}
}

// Publish publishes a new event
//
// @Summary Publish a new event
// @Description Publish a new event
// @Tags events
// @Accept json
// @Produce json
// @Param body body structs.CreateEvent true "Event data"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events [post]
// @Security Bearer
func (h *eventHandler) Publish(c *gin.Context) {
	var body structs.CreateEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Set source from header if not provided
	if body.Event.Source == "" {
		body.Event.Source = c.GetHeader("X-Source")
		if body.Event.Source == "" {
			body.Event.Source = "unknown"
		}
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
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/{id} [get]
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

// Search performs complex search queries
//
// @Summary Search events
// @Description Search events with complex queries and aggregations
// @Tags events
// @Accept json
// @Produce json
// @Param body body structs.SearchQuery true "Search query"
// @Success 200 {object} structs.SearchResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /search [post]
// @Security Bearer
func (h *eventHandler) Search(c *gin.Context) {
	var query structs.SearchQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Set default values
	if query.Size <= 0 {
		query.Size = 100
	}
	if query.Size > 1000 {
		query.Size = 1000 // Limit max size
	}

	result, err := h.event.Search(c.Request.Context(), &query)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// RealtimeStats gets real-time statistics
//
// @Summary Get real-time statistics
// @Description Get real-time event statistics and metrics
// @Tags events
// @Produce json
// @Param interval query string false "Time interval (1m, 5m, 15m, 1h)"
// @Param type query string false "Statistics type (overview, detailed)"
// @Success 200 {object} structs.RealtimeStats "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /stats/realtime [get]
// @Security Bearer
func (h *eventHandler) RealtimeStats(c *gin.Context) {
	params := &structs.StatsParams{
		Interval: c.DefaultQuery("interval", "5m"),
		Type:     c.DefaultQuery("type", "overview"),
	}

	// Parse time range if provided
	if start := c.Query("start"); start != "" {
		end := c.DefaultQuery("end", "")
		params.TimeRange = &structs.TimeRange{
			Start: start,
			End:   end,
		}
	}

	result, err := h.event.GetRealtimeStats(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Retry retries a failed event
//
// @Summary Retry a failed event
// @Description Retry processing of a failed event
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param body body structs.RetryParams false "Retry parameters"
// @Success 200 {object} structs.RetryResult "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/{id}/retry [post]
// @Security Bearer
func (h *eventHandler) Retry(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var params structs.RetryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		// If no body provided, use defaults
		params = structs.RetryParams{
			Reason:   "manual_retry",
			Priority: "normal",
		}
	}

	result, err := h.event.RetryEvent(c.Request.Context(), id, &params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Extended API implementations (existing functionality)

// List lists all events
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

	// Set default limit
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 500 {
		params.Limit = 500
	}

	result, err := h.event.List(c.Request.Context(), &params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// PublishExtended publishes a new event (extended API)
//
// @Summary Publish a new event (extended)
// @Description Publish a new event with full options
// @Tags rt
// @Accept json
// @Produce json
// @Param body body structs.CreateEvent true "Event data"
// @Success 200 {object} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/publish [post]
// @Security Bearer
func (h *eventHandler) PublishExtended(c *gin.Context) {
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

// Delete deletes an event
//
// @Summary Delete an event
// @Description Delete an event
// @Tags rt
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} resp.Exception "success"
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

// GetHistory gets event history
//
// @Summary Get event history
// @Description Get event history with filters
// @Tags rt
// @Produce json
// @Param type query string false "Event type"
// @Param source query string false "Event source"
// @Param limit query int false "Limit results"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/events/history [get]
// @Security Bearer
func (h *eventHandler) GetHistory(c *gin.Context) {
	params := &structs.ListEventParams{
		Type:   c.Query("type"),
		Source: c.Query("source"),
		Limit:  100, // Default limit
	}

	// Parse limit if provided
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit := parseInt(limitStr, 100); limit > 0 && limit <= 1000 {
			params.Limit = limit
		}
	}

	result, err := h.event.List(c.Request.Context(), params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result.Items)
}

// Batch operations

// PublishBatch publishes multiple events
//
// @Summary Publish multiple events
// @Description Publish multiple events in a single batch
// @Tags events
// @Accept json
// @Produce json
// @Param body body []structs.CreateEvent true "Array of event data"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/batch [post]
// @Security Bearer
func (h *eventHandler) PublishBatch(c *gin.Context) {
	var events []*structs.CreateEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if len(events) == 0 {
		resp.Fail(c.Writer, resp.BadRequest("at least one event is required"))
		return
	}

	if len(events) > 100 {
		resp.Fail(c.Writer, resp.BadRequest("maximum 100 events per batch"))
		return
	}

	// Set source from header if not provided
	source := c.GetHeader("X-Source")
	if source == "" {
		source = "unknown"
	}

	for _, event := range events {
		if event.Event.Source == "" {
			event.Event.Source = source
		}
	}

	results, err := h.event.PublishBatch(c.Request.Context(), events)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, results)
}

// GetFailedEvents gets failed events for retry
//
// @Summary Get failed events
// @Description Get events that failed processing
// @Tags events
// @Produce json
// @Param limit query int false "Limit results (default: 50, max: 200)"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/failed [get]
// @Security Bearer
func (h *eventHandler) GetFailedEvents(c *gin.Context) {
	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit := parseInt(limitStr, 50); parsedLimit > 0 && parsedLimit <= 200 {
			limit = parsedLimit
		}
	}

	results, err := h.event.GetFailedEvents(c.Request.Context(), limit)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, results)
}

// ProcessPendingEvents processes pending events
//
// @Summary Process pending events
// @Description Manually trigger processing of pending events
// @Tags events
// @Produce json
// @Param limit query int false "Limit events to process (default: 10, max: 50)"
// @Success 200 {array} structs.ReadEvent "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/process [post]
// @Security Bearer
func (h *eventHandler) ProcessPendingEvents(c *gin.Context) {
	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit := parseInt(limitStr, 10); parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	results, err := h.event.ProcessPendingEvents(c.Request.Context(), limit)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, results)
}

// UpdateEventStatus updates an event's status
//
// @Summary Update event status
// @Description Update the processing status of an event
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param body body object{status=string,error_message=string} true "Status update"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /events/{id}/status [put]
// @Security Bearer
func (h *eventHandler) UpdateEventStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	var body struct {
		Status       string `json:"status" binding:"required"`
		ErrorMessage string `json:"error_message,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	// Validate status
	validStatuses := []string{"pending", "processed", "failed", "retry"}
	isValid := false
	for _, status := range validStatuses {
		if body.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		resp.Fail(c.Writer, resp.BadRequest("invalid status. Valid values: pending, processed, failed, retry"))
		return
	}

	err := h.event.UpdateEventStatus(c.Request.Context(), id, body.Status, body.ErrorMessage)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// Health check and utility endpoints

// GetEventTypes gets available event types
//
// @Summary Get event types
// @Description Get list of available event types
// @Tags events
// @Produce json
// @Success 200 {object} object{types=[]string} "success"
// @Router /events/types [get]
// @Security Bearer
func (h *eventHandler) GetEventTypes(c *gin.Context) {
	// In production, this would query the database for distinct event types
	types := []string{
		"user_action",
		"system_log",
		"error",
		"audit",
		"notification",
		"metric",
		"alert",
		"business_event",
	}

	resp.Success(c.Writer, map[string]any{
		"types": types,
	})
}

// GetEventSources gets available event sources
//
// @Summary Get event sources
// @Description Get list of available event sources
// @Tags events
// @Produce json
// @Success 200 {object} object{sources=[]string} "success"
// @Router /events/sources [get]
// @Security Bearer
func (h *eventHandler) GetEventSources(c *gin.Context) {
	// In production, this would query the database for distinct sources
	sources := []string{
		"mobile_app",
		"web_app",
		"backend_service",
		"iot_device",
		"third_party",
		"system",
		"admin_panel",
		"api_gateway",
	}

	resp.Success(c.Writer, map[string]any{
		"sources": sources,
	})
}

// Helper functions

// parseInt parses string to int with default value
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return val
}
