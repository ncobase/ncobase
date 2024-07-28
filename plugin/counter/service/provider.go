package service

import "ncobase/plugin/counter/data"

// Service represents the counter service.
type Service struct {
}

// New creates a new service.
func New(d *data.Data) *Service {
	return &Service{}
}
