package event

import (
	"context"

	"github.com/ncobase/ncore/types"
)

// PublisherInterface defines an interface for publishing events
type PublisherInterface interface {
	PublishUserCreated(ctx context.Context, userID string, metadata *types.JSON)
	PublishUserUpdated(ctx context.Context, userID string, metadata *types.JSON)
	PublishUserDeleted(ctx context.Context, userID string, metadata *types.JSON)
	PublishPasswordChanged(ctx context.Context, userID string, metadata *types.JSON)
	PublishPasswordReset(ctx context.Context, userID string, metadata *types.JSON)
	PublishProfileUpdated(ctx context.Context, userID string, metadata *types.JSON)
	PublishStatusUpdated(ctx context.Context, userID string, metadata *types.JSON)
	PublishApiKeyGenerated(ctx context.Context, userID string, metadata *types.JSON)
	PublishApiKeyDeleted(ctx context.Context, userID string, metadata *types.JSON)
}
