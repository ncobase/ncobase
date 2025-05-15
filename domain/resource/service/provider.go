package service

import (
	"ncobase/resource/data"
	"ncobase/resource/event"
)

// Service is the struct for the resource service.
type Service struct {
	File  FileServiceInterface
	Batch BatchServiceInterface
	Quota QuotaServiceInterface
}

// New creates a new resource service.
func New(d *data.Data, publisher event.PublisherInterface) *Service {
	// Create image processor
	imageProcessor := NewImageProcessor()

	// Create quota service
	quotaConfig := &QuotaConfig{
		DefaultQuota:      10 * 1024 * 1024 * 1024, // 10GB default
		WarningThreshold:  0.8,                     // 80% warning
		EnableEnforcement: true,                    // Enforce quotas
		CheckInterval:     24 * 60 * 60,            // 24 hours in seconds
	}
	quotaService := NewQuotaService(d, publisher, quotaConfig)

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
