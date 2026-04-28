package routes

import (
	providers "purecore/app/Providers"
	"purecore/core"
)

// RegisterAPI is a backward-compatible wrapper that delegates to the
// RouteServiceProvider. New code should use the ServiceProvider directly:
//
//	app.AddProviders(&providers.RouteServiceProvider{})
func RegisterAPI(r *core.Router) {
	provider := &providers.RouteServiceProvider{}
	_ = provider.Register(r)
}
