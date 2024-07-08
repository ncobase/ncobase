package feature

import (
	"sync"
)

// EventBus represents a simple event bus for inter-feature communication
type EventBus struct {
	subscribers map[string][]func(interface{})
	mu          sync.RWMutex
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]func(interface{})),
	}
}

// Subscribe adds a subscriber for a specific event
func (eb *EventBus) Subscribe(eventName string, handler func(interface{})) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(eventName string, data interface{}) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if handlers, ok := eb.subscribers[eventName]; ok {
		for _, handler := range handlers {
			go handler(data)
		}
	}
}
