package handler

import (
	"ncobase/feature/content/handler/taxonomy"
	"ncobase/feature/content/handler/topic"
	"ncobase/feature/content/service"
)

type Handler struct {
	Taxonomy taxonomy.HandlerInterface
	Topic    topic.HandlerInterface
}

func New(svc *service.Service) *Handler {
	return &Handler{
		Taxonomy: taxonomy.New(svc),
		Topic:    topic.New(svc),
	}
}
