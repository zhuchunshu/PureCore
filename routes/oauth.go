package routes

import (
	controllers "purecore/app/Http/Controllers"
	"purecore/core"
)

// RegisterOAuthRoutes registers public OAuth authentication routes
func RegisterOAuthRoutes(router *core.Router) {
	oauthCtrl := &controllers.OAuthController{}

	router.Prefix("/api/v1/oauth").Group(func(r *core.Router) {
		// Get enabled OAuth providers
		r.Get("/providers", core.H(oauthCtrl.EnabledProviders))
		// Redirect to provider's authorization page
		r.Get("/:provider", core.H(oauthCtrl.Redirect))
		// Handle callback from provider
		r.Get("/:provider/callback", core.H(oauthCtrl.Callback))
	})
}
