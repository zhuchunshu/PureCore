package providers

import (
	controllers "purecore/app/Http/Controllers"
	middleware "purecore/app/Http/Middleware"
	"purecore/core"
)

// RouteServiceProvider registers all application routes.
// This provider implements the core.ServiceProvider interface,
// allowing modular registration of routes without modifying
// the main application bootstrap code.
type RouteServiceProvider struct{}

// Name returns the unique identifier for this provider
func (p *RouteServiceProvider) Name() string {
	return "route"
}

// Register sets up all route groups including public API routes,
// admin routes, and authenticated routes.
func (p *RouteServiceProvider) Register(router *core.Router) error {
	// Register named middleware for reuse across route groups
	router.RegisterNamedMiddlewares(map[string]core.MiddlewareFunc{
		"auth":       middleware.Auth(),
		"admin_auth": middleware.AdminAuth(),
		"cors":       middleware.Cors(),
		"lang":       middleware.Lang(),
	})

	userCtrl := &controllers.UserController{}
	userAuthCtrl := &controllers.UserAuthController{}
	sysCtrl := &controllers.SystemController{}
	optionCtrl := &controllers.OptionController{}
	docsCtrl := &controllers.DocsController{}

	// Public API routes
	router.Prefix("/api/v1").Group(func(r *core.Router) {
		r.Get("/ping", core.H(func(req *core.Request, res *core.Response) error {
			return res.Success("pong")
		}))
		r.Get("/system/info", core.H(sysCtrl.Info))
		// Documentation endpoints
		r.Get("/docs", core.H(docsCtrl.GetDoc))
		r.Get("/docs/list", core.H(docsCtrl.ListDocs))
		// Public user authentication routes
		r.Post("/auth/register", core.H(userAuthCtrl.Register))
		r.Post("/auth/login", core.H(userAuthCtrl.Login))
		r.Post("/auth/refresh", core.H(userAuthCtrl.Refresh))
	})

	// Admin routes (dynamic prefix from config)
	adminCtrl := &controllers.AdminAuthController{}
	adminPrefix := "/api/v1/" + core.GetConfig().AdminRoutePrefix()

	// Public admin routes (no authentication required)
	router.Prefix(adminPrefix).Group(func(r *core.Router) {
		r.Get("/auth/check", core.H(adminCtrl.CheckAdminExists))
		r.Post("/auth/login", core.H(adminCtrl.Login))
		r.Post("/auth/register", core.H(adminCtrl.CreateAdmin))
		r.Post("/auth/refresh", core.H(adminCtrl.Refresh))
		r.Get("/options", core.H(optionCtrl.GetAll))
	})

	// Admin routes requiring authentication
	router.Prefix(adminPrefix).Middleware(middleware.AdminAuth()).Group(func(r *core.Router) {
		r.Get("/auth/profile", core.H(adminCtrl.Profile))
		r.Post("/auth/change-password", core.H(adminCtrl.ChangePassword))
		r.Post("/options", core.H(optionCtrl.Set))
		r.Post("/options/batch", core.H(optionCtrl.BatchSet))
		// User management
		r.Get("/users", core.H(userCtrl.Index))
		r.Post("/users", core.H(userCtrl.Store))
		r.Get("/users/:id", core.H(userCtrl.Show))
		r.Put("/users/:id", core.H(userCtrl.Update))
		r.Delete("/users/:id", core.H(userCtrl.Destroy))
	})

	// Authenticated user routes
	router.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
		r.Get("/auth/profile", core.H(userAuthCtrl.Profile))
		r.Get("/users", core.H(userCtrl.Index))
		r.Post("/users", core.H(userCtrl.Store))
		r.Get("/users/:id", core.H(userCtrl.Show))
	})

	return nil
}

// Boot is called after all providers have been registered.
// Use this for any post-registration initialization if needed.
func (p *RouteServiceProvider) Boot() error {
	return nil
}
