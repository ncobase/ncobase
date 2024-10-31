package service

import (
	"ncobase/domain/resource/data"
)

// Service is the struct for the resource service.
type Service struct {
	Attachment AttachmentServiceInterface
}

// New creates a new resource service.
func New(d *data.Data) *Service {
	return &Service{
		Attachment: NewAttachmentService(d),
	}
}
