package handler

import (
	"ncobase/plugin/resource/service"
)

// Handler represents resource handler
type Handler struct {
	File  FileHandlerInterface
	Batch BatchHandlerInterface
	Quota QuotaHandlerInterface
	Admin AdminHandlerInterface
}

// New creates new resource handler
func New(svc *service.Service) *Handler {
	return &Handler{
		File:  NewFileHandler(svc),
		Batch: NewBatchHandler(svc.File, svc.Batch),
		Quota: NewQuotaHandler(svc.Quota),
		Admin: NewAdminHandler(svc.Admin),
	}
}
