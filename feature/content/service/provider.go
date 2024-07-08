package service

import (
	"ncobase/feature/content/data"
	"ncobase/feature/content/service/taxonomy"
	"ncobase/feature/content/service/topic"
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
