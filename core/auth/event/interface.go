package event

import (
	"context"

	"github.com/ncobase/ncore/types"
)

// PublisherInterface defines interface for publishing events
type PublisherInterface interface {
	PublishUserLogin(ctx context.Context, userID string, metadata *types.JSON)
	PublishUserCreated(ctx context.Context, userID string, metadata *types.JSON)
	PublishUserLogout(ctx context.Context, userID string, metadata *types.JSON)
	PublishPasswordChanged(ctx context.Context, userID string, metadata *types.JSON)
	PublishPasswordReset(ctx context.Context, userID string, metadata *types.JSON)
	PublishProfileUpdated(ctx context.Context, userID string, metadata *types.JSON)
	PublishStatusUpdated(ctx context.Context, userID string, metadata *types.JSON)
	PublishApiKeyGenerated(ctx context.Context, userID string, metadata *types.JSON)
	PublishApiKeyDeleted(ctx context.Context, userID string, metadata *types.JSON)
	PublishTokenRefreshed(ctx context.Context, userID string, metadata *types.JSON)
	PublishAuthCodeSent(ctx context.Context, userID string, metadata *types.JSON)
	PublishSessionCreated(ctx context.Context, userID, sessionID string, metadata *types.JSON)
	PublishSessionDestroyed(ctx context.Context, userID, sessionID string, metadata *types.JSON)
}
