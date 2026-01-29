package wrapper

import (
	"context"
	"fmt"
	resourceStructs "ncobase/plugin/resource/structs"

	"github.com/ncobase/ncore/data/paging"
	ext "github.com/ncobase/ncore/extension/types"
)

// ResourceFileServiceInterface defines file service interface for resource module
type ResourceFileServiceInterface interface {
	List(ctx context.Context, params *resourceStructs.ListFileParams) (paging.Result[*resourceStructs.ReadFile], error)
}

// ResourceFileWrapper wraps resource file service access with fallback behavior
type ResourceFileWrapper struct {
	em          ext.ManagerInterface
	fileService ResourceFileServiceInterface
}

// NewResourceFileWrapper creates a new resource file service wrapper
func NewResourceFileWrapper(em ext.ManagerInterface) *ResourceFileWrapper {
	wrapper := &ResourceFileWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

// loadServices loads resource file services using existing extension manager methods
func (w *ResourceFileWrapper) loadServices() {
	if fileSvc, err := w.em.GetCrossService("resource", "File"); err == nil {
		if service, ok := fileSvc.(ResourceFileServiceInterface); ok {
			w.fileService = service
		}
	}
}

// RefreshServices refreshes service references
func (w *ResourceFileWrapper) RefreshServices() {
	w.loadServices()
}

// ListFiles lists files via resource service
func (w *ResourceFileWrapper) ListFiles(ctx context.Context, params *resourceStructs.ListFileParams) (paging.Result[*resourceStructs.ReadFile], error) {
	if w.fileService != nil {
		return w.fileService.List(ctx, params)
	}
	return paging.Result[*resourceStructs.ReadFile]{Items: []*resourceStructs.ReadFile{}}, fmt.Errorf("resource file service not available")
}

// HasFileService checks if file service is available
func (w *ResourceFileWrapper) HasFileService() bool {
	return w.fileService != nil
}
