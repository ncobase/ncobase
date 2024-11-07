package templates

import "fmt"

func HandlerTemplate(name, extType string) string {
	return fmt.Sprintf(`package handler

import "ncobase/%s/%s/service"

// Handler represents the %s handler.
type Handler struct {
	// Add your handler fields here
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		// Initialize your handler fields here
	}
}

// Add your handler methods here
`, extType, name, name)
}
