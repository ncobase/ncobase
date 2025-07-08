package wrapper

import (
	"context"
	"fmt"
	resourceStructs "ncobase/resource/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// ResourceFileServiceInterface defines resource file service interface for content module
type ResourceFileServiceInterface interface {
	Get(ctx context.Context, slug string) (*resourceStructs.ReadFile, error)
	List(ctx context.Context, params *resourceStructs.ListFileParams) (any, error)
	Delete(ctx context.Context, slug string) error
}

// ResourceServiceWrapper wraps resource service access with fallback behavior
type ResourceServiceWrapper struct {
	em          ext.ManagerInterface
	fileService ResourceFileServiceInterface
}

// NewResourceServiceWrapper creates new resource service wrapper
func NewResourceServiceWrapper(em ext.ManagerInterface) *ResourceServiceWrapper {
	wrapper := &ResourceServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads resource services using extension manager
func (w *ResourceServiceWrapper) loadServices() {
	// Try to get resource file service
	if fileSvc, err := w.em.GetCrossService("resource", "File"); err == nil {
		if service, ok := fileSvc.(ResourceFileServiceInterface); ok {
			w.fileService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *ResourceServiceWrapper) RefreshServices() {
	w.loadServices()
}

// GetFile gets file by slug
func (w *ResourceServiceWrapper) GetFile(ctx context.Context, slug string) (*resourceStructs.ReadFile, error) {
	if w.fileService != nil {
		return w.fileService.Get(ctx, slug)
	}
	return nil, fmt.Errorf("resource file service not available")
}

// ListFiles lists files
func (w *ResourceServiceWrapper) ListFiles(ctx context.Context, params *resourceStructs.ListFileParams) (any, error) {
	if w.fileService != nil {
		return w.fileService.List(ctx, params)
	}
	return nil, fmt.Errorf("resource file service not available")
}

// DeleteFile deletes file
func (w *ResourceServiceWrapper) DeleteFile(ctx context.Context, slug string) error {
	if w.fileService != nil {
		return w.fileService.Delete(ctx, slug)
	}
	return fmt.Errorf("resource file service not available")
}

// HasFileService checks if file service is available
func (w *ResourceServiceWrapper) HasFileService() bool {
	return w.fileService != nil
}
