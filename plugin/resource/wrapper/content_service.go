package wrapper

import (
	"context"
	"fmt"
	contentStructs "ncobase/biz/content/structs"

	"github.com/ncobase/ncore/data/paging"
	ext "github.com/ncobase/ncore/extension/types"
)

// ContentMediaServiceInterface defines the content media service methods needed by resources.
type ContentMediaServiceInterface interface {
	List(ctx context.Context, params *contentStructs.ListMediaParams) (paging.Result[*contentStructs.ReadMedia], error)
}

// ContentTopicMediaServiceInterface defines the content topic-media methods needed by resources.
type ContentTopicMediaServiceInterface interface {
	List(ctx context.Context, params *contentStructs.ListTopicMediaParams) (paging.Result[*contentStructs.ReadTopicMedia], error)
}

// ContentTopicServiceInterface defines the content topic methods needed by resources.
type ContentTopicServiceInterface interface {
	GetByID(ctx context.Context, id string) (*contentStructs.ReadTopic, error)
}

// ContentServiceWrapper resolves content services through the extension manager.
type ContentServiceWrapper struct {
	em                ext.ManagerInterface
	mediaService      ContentMediaServiceInterface
	topicMediaService ContentTopicMediaServiceInterface
	topicService      ContentTopicServiceInterface
}

// NewContentServiceWrapper creates a content service wrapper.
func NewContentServiceWrapper(em ext.ManagerInterface) *ContentServiceWrapper {
	wrapper := &ContentServiceWrapper{em: em}
	wrapper.loadServices()
	return wrapper
}

func (w *ContentServiceWrapper) loadServices() {
	if w == nil || w.em == nil {
		return
	}

	if svc, err := w.em.GetCrossService("content", "Media"); err == nil {
		if mediaService, ok := svc.(ContentMediaServiceInterface); ok {
			w.mediaService = mediaService
		}
	}
	if svc, err := w.em.GetCrossService("content", "TopicMedia"); err == nil {
		if topicMediaService, ok := svc.(ContentTopicMediaServiceInterface); ok {
			w.topicMediaService = topicMediaService
		}
	}
	if svc, err := w.em.GetCrossService("content", "Topic"); err == nil {
		if topicService, ok := svc.(ContentTopicServiceInterface); ok {
			w.topicService = topicService
		}
	}
}

// RefreshServices refreshes content service references.
func (w *ContentServiceWrapper) RefreshServices() {
	w.loadServices()
}

// HasContentServices returns true when all content reference services are available.
func (w *ContentServiceWrapper) HasContentServices() bool {
	return w != nil && w.mediaService != nil && w.topicMediaService != nil
}

// ListMedia lists content media references.
func (w *ContentServiceWrapper) ListMedia(ctx context.Context, params *contentStructs.ListMediaParams) (paging.Result[*contentStructs.ReadMedia], error) {
	return w.mediaService.List(ctx, params)
}

// ListTopicMedia lists topic media references.
func (w *ContentServiceWrapper) ListTopicMedia(ctx context.Context, params *contentStructs.ListTopicMediaParams) (paging.Result[*contentStructs.ReadTopicMedia], error) {
	return w.topicMediaService.List(ctx, params)
}

// GetTopic gets a content topic by ID.
func (w *ContentServiceWrapper) GetTopic(ctx context.Context, id string) (*contentStructs.ReadTopic, error) {
	if w == nil || w.topicService == nil {
		return nil, fmt.Errorf("content topic service not available")
	}
	return w.topicService.GetByID(ctx, id)
}
