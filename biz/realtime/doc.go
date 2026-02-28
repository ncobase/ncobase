// Package realtime provides real-time communication and event streaming for the Ncobase platform.
//
// This package implements WebSocket-based real-time communication, notifications, and
// event broadcasting. It enables live updates, push notifications, and real-time
// collaboration features.
//
// Key Features:
//   - WebSocket connection management
//   - Real-time event broadcasting
//   - Push notifications
//   - Channel-based messaging
//   - Presence tracking
//   - Event subscription and filtering
//   - Connection lifecycle management
//
// Main Components:
//   - Handler: WebSocket endpoints and HTTP notification APIs
//   - Service: Business logic for real-time operations
//   - Repository: Data access for notifications and events
//
// Real-time Capabilities:
//   - WebSocket: Bidirectional real-time communication
//   - Notifications: Push notifications to users
//   - Events: System-wide event broadcasting
//   - Channels: Topic-based message routing
//   - Presence: Online/offline status tracking
//
// Usage Example:
//
//	// Send notification to user
//	err := realtimeService.SendNotification(ctx, &structs.NotificationInput{
//	    UserID:  userID,
//	    Type:    "info",
//	    Title:   "New Message",
//	    Message: "You have a new message from John",
//	    Data: map[string]interface{}{
//	        "messageID": messageID,
//	    },
//	})
//	if err != nil {
//	    return err
//	}
//
//	// Broadcast event to channel
//	err = realtimeService.BroadcastEvent(ctx, &structs.EventInput{
//	    Channel: "space:123",
//	    Event:   "user.joined",
//	    Data: map[string]interface{}{
//	        "userID":   userID,
//	        "username": "john.doe",
//	    },
//	})
//
//	// Subscribe to channel
//	subscription, err := realtimeService.Subscribe(ctx, &structs.SubscribeInput{
//	    UserID:   userID,
//	    Channels: []string{"space:123", "notifications"},
//	})
//
// The package provides scalable real-time communication infrastructure suitable for
// collaborative applications, live dashboards, and notification systems.
package realtime
