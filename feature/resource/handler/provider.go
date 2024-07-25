package handler

import (
	"ncobase/feature/resource/service"
)

// Handler represents the resource handler.
type Handler struct {
	Attachment AttachmentHandlerInterface
}

// New creates new resource handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Attachment: NewAttachmentHandler(svc),
	}
}
