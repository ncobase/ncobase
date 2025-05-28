package event

import (
	"context"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/types"
)

// Event types for auth module
const (
	UserLogin            = "user.login"
	UserCreated          = "user.created"
	UserLogout           = "user.logout"
	UserPasswordChanged  = "user.password_changed"
	UserPasswordReset    = "user.password_reset"
	UserProfileUpdated   = "user.profile_updated"
	UserStatusUpdated    = "user.status_updated"
	UserApiKeyGen        = "user.apikey_generated"
	UserApiKeyDel        = "user.apikey_deleted"
	UserTokenRefresh     = "user.token_refreshed"
	UserAuthCodeSent     = "user.auth_code_sent"
	UserSessionCreated   = "user.session_created"
	UserSessionDestroyed = "user.session_destroyed"
	UserSessionExpired   = "user.session_expired"
)

// publisher implements PublisherInterface
type publisher struct {
	em ext.ManagerInterface
}

// NewPublisher creates new event publisher
func NewPublisher(em ext.ManagerInterface) PublisherInterface {
	return &publisher{em: em}
}

// PublishUserLogin publishes user login event
func (p *publisher) PublishUserLogin(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserLogin, userID, "User logged in", metadata)
}

// PublishUserCreated publishes user created event
func (p *publisher) PublishUserCreated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserCreated, userID, "User account created", metadata)
}

// PublishUserLogout publishes user logout event
func (p *publisher) PublishUserLogout(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserLogout, userID, "User logged out", metadata)
}

// PublishPasswordChanged publishes password changed event
func (p *publisher) PublishPasswordChanged(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserPasswordChanged, userID, "User password changed", metadata)
}

// PublishPasswordReset publishes password reset event
func (p *publisher) PublishPasswordReset(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserPasswordReset, userID, "User password reset", metadata)
}

// PublishProfileUpdated publishes profile updated event
func (p *publisher) PublishProfileUpdated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserProfileUpdated, userID, "User profile updated", metadata)
}

// PublishStatusUpdated publishes status updated event
func (p *publisher) PublishStatusUpdated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserStatusUpdated, userID, "User status updated", metadata)
}

// PublishApiKeyGenerated publishes API key generated event
func (p *publisher) PublishApiKeyGenerated(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserApiKeyGen, userID, "API key generated", metadata)
}

// PublishApiKeyDeleted publishes API key deleted event
func (p *publisher) PublishApiKeyDeleted(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserApiKeyDel, userID, "API key deleted", metadata)
}

// PublishTokenRefreshed publishes token refreshed event
func (p *publisher) PublishTokenRefreshed(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserTokenRefresh, userID, "Access token refreshed", metadata)
}

// PublishAuthCodeSent publishes auth code sent event
func (p *publisher) PublishAuthCodeSent(ctx context.Context, userID string, metadata *types.JSON) {
	p.publishEvent(ctx, UserAuthCodeSent, userID, "Auth code sent", metadata)
}

// PublishSessionCreated publishes session created event
func (p *publisher) PublishSessionCreated(ctx context.Context, userID, sessionID string, metadata *types.JSON) {
	eventData := &types.JSON{
		"user_id":    userID,
		"session_id": sessionID,
		"details":    "User session created",
		"timestamp":  time.Now().UnixMilli(),
		"metadata":   metadata,
		"event_context": map[string]any{
			"event_name": UserSessionCreated,
		},
	}
	p.em.PublishEvent(UserSessionCreated, eventData)
}

// PublishSessionDestroyed publishes session destroyed event
func (p *publisher) PublishSessionDestroyed(ctx context.Context, userID, sessionID string, metadata *types.JSON) {
	eventData := &types.JSON{
		"user_id":    userID,
		"session_id": sessionID,
		"details":    "User session destroyed",
		"timestamp":  time.Now().UnixMilli(),
		"metadata":   metadata,
		"event_context": map[string]any{
			"event_name": UserSessionDestroyed,
		},
	}
	p.em.PublishEvent(UserSessionDestroyed, eventData)
}

// publishEvent is helper method to publish events
func (p *publisher) publishEvent(_ context.Context, eventName, userID, details string, metadata *types.JSON) {
	if p.em == nil {
		return
	}

	eventData := &types.JSON{
		"user_id":   userID,
		"details":   details,
		"timestamp": time.Now().UnixMilli(),
		"metadata":  metadata,
		"event_context": map[string]any{
			"event_name": eventName,
		},
	}

	p.em.PublishEvent(eventName, eventData)
}
