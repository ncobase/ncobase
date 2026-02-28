package service

import (
	"ncobase/plugin/sample/data/repository"
)

// Service aggregates all services for the sample plugin
type Service struct {
	Sample SampleServiceInterface
}

// NewService creates a new service instance
func NewService(repo *repository.Repository) *Service {
	return &Service{
		Sample: NewSampleService(repo.Sample),
	}
}
