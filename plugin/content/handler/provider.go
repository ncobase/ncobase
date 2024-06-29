package handler

import (
	"ncobase/plugin/content/handler/taxonomy"
	"ncobase/plugin/content/handler/topic"
	"ncobase/plugin/content/service"
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
