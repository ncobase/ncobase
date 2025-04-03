package templates

import "fmt"

func ServiceTemplate(name, extType string) string {
	return fmt.Sprintf(`package service

import (
	"ncobase/ncore/config"
	"ncobase/%s/%s/data"
)

// Service represents the %s service.
type Service struct {
	// Add your service fields here
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		// Initialize your service fields here
	}
}

// Add your service methods here
`, extType, name, name)
}
