package handler

import (
	"ncobase/plugin/resource/service"
)

// Handler represents the resource handler.
type Handler struct {
	Asset AssetHandlerInterface
}

// New creates new resource handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Asset: NewAssetHandler(svc),
	}
}
