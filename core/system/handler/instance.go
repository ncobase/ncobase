package handler

import "ncobase/core/system/service"

// InstanceHandlerInterface represents the instance handler interface.
type InstanceHandlerInterface any

// instanceHandler represents the instance handler.
type instanceHandler struct {
	s *service.Service
}

// NewInstanceHandler creates new instance handler.
func NewInstanceHandler(svc *service.Service) InstanceHandlerInterface {
	return &instanceHandler{
		s: svc,
	}
}
