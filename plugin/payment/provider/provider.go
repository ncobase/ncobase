package provider

import (
	"fmt"
	"ncobase/plugin/payment/structs"
	"sync"
)

var (
	// providersMu protects the providers map
	providersMu sync.RWMutex

	// providers maps provider names to factory functions
	providers = make(map[structs.PaymentProvider]ProviderFactory)
)

// Register registers a new payment provider factory
func Register(name structs.PaymentProvider, factory ProviderFactory) {
	providersMu.Lock()
	defer providersMu.Unlock()

	if factory == nil {
		panic("provider: Register factory is nil")
	}

	if _, exists := providers[name]; exists {
		panic(fmt.Sprintf("provider: Register called twice for provider %s", name))
	}

	providers[name] = factory
}

// New creates a new payment provider instance
func New(name structs.PaymentProvider, config structs.ProviderConfig) (Provider, error) {
	providersMu.RLock()
	factoryFn, ok := providers[name]
	providersMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider: unknown provider %s", name)
	}

	return factoryFn(config)
}

// GetProviders returns a list of all registered provider names
func GetProviders() []structs.PaymentProvider {
	providersMu.RLock()
	defer providersMu.RUnlock()

	names := make([]structs.PaymentProvider, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}

	return names
}
