package repository

import (
	"ncobase/plugin/proxy/data"
)

// Repository represents the proxy repository.
type Repository struct {
	Endpoint    EndpointRepositoryInterface
	Route       RouteRepositoryInterface
	Transformer TransformerRepositoryInterface
	Log         LogRepositoryInterface
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		Endpoint:    NewEndpointRepository(d),
		Route:       NewRouteRepository(d),
		Transformer: NewTransformerRepository(d),
		Log:         NewLogRepository(d),
	}
}
