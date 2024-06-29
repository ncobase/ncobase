package handler

import (
	"ncobase/plugin/asset/handler/asset"
	"ncobase/plugin/asset/service"
)

type Handler struct {
	Asset asset.HandlerInterface
}

func New(svc *service.Service) *Handler {
	return &Handler{
		Asset: asset.New(svc),
	}
}
