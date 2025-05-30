package event

import (
	"context"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/types"
)

// publisher represents an event publisher
type publisher struct {
	em ext.ManagerInterface
}

// NewPublisher creates a new event publisher
func NewPublisher(em ext.ManagerInterface) PublisherInterface {
	return &publisher{
		em: em,
	}
}

// PublishUserCreated publishes a user created event
func (p *publisher) PublishUserCreated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.created", userID, "User created", metadata)
}

// PublishUserUpdated publishes a user updated event
func (p *publisher) PublishUserUpdated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.updated", userID, "User updated", metadata)
}

// PublishUserDeleted publishes a user deleted event
func (p *publisher) PublishUserDeleted(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.deleted", userID, "User deleted", metadata)
}

// PublishPasswordChanged publishes a password changed event
func (p *publisher) PublishPasswordChanged(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.password_changed", userID, "Password changed", metadata)
}

// PublishPasswordReset publishes a password reset event
func (p *publisher) PublishPasswordReset(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.password_reset", userID, "Password reset", metadata)
}

// PublishProfileUpdated publishes a profile updated event
func (p *publisher) PublishProfileUpdated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.profile_updated", userID, "Profile updated", metadata)
}

// PublishStatusUpdated publishes a status updated event
func (p *publisher) PublishStatusUpdated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.status_updated", userID, "Status updated", metadata)
}

// PublishApiKeyGenerated publishes an API key generated event
func (p *publisher) PublishApiKeyGenerated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.apikey_generated", userID, "API key generated", metadata)
}

// PublishApiKeyDeleted publishes an API key deleted event
func (p *publisher) PublishApiKeyDeleted(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, "user.apikey_deleted", userID, "API key deleted", metadata)
}

// publishEvent is a helper method to publish events
func (p *publisher) publishEvent(_ context.Context, eventType, userID, details string, metadata *types.JSON) {
	// Create event data
	eventData := &types.JSON{
		"user_id":   userID,
		"timestamp": time.Now().UnixMilli(),
		"details":   details,
		"metadata":  metadata,
	}

	// Publish event
	if p.em != nil {
		p.em.PublishEvent(eventType, eventData)
	}
}
