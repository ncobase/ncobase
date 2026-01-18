package handler

import (
	"ncobase/biz/content/service"
)

// Handler represents the content handler.
type Handler struct {
	Taxonomy     TaxonomyHandlerInterface
	Topic        TopicHandlerInterface
	Channel      ChannelHandlerInterface
	Distribution DistributionHandlerInterface
	Media        MediaHandlerInterface
	TopicMedia   TopicMediaHandlerInterface
}

// New creates a new handler.
func New(svc *service.Service) *Handler {
	return &Handler{
		Taxonomy:     NewTaxonomyHandler(svc),
		Topic:        NewTopicHandler(svc),
		Channel:      NewChannelHandler(svc),
		Distribution: NewDistributionHandler(svc),
		Media:        NewMediaHandler(svc),
		TopicMedia:   NewTopicMediaHandler(svc),
	}
}
