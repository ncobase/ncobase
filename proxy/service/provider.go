package service

import (
	"ncobase/proxy/data"

	"github.com/ncobase/ncore/config"
)

// Service represents the proxy service.
type Service struct {
	Endpoint    EndpointServiceInterface
	Route       RouteServiceInterface
	Transformer TransformerServiceInterface
	Log         LogServiceInterface
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		Endpoint:    NewEndpointService(d),
		Route:       NewRouteService(d),
		Transformer: NewTransformerService(d),
		Log:         NewLogService(d),
	}
}
