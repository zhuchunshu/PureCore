package core

// ServiceProvider defines the interface that all service providers must implement.
// Providers are the central place to register routes, middleware, and other services.
// This design allows new features to be added by simply creating a new provider
// and registering it, without modifying existing code.
type ServiceProvider interface {
	// Name returns a unique identifier for this provider
	Name() string
	// Register is called when the application is booting up.
	// Use this to register routes, middleware, and other services.
	Register(router *Router) error
	// Boot is called after all providers have been registered.
	// Use this for initialization that depends on other services being available.
	Boot() error
}

// ServiceProviderRegistry manages the collection of service providers
type ServiceProviderRegistry struct {
	providers []ServiceProvider
}

// NewServiceProviderRegistry creates a new registry
func NewServiceProviderRegistry() *ServiceProviderRegistry {
	return &ServiceProviderRegistry{
		providers: make([]ServiceProvider, 0),
	}
}

// Add registers one or more service providers
func (r *ServiceProviderRegistry) Add(providers ...ServiceProvider) {
	r.providers = append(r.providers, providers...)
}

// RegisterAll calls Register on all providers in order
func (r *ServiceProviderRegistry) RegisterAll(router *Router) error {
	for _, p := range r.providers {
		if err := p.Register(router); err != nil {
			return err
		}
	}
	return nil
}

// BootAll calls Boot on all providers in order
func (r *ServiceProviderRegistry) BootAll() error {
	for _, p := range r.providers {
		if err := p.Boot(); err != nil {
			return err
		}
	}
	return nil
}
