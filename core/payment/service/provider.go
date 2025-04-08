package service

import (
	"ncobase/core/payment/data"
	"ncore/pkg/config"
)

// Service represents the payment service.
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
