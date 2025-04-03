package handler

import (
	"ncobase/core/realtime/service"
	"ncobase/core/realtime/structs"
	"ncobase/ncore/ecode"
	"ncobase/ncore/resp"

	"github.com/gin-gonic/gin"
)

// NotificationHandler is the interface for the notification handler.
type NotificationHandler interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	List(c *gin.Context)
	MarkAsRead(c *gin.Context)
	MarkAllAsRead(c *gin.Context)
	MarkAsUnread(c *gin.Context)
	MarkAllAsUnread(c *gin.Context)
}

// notificationHandler represents the notification handler.
type notificationHandler struct {
	notification service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(n service.NotificationService) NotificationHandler {
	return &notificationHandler{notification: n}
}

// Create creates a new notification
//
// @Summary Create a new notification
// @Description Create a new notification
// @Tags rt
// @Accept json
// @Produce json
// @Param body body structs.CreateNotification true "Notification data"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications [post]
// @Security Bearer
func (h *notificationHandler) Create(c *gin.Context) {
	var body structs.CreateNotification
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.notification.Create(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Get gets a notification by ID
//
// @Summary Get a notification by ID
// @Description Get a notification by ID
// @Tags rt
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/{id} [get]
// @Security Bearer
func (h *notificationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.notification.Get(c.Request.Context(), &structs.FindNotification{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Update updates a notification
//
// @Summary Update a notification
// @Description Update a notification
// @Tags rt
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Param body body structs.UpdateNotification true "Notification data"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/{id} [put]
// @Security Bearer
func (h *notificationHandler) Update(c *gin.Context) {
	var body structs.UpdateNotification
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	body.ID = c.Param("id")
	if body.ID == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	result, err := h.notification.Update(c.Request.Context(), &body)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// Delete deletes a notification
//
// @Summary Delete a notification
// @Description Delete a notification
// @Tags rt
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/{id} [delete]
// @Security Bearer
func (h *notificationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.notification.Delete(c.Request.Context(), &structs.FindNotification{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// List handles listing all notifications.
//
// @Summary List all notifications
// @Description Retrieve a list of notifications based on the provided query parameters
// @Tags rt
// @Produce json
// @Param params query structs.ListNotificationParams true "List notifications parameters"
// @Success 200 {array} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications [get]
// @Security Bearer
func (h *notificationHandler) List(c *gin.Context) {
	var params structs.ListNotificationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	result, err := h.notification.List(c.Request.Context(), &params)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer, result)
}

// MarkAsRead marks a notification as read
//
// @Summary Mark a notification as read
// @Description Mark a notification as read
// @Tags rt
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/{id}/read [put]
// @Security Bearer
func (h *notificationHandler) MarkAsRead(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.notification.MarkAsRead(c.Request.Context(), &structs.FindNotification{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// MarkAllAsRead marks all notifications as read
//
// @Summary Mark all notifications as read
// @Description Mark all notifications as read
// @Tags rt
// @Produce json
// @Success 200 {object} map[string]any{message=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/read-all [put]
// @Security Bearer
func (h *notificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id") // From auth middleware
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("user not authenticated"))
		return
	}

	err := h.notification.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// MarkAsUnread marks a notification as unread
//
// @Summary Mark a notification as unread
// @Description Mark a notification as unread
// @Tags rt
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} structs.ReadNotification "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/{id}/unread [put]
// @Security Bearer
func (h *notificationHandler) MarkAsUnread(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		resp.Fail(c.Writer, resp.BadRequest(ecode.FieldIsRequired("id")))
		return
	}

	err := h.notification.MarkAsUnread(c.Request.Context(), &structs.FindNotification{ID: id})
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}

// MarkAllAsUnread marks all notifications as unread
//
// @Summary Mark all notifications as unread
// @Description Mark all notifications as unread
// @Tags rt
// @Produce json
// @Success 200 {object} map[string]any{message=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /rt/notifications/unread-all [put]
// @Security Bearer
func (h *notificationHandler) MarkAllAsUnread(c *gin.Context) {
	userID := c.GetString("user_id") // From auth middleware
	if userID == "" {
		resp.Fail(c.Writer, resp.UnAuthorized("user not authenticated"))
		return
	}

	err := h.notification.MarkAllAsUnread(c.Request.Context(), userID)
	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
		return
	}

	resp.Success(c.Writer)
}
