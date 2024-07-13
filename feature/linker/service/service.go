package service

import "ncobase/feature/linker/data"

// Service is the struct for the relationship service.
type Service struct {
	d *data.Data
}

// New creates a new relationship service.
func New(d *data.Data) *Service {
	return &Service{
		d: d,
	}
}
