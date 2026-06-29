package service

import (
	"ncobase/plugin/resource/data"
	"ncobase/plugin/resource/event"
	"ncobase/plugin/resource/wrapper"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service contains all resource services
type Service struct {
	File    FileServiceInterface
	Batch   BatchServiceInterface
	Quota   QuotaServiceInterface
	Admin   AdminServiceInterface
	Space   *wrapper.SpaceServiceWrapper
	Content *wrapper.ContentServiceWrapper
	Config  ResourceConfigProvider
}

// New creates new resource service
func New(em ext.ManagerInterface, d *data.Data, publisher event.PublisherInterface, configProvider ResourceConfigProvider) *Service {
	if configProvider == nil {
		configProvider = NewDefaultConfigProvider()
	}

	// Create image processor
	imageProcessor := NewImageProcessor()

	// Create space service wrapper
	spaceWrapper := wrapper.NewSpaceServiceWrapper(em)

	// Create content service wrapper
	contentWrapper := wrapper.NewContentServiceWrapper(em)

	// Create quota service
	quotaService := NewQuotaService(d, publisher, configProvider, spaceWrapper)

	// Create file service
	fileService := NewFileService(d, imageProcessor, quotaService, publisher, configProvider, contentWrapper)

	// Create batch service
	batchService := NewBatchService(fileService, imageProcessor, publisher, configProvider)

	// Create admin service
	adminService := NewAdminService(d, quotaService, fileService)

	return &Service{
		File:    fileService,
		Batch:   batchService,
		Quota:   quotaService,
		Admin:   adminService,
		Space:   spaceWrapper,
		Content: contentWrapper,
		Config:  configProvider,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	if s.Quota != nil {
		s.Quota.RefreshSpaceServices()
	}
	if s.Space != nil {
		s.Space.RefreshServices()
	}
	if s.Content != nil {
		s.Content.RefreshServices()
	}
}
