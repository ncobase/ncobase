package handler

import (
	"ncobase/domain/content/service"
)

// Handler represents the content handler.
type Handler struct {
	Taxonomy TaxonomyHandlerInterface
	Topic    TopicHandlerInterface
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Taxonomy: NewTaxonomyHandler(svc),
		Topic:    NewTopicHandler(svc),
	}
}
