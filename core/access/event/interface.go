package event

import (
	"context"

	"github.com/ncobase/ncore/types"
)

// Handler defines event handler function
type Handler func(any)

// HandlerProvider defines interface for providing event handlers
type HandlerProvider interface {
	GetHandlers() map[string]Handler
}

// PublisherInterface defines interface for publishing events
type PublisherInterface interface {
	PublishUserEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON)
	PublishSystemEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON)
	PublishSecurityEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON)
	PublishDataEvent(ctx context.Context, eventType, userID, resourceType, resourceID, details string, metadata *types.JSON)
}
