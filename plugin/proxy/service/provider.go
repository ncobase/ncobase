package service

import (
	"ncobase/proxy/data"
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
func New(d *data.Data) *Service {
	// Create the processor service
	processorSvc := NewProcessorService()

	return &Service{
		Endpoint:    NewEndpointService(d),
		Route:       NewRouteService(d),
		Transformer: NewTransformerService(d),
		Log:         NewLogService(d),
		Processor:   processorSvc,
	}
}
