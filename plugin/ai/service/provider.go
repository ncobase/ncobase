package service

import (
	"ncobase/plugin/ai/data"
	"ncobase/plugin/ai/data/repository"
)

type Service struct {
	AI  AIServiceInterface
	Run RunServiceInterface
}

func New(d *data.Data, configProvider ConfigProvider) *Service {
	repos := repository.New(d)
	ai := NewAIService(repos.Run, configProvider)
	return &Service{
		AI:  ai,
		Run: NewRunService(repos.Run),
	}
}
