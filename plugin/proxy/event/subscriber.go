package event

import (
	"context"
	orgStructs "ncobase/core/organization/structs"
	spaceStructs "ncobase/core/space/structs"
	userStructs "ncobase/core/user/structs"
	"ncobase/plugin/proxy/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// Subscriber handles event subscriptions and processing for the proxy module
type Subscriber struct {
	publisher *Publisher

	usw *wrapper.UserServiceWrapper
	tsw *wrapper.SpaceServiceWrapper
	ssw *wrapper.OrganizationServiceWrapper
	asw *wrapper.AccessServiceWrapper
}

// NewSubscriber creates a new event subscriber
func NewSubscriber(
	em ext.ManagerInterface,
	publisher *Publisher,
) *Subscriber {
	usw := wrapper.NewUserServiceWrapper(em)
	tsw := wrapper.NewSpaceServiceWrapper(em)
	ssw := wrapper.NewOrganizationServiceWrapper(em)
	asw := wrapper.NewAccessServiceWrapper(em)
	return &Subscriber{
		publisher: publisher,
		usw:       usw,
		tsw:       tsw,
		ssw:       ssw,
		asw:       asw,
	}
}

// Initialize sets up event subscriptions with the provided manager
func (s *Subscriber) Initialize(manager ext.ManagerInterface) {
	if manager == nil {
		logger.Warn(context.Background(), "Proxy Event support is disabled: no extension manager provided")
		return
	}

	// Subscribe to events
	manager.SubscribeEvent(EventResponseReceived, s.handleResponseReceived)
	manager.SubscribeEvent(EventRequestError, s.handleRequestError)
	manager.SubscribeEvent(EventCircuitBreakerTripped, s.handleCircuitBreakerTripped)
}

// RefreshDependencies refreshes the dependencies of the subscriber
func (s *Subscriber) RefreshDependencies() {
	s.usw.RefreshServices()
	s.tsw.RefreshServices()
	s.ssw.RefreshServices()
	s.asw.RefreshServices()
}

// handleResponseReceived processes response data and potentially triggers additional events
func (s *Subscriber) handleResponseReceived(data any) {
	eventData, ok := data.(ext.EventData)
	if !ok {
		logger.Error(context.Background(), "Invalid event data format")
		return
	}

	proxyData, ok := eventData.Data.(*ProxyEventData)
	if !ok {
		logger.Error(context.Background(), "Invalid proxy event data format")
		return
	}

	// Log the event for debugging
	logger.Debugf(context.Background(), "Received response from %s endpoint (route: %s)",
		proxyData.EndpointURL, proxyData.RoutePath)

	// Process based on status code
	if proxyData.StatusCode >= 200 && proxyData.StatusCode < 300 {
		// Successful response, might trigger synchronization with internal services
		s.processSuccessfulResponse(context.Background(), proxyData)
	} else if proxyData.StatusCode >= 400 {
		// Error response, might need error handling or retry logic
		s.processErrorResponse(context.Background(), proxyData)
	}
}

// handleRequestError handles errors that occurred during proxy requests
func (s *Subscriber) handleRequestError(data any) {
	eventData, ok := data.(ext.EventData)
	if !ok {
		logger.Error(context.Background(), "Invalid event data format")
		return
	}

	proxyData, ok := eventData.Data.(*ProxyEventData)
	if !ok {
		logger.Error(context.Background(), "Invalid proxy event data format")
		return
	}

	logger.Warnf(context.Background(), "Error in proxy request to %s: %s",
		proxyData.EndpointURL, proxyData.Error)

	// Notify administrators or trigger fallback mechanisms
	s.notifyErrorHandlers(context.Background(), proxyData)
}

// handleCircuitBreakerTripped handles circuit breaker events
func (s *Subscriber) handleCircuitBreakerTripped(data any) {
	eventData, ok := data.(ext.EventData)
	if !ok {
		logger.Error(context.Background(), "Invalid event data format")
		return
	}

	proxyData, ok := eventData.Data.(*ProxyEventData)
	if !ok {
		logger.Error(context.Background(), "Invalid proxy event data format")
		return
	}

	logger.Warnf(context.Background(), "Circuit breaker tripped for endpoint %s",
		proxyData.EndpointURL)

	// Notify administrators or trigger service degradation modes
	s.handleServiceDegradation(context.Background(), proxyData)
}

// processSuccessfulResponse handles successful responses that might need synchronization
func (s *Subscriber) processSuccessfulResponse(ctx context.Context, data *ProxyEventData) {
	// Example: If this was a CRM contact update, synchronize with user service
	if data.Metadata != nil {
		if entityType, ok := data.Metadata["entity_type"].(string); ok {
			switch entityType {
			case "contact":
				s.syncContactWithUserService(ctx, data)
			case "payment":
				s.processPaymentUpdate(ctx, data)
			case "message":
				s.notifyMessageRecipients(ctx, data)
			}
		}
	}
}

// processErrorResponse handles error responses that might need intervention
func (s *Subscriber) processErrorResponse(ctx context.Context, data *ProxyEventData) {
	// Determine if we need to retry or notify administrators
	if data.StatusCode >= 500 {
		// Server error, might be temporary
		logger.Warnf(ctx, "Server error from %s (status: %d), considering retry",
			data.EndpointURL, data.StatusCode)

		// Could schedule a retry event here
	} else if data.StatusCode >= 400 && data.StatusCode < 500 {
		// Client error, might need configuration update
		logger.Warnf(ctx, "Client error to %s (status: %d), may need configuration update",
			data.EndpointURL, data.StatusCode)

		// Notify administrators
	}
}

// notifyErrorHandlers sends notifications about proxy errors
func (s *Subscriber) notifyErrorHandlers(ctx context.Context, data *ProxyEventData) {
	// In a real implementation, this might:
	// 1. Send an email/Slack notification
	// 2. Create an incident in an incident management system
	// 3. Log to a specialized error tracking service

	logger.Infof(ctx, "Error notification sent for endpoint %s: %s",
		data.EndpointURL, data.Error)
}

// handleServiceDegradation manages service degradation when circuit breakers trip
func (s *Subscriber) handleServiceDegradation(ctx context.Context, data *ProxyEventData) {
	// In a real implementation, this might:
	// 1. Switch to a backup endpoint
	// 2. Enable fallback mode using cached data
	// 3. Update a status dashboard

	logger.Infof(ctx, "Service degradation handling for endpoint %s", data.EndpointURL)
}

// syncContactWithUserService synchronizes contact data with user service
func (s *Subscriber) syncContactWithUserService(ctx context.Context, data *ProxyEventData) {
	if data.Metadata == nil {
		return
	}

	contactData, ok := data.Metadata["contact_data"].(map[string]any)
	if !ok {
		logger.Warnf(ctx, "Missing contact data in metadata")
		return
	}

	// Extract email to find matching user
	email, ok := contactData["email"].(string)
	if !ok || email == "" {
		logger.Warnf(ctx, "Missing email in contact data")
		return
	}

	// Find user by email
	user, err := s.usw.FindUser(ctx, &userStructs.FindUser{Email: email})
	if err != nil {
		logger.Warnf(ctx, "Failed to find user with email %s: %v", email, err)
		return
	}

	// Update user data with contact information
	// This is a simplified example - real implementation would map fields appropriately
	updates := make(map[string]any)

	if name, ok := contactData["full_name"].(string); ok && name != "" {
		updates["name"] = name
	}

	if phone, ok := contactData["phone"].(string); ok && phone != "" {
		updates["phone"] = phone
	}

	// Apply updates to user
	if len(updates) > 0 {
		_, err := s.usw.UpdateUser(ctx, user.ID, updates)
		if err != nil {
			logger.Errorf(ctx, "Failed to update user with contact data: %v", err)
			return
		}
		logger.Infof(ctx, "User %s synchronized with contact data from CRM", user.ID)
	}
}

// processPaymentUpdate processes payment updates
func (s *Subscriber) processPaymentUpdate(ctx context.Context, data *ProxyEventData) {
	if data.Metadata == nil {
		return
	}

	paymentData, ok := data.Metadata["payment_data"].(map[string]any)
	if !ok {
		logger.Warnf(ctx, "Missing payment data in metadata")
		return
	}

	// Extract space ID to find matching space
	spaceID, ok := paymentData["space_id"].(string)
	if !ok || spaceID == "" {
		logger.Warnf(ctx, "Missing space ID in payment data")
		return
	}

	// Get payment status
	status, ok := paymentData["status"].(string)
	if !ok {
		logger.Warnf(ctx, "Missing payment status in payment data")
		return
	}

	// Update space billing status based on payment result
	var billingStatus string
	switch status {
	case "succeeded":
		billingStatus = "active"
	case "failed":
		billingStatus = "failed"
	case "pending":
		billingStatus = "pending"
	default:
		billingStatus = "unknown"
	}

	// Update space
	extras := &map[string]any{
		"billing_status": billingStatus,
	}

	if amount, ok := paymentData["amount"].(float64); ok {
		(*extras)["last_payment_amount"] = amount
	}

	if paymentID, ok := paymentData["payment_id"].(string); ok {
		(*extras)["last_payment_id"] = paymentID
	}

	_, err := s.tsw.UpdateSpace(ctx, &spaceStructs.UpdateSpaceBody{ID: spaceID, SpaceBody: spaceStructs.SpaceBody{Extras: extras}})
	if err != nil {
		logger.Errorf(ctx, "Failed to update space billing status: %v", err)
		return
	}

	logger.Infof(ctx, "Space %s billing status updated to %s", spaceID, billingStatus)
}

// notifyMessageRecipients notifies message recipients
func (s *Subscriber) notifyMessageRecipients(ctx context.Context, data *ProxyEventData) {
	if data.Metadata == nil {
		return
	}

	messageData, ok := data.Metadata["message_data"].(map[string]any)
	if !ok {
		logger.Warnf(ctx, "Missing message data in metadata")
		return
	}

	// Extract group ID
	orgID, ok := messageData["org_id"].(string)
	if !ok || orgID == "" {
		logger.Warnf(ctx, "Missing group ID in message data")
		return
	}

	// Get the group to find members
	_, err := s.ssw.GetOrganization(ctx, &orgStructs.FindOrganization{Organization: orgID})
	if err != nil {
		logger.Errorf(ctx, "Failed to find group %s: %v", orgID, err)
		return
	}

	// Get message content
	content, ok := messageData["content"].(string)
	if !ok || content == "" {
		logger.Warnf(ctx, "Missing message content")
		return
	}

	// Get message sender
	senderID, ok := messageData["sender_id"].(string)
	if !ok || senderID == "" {
		logger.Warnf(ctx, "Missing sender ID")
		return
	}

	// Notify group members about the new message
	// In a real implementation, this would use a notification service
	logger.Infof(ctx, "New message in group %s from user %s: %s",
		orgID, senderID, content[:min(len(content), 30)])
}
