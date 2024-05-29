package service

import (
	"stocms/internal/config"
	"stocms/internal/data"
)

// Service represents a service definition.
type Service struct {
	conf   *config.Config
	sample data.ISample
}

// New creates a Service instance and returns it.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		conf:   conf,
		sample: data.NewSample(d),
	}
}
