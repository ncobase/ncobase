package service

import (
	"context"
	"errors"
	"fmt"
	"ncobase/proxy/event"
	"ncobase/proxy/structs"

	ext "github.com/ncobase/ncore/extension/types"
)

// ProcessorServiceInterface defines methods for processing data between third-party APIs and internal services
type ProcessorServiceInterface interface {
	// SetEventPublisher sets the event publisher for the processor service
	SetEventPublisher(publisher *event.Publisher)
	// PreProcess processes data before sending to third-party API
	PreProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error)
	// PostProcess processes data after receiving from third-party API
	PostProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error)
	// RegisterHook registers a processing hook for specific endpoint and route
	RegisterHook(endpointID, routeID string, preHook, postHook ProcessorHook) error
	// GetHooks gets hooks for a specific endpoint and route
	GetHooks(endpointID, routeID string) (*ProcessorHookSet, bool)
	// PublishEvent publishes a proxy-related event
	PublishEvent(manager ext.ManagerInterface, eventName string, eventData *event.ProxyEventData)
}

// ProcessorHook defines a function type for data processing hooks
type ProcessorHook func(ctx context.Context, data []byte) ([]byte, error)

// ProcessorHookSet contains hooks for a specific endpoint and route
type ProcessorHookSet struct {
	PreHook  ProcessorHook
	PostHook ProcessorHook
}

// processorService implements ProcessorServiceInterface
type processorService struct {
	hooks     map[string]*ProcessorHookSet // key is "endpointID:routeID"
	publisher *event.Publisher
}

// NewProcessorService creates a new processor service
func NewProcessorService() ProcessorServiceInterface {
	return &processorService{
		hooks: make(map[string]*ProcessorHookSet),
	}
}

// SetEventPublisher sets the event publisher for the processor service
func (s *processorService) SetEventPublisher(publisher *event.Publisher) {
	s.publisher = publisher
}

// PreProcess processes data before sending to third-party API
func (s *processorService) PreProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error) {
	// Get hooks for this endpoint and route
	hookSet, exists := s.GetHooks(endpoint.ID, route.ID)
	if !exists || hookSet.PreHook == nil {
		return data, nil // No pre-processing hook, return original data
	}

	// Execute the pre-processing hook
	return hookSet.PreHook(ctx, data)
}

// PostProcess processes data after receiving from third-party API
func (s *processorService) PostProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error) {
	// Get hooks for this endpoint and route
	hookSet, exists := s.GetHooks(endpoint.ID, route.ID)
	if !exists || hookSet.PostHook == nil {
		return data, nil // No post-processing hook, return original data
	}

	// Execute the post-processing hook
	return hookSet.PostHook(ctx, data)
}

// RegisterHook registers a processing hook for a specific endpoint and route
func (s *processorService) RegisterHook(endpointID, routeID string, preHook, postHook ProcessorHook) error {
	key := fmt.Sprintf("%s:%s", endpointID, routeID)

	// Check if a hook is already registered
	if _, exists := s.hooks[key]; exists {
		return errors.New("hooks already registered for endpoint " + endpointID + " and route " + routeID)
	}

	// Register the hooks
	s.hooks[key] = &ProcessorHookSet{
		PreHook:  preHook,
		PostHook: postHook,
	}

	return nil
}

// GetHooks returns hooks for a specific endpoint and route
func (s *processorService) GetHooks(endpointID, routeID string) (*ProcessorHookSet, bool) {
	key := fmt.Sprintf("%s:%s", endpointID, routeID)
	hookSet, exists := s.hooks[key]
	return hookSet, exists
}

// PublishEvent publishes an event related to proxy operations
func (s *processorService) PublishEvent(manager ext.ManagerInterface, eventName string, eventData *event.ProxyEventData) {
	if s.publisher != nil {
		s.publisher.Publish(eventName, eventData)
	} else if manager != nil {
		// Fallback to direct publishing if no publisher is set
		manager.PublishEvent(eventName, eventData)
	}
}
