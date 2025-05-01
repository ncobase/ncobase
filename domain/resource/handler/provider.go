package handler

import (
	"ncobase/domain/resource/service"
)

// Handler represents the resource handler.
type Handler struct {
	File  FileHandlerInterface
	Batch BatchHandlerInterface
	Quota QuotaHandlerInterface
}

// New creates new resource handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		File:  NewFileHandler(svc),
		Batch: NewBatchHandler(svc.File, svc.Batch),
		Quota: NewQuotaHandler(svc.Quota),
	}
}
