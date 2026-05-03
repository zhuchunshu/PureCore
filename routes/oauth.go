package routes

import (
	controllers "purecore/app/Http/Controllers"
	middleware "purecore/app/Http/Middleware"
	"purecore/core"
)

// RegisterOAuthRoutes registers all OAuth-related API routes.
// Public routes: provider list, authorize, callback, register-via-oauth
// Authenticated routes: bind, list accounts, unlink
// Admin routes: get/set provider settings
func RegisterOAuthRoutes(router *core.Router) {
	oauthCtrl := &controllers.OAuthController{}
	adminPrefix := "/api/v1/" + core.GetConfig().AdminRoutePrefix()

	// Public OAuth routes (no auth required)
	router.Prefix("/api/v1/oauth").Group(func(r *core.Router) {
		r.Get("/providers", core.H(oauthCtrl.Providers))
		r.Get("/:provider/authorize", core.H(oauthCtrl.Authorize))
		r.Get("/:provider/callback", core.H(oauthCtrl.Callback))
		r.Post("/register", core.H(oauthCtrl.Register))
	})

	// Authenticated user OAuth routes (require auth)
	router.Prefix("/api/v1/oauth").Middleware(middleware.Auth()).Group(func(r *core.Router) {
		r.Post("/bind", core.H(oauthCtrl.Bind))
		r.Get("/accounts", core.H(oauthCtrl.Accounts))
		r.Delete("/accounts/:id", core.H(oauthCtrl.Unlink))
	})

	// Admin OAuth settings routes (require admin auth)
	router.Prefix(adminPrefix + "/oauth").Middleware(middleware.AdminAuth()).Group(func(r *core.Router) {
		r.Get("/settings", core.H(oauthCtrl.AdminGetSettings))
		r.Post("/settings", core.H(oauthCtrl.AdminSetSettings))
	})
}
