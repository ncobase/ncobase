package service

import (
	"context"
	"stocms/internal/config"
	"stocms/internal/data"
	"stocms/internal/data/repository"
)

// Service represents a service definition.
type Service struct {
	conf   *config.Config
	d      *data.Data
	domain repo.Domain
	user   repo.User
}

// New creates a Service instance and returns it.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		conf:   conf,
		d:      d,
		domain: repo.NewDomain(d),
		user:   repo.NewUser(d),
	}
}

// Ping check server
func (svc *Service) Ping(ctx context.Context) error {
	return svc.d.Ping(ctx)
}
