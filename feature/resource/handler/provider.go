package handler

import (
	"ncobase/feature/resource/handler/asset"
	"ncobase/feature/resource/service"
)

type Handler struct {
	Asset asset.HandlerInterface
}

func New(svc *service.Service) *Handler {
	return &Handler{
		Asset: asset.New(svc),
	}
}
