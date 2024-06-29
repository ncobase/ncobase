package service

import (
	"ncobase/plugin/content/data"
	"ncobase/plugin/content/service/taxonomy"
	"ncobase/plugin/content/service/topic"
)

type Service struct {
	Taxonomy taxonomy.ServiceInterface
	Topic    topic.ServiceInterface
}

func New(d *data.Data) *Service {
	return &Service{
		Taxonomy: taxonomy.New(d),
		Topic:    topic.New(d),
	}
}
