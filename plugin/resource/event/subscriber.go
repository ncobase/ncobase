package event

import (
	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// SubscriberInterface defines event subscription methods
type SubscriberInterface interface {
	Subscribe(em ext.ManagerInterface)
	Unsubscribe(em ext.ManagerInterface)
	SetQuotaUpdater(updater QuotaUpdaterInterface)
}

// subscriber manages event subscriptions for the resource plugin
type subscriber struct {
	handler HandlerInterface
}

// NewSubscriber creates new event subscriber.
func NewSubscriber(em ext.ManagerInterface, notifier NotifierInterface) SubscriberInterface {
	return &subscriber{
		handler: NewHandler(em, notifier),
	}
}

// SetQuotaUpdater sets the quota updater for the event handler
func (s *subscriber) SetQuotaUpdater(updater QuotaUpdaterInterface) {
	if s.handler != nil {
		s.handler.SetQuotaUpdater(updater)
	}
}

// Subscribe subscribes to all relevant events
func (s *subscriber) Subscribe(em ext.ManagerInterface) {
	if em == nil || s.handler == nil {
		logger.Warn(nil, "Cannot subscribe to events: missing extension manager or handler")
		return
	}

	// Subscribe to file events
	em.SubscribeEvent(FileCreated, s.handler.HandleFileCreated)
	em.SubscribeEvent(FileUpdated, s.handler.HandleFileUpdated)
	em.SubscribeEvent(FileDeleted, s.handler.HandleFileDeleted)
	em.SubscribeEvent(FileAccessed, s.handler.HandleFileAccessed)

	// Subscribe to quota events
	em.SubscribeEvent(StorageQuotaWarning, s.handler.HandleQuotaWarning)
	em.SubscribeEvent(StorageQuotaExceeded, s.handler.HandleQuotaExceeded)

	// Subscribe to batch operation events
	em.SubscribeEvent(BatchUploadStarted, s.handler.HandleBatchUploadStarted)
	em.SubscribeEvent(BatchUploadComplete, s.handler.HandleBatchUploadComplete)
	em.SubscribeEvent(BatchUploadFailed, s.handler.HandleBatchUploadFailed)
}

// Unsubscribe unsubscribes from all events
func (s *subscriber) Unsubscribe(em ext.ManagerInterface) {
	if em == nil {
		return
	}

}
