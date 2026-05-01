package routes

import (
	controllers "purecore/app/Http/Controllers"
	middleware "purecore/app/Http/Middleware"
	"purecore/core"
)

// RegisterSessionRoutes registers all session management routes.
// Session routes require user authentication (Auth middleware).
func RegisterSessionRoutes(router *core.Router) {
	userSessionCtrl := &controllers.UserSessionController{}

	router.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
		r.Get("/sessions", core.H(userSessionCtrl.Index))
		r.Delete("/sessions", core.H(userSessionCtrl.RevokeAll))
		r.Delete("/sessions/:id", core.H(userSessionCtrl.Revoke))
	})
}
