package event

// Handler defines a generic event handler function
type Handler func(any)

// HandlerProvider defines an interface for providing event handlers
type HandlerProvider interface {
	// GetHandlers returns a map of event name to handler functions
	GetHandlers() map[string]Handler
}
