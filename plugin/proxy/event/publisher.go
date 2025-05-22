package event

import (
	"time"

	ext "github.com/ncobase/ncore/extension/types"
)

// Event names for proxy module
const (
	// Request events

	EventRequestReceived     = "proxy.request.received"
	EventRequestPreProcessed = "proxy.request.preprocessed"
	EventRequestTransformed  = "proxy.request.transformed"
	EventRequestSent         = "proxy.request.sent"

	// Response events

	EventResponseReceived      = "proxy.response.received"
	EventResponseTransformed   = "proxy.response.transformed"
	EventResponsePostProcessed = "proxy.response.postprocessed"
	EventResponseSent          = "proxy.response.sent"

	// Error events

	EventRequestError          = "proxy.request.error"
	EventResponseError         = "proxy.response.error"
	EventCircuitBreakerTripped = "proxy.circuit_breaker.tripped"
	EventCircuitBreakerReset   = "proxy.circuit_breaker.reset"
)

// ProxyEventData represents event data specific to proxy operations
type ProxyEventData struct {
	Timestamp   time.Time      `json:"timestamp"`
	EndpointID  string         `json:"endpoint_id"`
	EndpointURL string         `json:"endpoint_url"`
	RouteID     string         `json:"route_id"`
	RoutePath   string         `json:"route_path"`
	Method      string         `json:"method"`
	StatusCode  int            `json:"status_code,omitempty"`
	Duration    int            `json:"duration,omitempty"`
	Error       string         `json:"error,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// Publisher provides functionality for publishing proxy events
type Publisher struct {
	manager ext.ManagerInterface
}

// NewPublisher creates a new event publisher
func NewPublisher(manager ext.ManagerInterface) *Publisher {
	return &Publisher{
		manager: manager,
	}
}

// Publish publishes an event with the given name and data
func (p *Publisher) Publish(eventName string, eventData *ProxyEventData) {
	if p.manager == nil || eventData == nil {
		return
	}

	// Add default timestamp if not set
	if eventData.Timestamp.IsZero() {
		eventData.Timestamp = time.Now()
	}

	// Publish the event
	p.manager.PublishEvent(eventName, eventData)
}
