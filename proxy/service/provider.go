package service

import (
	"ncobase/proxy/data"

	accessService "ncobase/core/access/service"
	spaceService "ncobase/core/space/service"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"

	"github.com/ncobase/ncore/config"
)

// Service represents the proxy service.
type Service struct {
	Endpoint    EndpointServiceInterface
	Route       RouteServiceInterface
	Transformer TransformerServiceInterface
	Log         LogServiceInterface
	Processor   ProcessorServiceInterface
}

// New creates a new service.
func New(conf *config.Config, d *data.Data,
	us *userService.Service,
	ts *tenantService.Service,
	ss *spaceService.Service,
	as *accessService.Service) *Service {

	// Create the processor service with injected internal services
	processorSvc := NewProcessorService(us, ts, ss, as)

	return &Service{
		Endpoint:    NewEndpointService(d),
		Route:       NewRouteService(d),
		Transformer: NewTransformerService(d),
		Log:         NewLogService(d),
		Processor:   processorSvc,
	}
}
