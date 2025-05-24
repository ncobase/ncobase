package event

import (
	"context"
	"time"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/types"
)

// publisher implements PublisherInterface
type publisher struct {
	em ext.ManagerInterface
}

// NewPublisher creates new event publisher
func NewPublisher(em ext.ManagerInterface) PublisherInterface {
	return &publisher{em: em}
}

// PublishUserEvent publishes user-related events
func (p *publisher) PublishUserEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON) {
	p.publishEvent(ctx, CategoryUser+"."+eventType, userID, details, metadata)
}

// PublishSystemEvent publishes system-related events
func (p *publisher) PublishSystemEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON) {
	p.publishEvent(ctx, CategorySystem+"."+eventType, userID, details, metadata)
}

// PublishSecurityEvent publishes security-related events
func (p *publisher) PublishSecurityEvent(ctx context.Context, eventType, userID, details string, metadata *types.JSON) {
	p.publishEvent(ctx, CategorySecurity+"."+eventType, userID, details, metadata)
}

// PublishDataEvent publishes data-related events
func (p *publisher) PublishDataEvent(ctx context.Context, eventType, userID, resourceType, resourceID, details string, metadata *types.JSON) {
	if metadata == nil {
		metadata = &types.JSON{}
	}
	(*metadata)["resource_type"] = resourceType
	(*metadata)["resource_id"] = resourceID

	p.publishEvent(ctx, CategoryData+"."+eventType, userID, details, metadata)
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
	}

	p.em.PublishEvent(eventName, eventData)
}
