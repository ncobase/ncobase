package service

import (
	"fmt"
	"ncobase/plugin/payment/data"
	"ncobase/plugin/payment/data/repository"
	"ncobase/plugin/payment/event"
	"ncobase/plugin/payment/provider"
	"ncobase/plugin/payment/structs"
	"sync"

	"github.com/ncobase/ncore/logging/logger"
)

// ProviderServiceInterface defines the interface for provider service operations
type ProviderServiceInterface interface {
	GetProvider(structs.PaymentProvider, structs.ProviderConfig) (provider.Provider, error)
	GetAllProviders() []structs.PaymentProvider
	ClearCache()
}

// providerService manages payment providers
type providerService struct {
	// Cache providers to avoid recreating them
	providerCache map[structs.PaymentProvider]map[string]provider.Provider
	mu            sync.RWMutex
}

// NewProviderService creates a new provider service
func NewProviderService() ProviderServiceInterface {
	return &providerService{
		providerCache: make(map[structs.PaymentProvider]map[string]provider.Provider),
	}
}

// GetProvider gets or creates a payment provider instance
func (s *providerService) GetProvider(providerName structs.PaymentProvider, config structs.ProviderConfig) (provider.Provider, error) {
	// Generate a cache key based on the config
	cacheKey := s.generateCacheKey(config)

	// Use a double-checked locking pattern
	s.mu.RLock()
	providersByConfig, exists := s.providerCache[providerName]
	if exists {
		if p, ok := providersByConfig[cacheKey]; ok {
			s.mu.RUnlock()
			return p, nil
		}
	}
	s.mu.RUnlock()

	// Create a new provider
	p, err := provider.New(providerName, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment provider %s: %w", providerName, err)
	}

	// Cache the provider with proper locking
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again in case another goroutine created this provider
	// while we were creating ours
	providersByConfig, exists = s.providerCache[providerName]
	if exists {
		if p, ok := providersByConfig[cacheKey]; ok {
			return p, nil
		}
	} else {
		s.providerCache[providerName] = make(map[string]provider.Provider)
	}

	s.providerCache[providerName][cacheKey] = p
	return p, nil
}

// GetAllProviders returns a list of all registered provider names
func (s *providerService) GetAllProviders() []structs.PaymentProvider {
	return provider.GetProviders()
}

// generateCacheKey generates a cache key based on the config
func (s *providerService) generateCacheKey(config structs.ProviderConfig) string {
	// A simple implementation - in production, you might want to use a more robust method
	// like serializing the config to JSON and hashing it

	// For now, just use a combination of keys and values as the cache key
	key := ""

	// Sort keys for deterministic key generation
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}

	// Simple alphabetical sort - not the most efficient but works for this example
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Build the key
	for _, k := range keys {
		key += fmt.Sprintf("%s=%v;", k, config[k])
	}

	return key
}

// ClearCache clears the provider cache
func (s *providerService) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.providerCache = make(map[structs.PaymentProvider]map[string]provider.Provider)

	logger.Info(nil, "Payment provider cache cleared")
}

// Service is the main service provider for the payment module
type Service struct {
	Channel      ChannelServiceInterface
	Order        OrderServiceInterface
	Log          LogServiceInterface
	Product      ProductServiceInterface
	Subscription SubscriptionServiceInterface
	Provider     ProviderServiceInterface
}

// New creates a new payment service
func New(d *data.Data, publisher event.PublisherInterface) *Service {
	// Repositories
	channelRepo := repository.NewChannelRepository(d)
	orderRepo := repository.NewOrderRepository(d)
	logRepo := repository.NewLogRepository(d)
	productRepo := repository.NewProductRepository(d)
	subscriptionRepo := repository.NewSubscriptionRepository(d)

	// Provider service
	providerSvc := NewProviderService()

	// Services
	channelSvc := NewChannelService(channelRepo, publisher)
	logSvc := NewLogService(logRepo)
	productSvc := NewProductService(productRepo, publisher)
	subscriptionSvc := NewSubscriptionService(subscriptionRepo, productRepo, channelRepo, orderRepo, publisher, providerSvc)
	orderSvc := NewOrderService(orderRepo, channelRepo, logRepo, publisher, providerSvc)

	return &Service{
		Channel:      channelSvc,
		Order:        orderSvc,
		Log:          logSvc,
		Product:      productSvc,
		Subscription: subscriptionSvc,
		Provider:     providerSvc,
	}
}
