package oauth

import (
	"fmt"
	"sync"
)

// Registry manages all registered OAuth providers.
// Providers are registered during init() and can be queried at runtime.
// This is the single source of truth for which OAuth providers are available.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider // keyed by provider Name()
}

var globalRegistry = &Registry{
	providers: make(map[string]Provider),
}

// Register adds one or more providers to the global registry.
// Typically called from init() functions in provider files.
// Panics if a provider with the same Name() is already registered.
func Register(providers ...Provider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	for _, p := range providers {
		name := p.Name()
		if _, exists := globalRegistry.providers[name]; exists {
			panic(fmt.Sprintf("oauth: provider %q already registered", name))
		}
		globalRegistry.providers[name] = p
	}
}

// Get returns a provider by name, or nil if not found.
func Get(name string) Provider {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return globalRegistry.providers[name]
}

// All returns all registered providers in no particular order.
func All() []Provider {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	result := make([]Provider, 0, len(globalRegistry.providers))
	for _, p := range globalRegistry.providers {
		result = append(result, p)
	}
	return result
}

// Count returns the number of registered providers.
func Count() int {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	return len(globalRegistry.providers)
}
