package handler

import (
	"ncobase/feature/asset/handler/asset"
	"ncobase/feature/asset/service"
)

type Handler struct {
	Asset asset.HandlerInterface
}

func New(svc *service.Service) *Handler {
	return &Handler{
		Asset: asset.New(svc),
	}
}
