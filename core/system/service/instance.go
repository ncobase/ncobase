package service

import "ncobase/core/system/data"

// InstanceServiceInterface is the interface for the service.
type InstanceServiceInterface interface {
}

// instanceService is the struct for the service.
type instanceService struct {
}

// NewInstanceService creates a new instance service.
func NewInstanceService(d *data.Data) InstanceServiceInterface {
	return &instanceService{}
}
