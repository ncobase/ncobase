package handler

import (
	"context"
	"ncobase/access/event"
	"ncobase/access/service"
	"ncobase/access/structs"
	"strings"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/types"
)

// EventHandlerInterface defines interface for event handler operations
type EventHandlerInterface interface {
	GetHandlers() map[string]event.Handler
}

// eventHandler provides event handlers for access module
type eventHandler struct {
	activity service.ActivityServiceInterface
}

// NewEventProvider creates new event handler provider
func NewEventProvider(svc *service.Service) EventHandlerInterface {
	return &eventHandler{
		activity: svc.Activity,
	}
}

// GetHandlers returns map of event handlers
func (e *eventHandler) GetHandlers() map[string]event.Handler {
	return map[string]event.Handler{
		// User events
		event.UserLogin:           e.handleUserEvent,
		event.UserCreated:         e.handleUserEvent,
		event.UserUpdated:         e.handleUserEvent,
		event.UserDeleted:         e.handleUserEvent,
		event.UserPasswordChanged: e.handleUserEvent,
		event.UserPasswordReset:   e.handleUserEvent,
		event.UserProfileUpdated:  e.handleUserEvent,
		event.UserStatusUpdated:   e.handleUserEvent,
		event.UserApiKeyGen:       e.handleUserEvent,
		event.UserApiKeyDel:       e.handleUserEvent,
		event.UserAuthCodeSent:    e.handleUserEvent,

		// System events
		event.SystemModified:  e.handleSystemEvent,
		event.SystemStarted:   e.handleSystemEvent,
		event.SystemStopped:   e.handleSystemEvent,
		event.SystemRestarted: e.handleSystemEvent,
		event.SystemUpgraded:  e.handleSystemEvent,

		// Security events
		event.SecurityIncident:  e.handleSecurityEvent,
		event.SecurityViolation: e.handleSecurityEvent,
		event.SecurityBlocked:   e.handleSecurityEvent,

		// Data events
		event.DataAccessed:   e.handleDataEvent,
		event.DataModified:   e.handleDataEvent,
		event.DataExported:   e.handleDataEvent,
		event.DataImported:   e.handleDataEvent,
		event.DataShared:     e.handleDataEvent,
		event.DataDownloaded: e.handleDataEvent,
	}
}

// handleUserEvent handles all user-related events
func (e *eventHandler) handleUserEvent(data any) {
	e.handleGenericEvent(data, event.CategoryUser)
}

// handleSystemEvent handles all system-related events
func (e *eventHandler) handleSystemEvent(data any) {
	e.handleGenericEvent(data, event.CategorySystem)
}

// handleSecurityEvent handles all security-related events
func (e *eventHandler) handleSecurityEvent(data any) {
	e.handleGenericEvent(data, event.CategorySecurity)
}

// handleDataEvent handles all data-related events
func (e *eventHandler) handleDataEvent(data any) {
	e.handleGenericEvent(data, event.CategoryData)
}

// handleGenericEvent handles events generically based on category
func (e *eventHandler) handleGenericEvent(data any, category string) {
	ctx := context.Background()

	payload, err := ext.ExtractEventPayload(data)
	if err != nil {
		logger.Errorf(ctx, "Failed to extract %s event payload: %v", category, err)
		return
	}

	userID := ext.SafeGet[string](payload, "user_id")
	details := ext.SafeGet[string](payload, "details")
	metadata := ext.SafeGet[types.JSON](payload, "metadata")

	// Extract activity type from event context or event name
	activityType := e.extractActivityType(ctx, category, *payload, data)

	if _, err := e.activity.LogActivity(ctx, userID, &structs.CreateActivityRequest{
		Type:     activityType,
		Details:  details,
		Metadata: &metadata,
	}); err != nil {
		logger.Errorf(ctx, "Failed to log %s activity: %v", category, err)
	}
}

// extractActivityType extracts activity type from event context or event name
func (e *eventHandler) extractActivityType(ctx context.Context, category string, payload types.JSON, originalData any) string {
	// Method 1: Try to get from event context in payload
	if eventCtx, ok := payload["event_context"].(map[string]any); ok {
		if eventName, exists := eventCtx["event_name"].(string); exists {
			return e.parseEventType(eventName, category)
		}
	}

	// Method 2: Try to extract from event name in payload directly
	if eventName, ok := payload["event_name"].(string); ok {
		return e.parseEventType(eventName, category)
	}

	// Method 3: Try to get event name from extension manager context (if available)
	if eventData, ok := originalData.(map[string]any); ok {
		if eventName, exists := eventData["_event_name"].(string); exists {
			return e.parseEventType(eventName, category)
		}
	}

	// Final fallback to category as type
	return category
}

// parseEventType parses event type from event name
func (e *eventHandler) parseEventType(eventName, category string) string {
	prefix := category + "."
	if strings.HasPrefix(eventName, prefix) && len(eventName) > len(prefix) {
		return eventName[len(prefix):]
	}
	return category
}
