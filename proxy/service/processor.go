package service

import (
	"context"
	"encoding/json"
	"fmt"
	spaceStructs "ncobase/core/space/structs"
	"ncobase/proxy/structs"

	accessService "ncobase/core/access/service"
	spaceService "ncobase/core/space/service"
	tenantService "ncobase/core/tenant/service"
	userService "ncobase/core/user/service"

	ext "github.com/ncobase/ncore/extension/types"
	"github.com/ncobase/ncore/logging/logger"
)

// ProcessorServiceInterface defines methods for processing data between third-party APIs and internal services
type ProcessorServiceInterface interface {
	// PreProcess Process data before sending to third-party API
	PreProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error)
	// PostProcess Process data after receiving from third-party API
	PostProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error)
	// RegisterHook Register a processing hook for specific endpoint and route
	RegisterHook(endpointID, routeID string, preHook, postHook ProcessorHook) error
	// GetHooks Get hooks for a specific endpoint and route
	GetHooks(endpointID, routeID string) (*ProcessorHookSet, bool)
	// InitializeEventSupport Initialize event support
	InitializeEventSupport(manager ext.ManagerInterface)
	// PublishEvent Publish a proxy-related event
	PublishEvent(manager ext.ManagerInterface, eventName string, eventData *ProxyEventData)
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
	hooks map[string]*ProcessorHookSet // key is "endpointID:routeID"

	// Internal services
	us *userService.Service
	ts *tenantService.Service
	ss *spaceService.Service
	as *accessService.Service
}

// NewProcessorService creates a new processor service
func NewProcessorService(
	us *userService.Service,
	ts *tenantService.Service,
	ss *spaceService.Service,
	as *accessService.Service,
) ProcessorServiceInterface {
	return &processorService{
		hooks: make(map[string]*ProcessorHookSet),
		us:    us,
		ts:    ts,
		ss:    ss,
		as:    as,
	}
}

// PreProcess processes data before sending to third-party API
func (s *processorService) PreProcess(ctx context.Context, endpoint *structs.ReadEndpoint, route *structs.ReadRoute, data []byte) ([]byte, error) {
	// Get hooks for this endpoint and route
	hookSet, exists := s.GetHooks(endpoint.ID, route.ID)
	if !exists || hookSet.PreHook == nil {
		return data, nil // No pre-processing hook, return original data
	}

	// Execute the pre-processing hook
	logger.Infof(ctx, "Executing pre-processing hook for endpoint %s, route %s", endpoint.ID, route.ID)
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
	logger.Infof(ctx, "Executing post-processing hook for endpoint %s, route %s", endpoint.ID, route.ID)
	return hookSet.PostHook(ctx, data)
}

// RegisterHook registers a processing hook for a specific endpoint and route
func (s *processorService) RegisterHook(endpointID, routeID string, preHook, postHook ProcessorHook) error {
	key := fmt.Sprintf("%s:%s", endpointID, routeID)

	// Check if a hook is already registered
	if _, exists := s.hooks[key]; exists {
		return fmt.Errorf("hooks already registered for endpoint %s and route %s", endpointID, routeID)
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

// UserDataEnrichmentHook Example hook implementation for user data enrichment
func UserDataEnrichmentHook(userSvc *userService.Service) ProcessorHook {
	return func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input data: %w", err)
		}

		// Check if the data contains user IDs
		if userID, ok := inputData["user_id"].(string); ok {
			// Get user details from the user service
			user, err := userSvc.User.GetByID(ctx, userID)
			if err != nil {
				logger.Warnf(ctx, "Failed to get user by ID %s: %v", userID, err)
			} else {
				// Enrich data with user information
				inputData["user"] = user
			}
		}

		// Convert back to JSON
		enrichedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal enriched data: %w", err)
		}

		return enrichedData, nil
	}
}

// TenantDataProcessingHook Example hook implementation for tenant data processing
func TenantDataProcessingHook(tenantSvc *tenantService.Service) ProcessorHook {
	return func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input data: %w", err)
		}

		// Process tenant-specific data
		if tenantID, ok := inputData["tenant_id"].(string); ok {
			// Get tenant details from the tenant service
			tenant, err := tenantSvc.Tenant.GetBySlug(ctx, tenantID)
			if err != nil {
				logger.Warnf(ctx, "Failed to get tenant by ID %s: %v", tenantID, err)
			} else {
				// Process data with tenant information
				inputData["tenant"] = tenant
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processed data: %w", err)
		}

		return processedData, nil
	}
}

// SpaceDataProcessingHook Example hook implementation for space data processing
func SpaceDataProcessingHook(spaceSvc *spaceService.Service) ProcessorHook {
	return func(ctx context.Context, data []byte) ([]byte, error) {
		// Parse input data
		var inputData map[string]any
		if err := json.Unmarshal(data, &inputData); err != nil {
			return nil, fmt.Errorf("failed to parse input data: %w", err)
		}

		// Process space-specific data
		if groupID, ok := inputData["group_id"].(string); ok {
			// Get group details from the space service
			group, err := spaceSvc.Group.Get(ctx, &spaceStructs.FindGroup{Group: groupID})
			if err != nil {
				logger.Warnf(ctx, "Failed to get group by ID %s: %v", groupID, err)
			} else {
				// Process data with group information
				inputData["group"] = group
			}
		}

		// Convert back to JSON
		processedData, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal processed data: %w", err)
		}

		return processedData, nil
	}
}
