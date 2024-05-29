package server

import (
	"stocms/internal/config"
	"stocms/internal/data"
	"stocms/internal/handler"
	"stocms/internal/service"
)

// initialize initializes the database, services, and handlers.
func initialize(cfg *config.Config) (*handler.Handler, *service.Service, func(), error) {
	// Initialize database
	d, cleanup, err := data.New(&cfg.Data)
	if err != nil {
		return nil, nil, nil, err
	}

	// Initialize services
	svc := service.New(cfg, d)

	// Initialize handlers
	h := handler.New(svc)

	return h, svc, cleanup, nil
}
