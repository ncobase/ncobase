package handler

import (
	"ncobase/resource/service"
)

// Handler represents resource handler (HTTP handlers only)
type Handler struct {
	File  FileHandlerInterface
	Batch BatchHandlerInterface
	Quota QuotaHandlerInterface
}

// New creates new resource handler
func New(svc *service.Service) *Handler {
	return &Handler{
		File:  NewFileHandler(svc),
		Batch: NewBatchHandler(svc.File, svc.Batch),
		Quota: NewQuotaHandler(svc.Quota),
	}
}
