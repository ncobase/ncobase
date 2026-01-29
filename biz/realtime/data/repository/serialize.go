package repository

import (
	"ncobase/biz/realtime/data/ent"
	"ncobase/biz/realtime/structs"
)

// SerializeNotification converts ent.Notification to structs.ReadNotification.
func SerializeNotification(n *ent.Notification) *structs.ReadNotification {
	if n == nil {
		return nil
	}
	return &structs.ReadNotification{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      n.Type,
		UserID:    n.UserID,
		Status:    n.Status,
		ChannelID: n.ChannelID,
		Links:     n.Links,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// SerializeNotifications converts []*ent.Notification to []*structs.ReadNotification.
func SerializeNotifications(notifications []*ent.Notification) []*structs.ReadNotification {
	result := make([]*structs.ReadNotification, len(notifications))
	for i, n := range notifications {
		result[i] = SerializeNotification(n)
	}
	return result
}

// SerializeChannel converts ent.RTChannel to structs.ReadChannel.
func SerializeChannel(ch *ent.RTChannel) *structs.ReadChannel {
	if ch == nil {
		return nil
	}
	return &structs.ReadChannel{
		ID:          ch.ID,
		Name:        ch.Name,
		Description: ch.Description,
		Type:        ch.Type,
		Status:      ch.Status,
		Extras:      ch.Extras,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
}

// SerializeChannels converts []*ent.RTChannel to []*structs.ReadChannel.
func SerializeChannels(channels []*ent.RTChannel) []*structs.ReadChannel {
	result := make([]*structs.ReadChannel, len(channels))
	for i, ch := range channels {
		result[i] = SerializeChannel(ch)
	}
	return result
}

// SerializeSubscription converts ent.Subscription to structs.ReadSubscription.
func SerializeSubscription(sub *ent.Subscription) *structs.ReadSubscription {
	if sub == nil {
		return nil
	}
	return &structs.ReadSubscription{
		ID:        sub.ID,
		UserID:    sub.UserID,
		ChannelID: sub.ChannelID,
		Status:    sub.Status,
		CreatedAt: sub.CreatedAt,
		UpdatedAt: sub.UpdatedAt,
	}
}

// SerializeEvent converts ent.Event to structs.ReadEvent.
func SerializeEvent(e *ent.Event) *structs.ReadEvent {
	if e == nil {
		return nil
	}
	result := &structs.ReadEvent{
		ID:         e.ID,
		Type:       e.Type,
		Source:     e.Source,
		Payload:    e.Payload,
		Status:     e.Status,
		Priority:   e.Priority,
		CreatedAt:  e.CreatedAt,
		RetryCount: e.RetryCount,
	}

	if e.ProcessedAt != 0 {
		result.ProcessedAt = &e.ProcessedAt
	}

	if e.ErrorMessage != "" {
		result.ErrorMessage = e.ErrorMessage
	}

	return result
}

// SerializeEvents converts []*ent.Event to []*structs.ReadEvent.
func SerializeEvents(events []*ent.Event) []*structs.ReadEvent {
	result := make([]*structs.ReadEvent, len(events))
	for i, e := range events {
		result[i] = SerializeEvent(e)
	}
	return result
}
