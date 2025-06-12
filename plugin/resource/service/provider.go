package service

import (
	"ncobase/resource/data"
	"ncobase/resource/event"

	ext "github.com/ncobase/ncore/extension/types"
)

// Service contains all resource services
type Service struct {
	File  FileServiceInterface
	Batch BatchServiceInterface
	Quota QuotaServiceInterface
}

// New creates new resource service
func New(em ext.ManagerInterface, d *data.Data, publisher event.PublisherInterface) *Service {
	// Create image processor
	imageProcessor := NewImageProcessor()

	// Create quota service
	quotaConfig := &QuotaConfig{
		DefaultQuota:      10 * 1024 * 1024 * 1024, // 10GB default
		WarningThreshold:  0.8,                     // 80% warning
		EnableEnforcement: true,                    // Enforce quotas
		CheckInterval:     24 * 60 * 60,            // 24 hours in seconds
	}
	quotaService := NewQuotaService(d, publisher, quotaConfig, em)

	// Create file service
	fileService := NewFileService(d, imageProcessor, quotaService, publisher)

	// Create batch service
	batchService := NewBatchService(fileService, imageProcessor, publisher)

	return &Service{
		File:  fileService,
		Batch: batchService,
		Quota: quotaService,
	}
}

// RefreshDependencies refreshes external service dependencies
func (s *Service) RefreshDependencies() {
	if s.Quota != nil {
		s.Quota.RefreshSpaceServices()
	}
}
